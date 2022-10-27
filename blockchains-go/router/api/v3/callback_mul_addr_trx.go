package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
	"time"
)

type CallbackMulAddrTrxResp struct {
	SeqNo string `json:"seqNo"`
}

func CallbackMulAddrTrx(ctx *gin.Context) {
	bodyByte, _ := ioutil.ReadAll(ctx.Request.Body)
	log.Infof("接收来自walletServer的订单交易回调的内容是：%s", string(bodyByte))
	info := &CallbackMulAddrTrxResp{}
	if err := json.Unmarshal(bodyByte, info); err != nil {
		httpresp.HttpRespErrWithMsg(ctx, fmt.Sprintf("json.Unmarshal err %v", err))
		return
	}

	if err := processMulAddrTrx(info); err != nil {
		log.Errorf("处理来自walletServer的回调订单交易失败 %v", err)
		httpresp.HttpRespErrWithMsg(ctx, err.Error())
		return
	}
	httpresp.HttpRespCodeOkOnly(ctx)
}

func processMulAddrTrx(info *CallbackMulAddrTrxResp) error {
	errMsg := ""
	seqNo := info.SeqNo
	if seqNo == "" {
		return errors.New("seqNo 不能为空")
	}

	orderTx, err := dao.GetOrderTxBySeqNo(seqNo)
	if err != nil {
		return fmt.Errorf("根据seqNo(%s)获取数据失败 %v", seqNo, err)
	}
	if orderTx == nil {
		return fmt.Errorf("根据seqNo(%s)获取数据为空", seqNo)
	}

	//处理订单
	applyOrder, err := api.OrderService.GetApplyOrder(orderTx.OuterOrderNo)
	if err != nil {
		return fmt.Errorf("没有这个订单：%s", orderTx.OuterOrderNo)
	}

	// 多地址出账，apply的状态为47
	// 但是如果交易中途处理失败，会变成49，所以合理兼容49的情况
	if applyOrder.Status != int(entity.ApplyStatus_CreateOk) && applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		if applyOrder.Status == int(entity.ApplyStatus_Merge) {
			errMsg = fmt.Sprintf("归集订单忽略 :%s", orderTx.OuterOrderNo)
		} else if applyOrder.Status == int(entity.ApplyStatus_Fee) {
			errMsg = fmt.Sprintf("打手续费订单忽略 :%s", orderTx.OuterOrderNo)
		} else {
			errMsg = fmt.Sprintf("订单状态异常：%s,状态%d", orderTx.OuterOrderNo, applyOrder.Status)
		}
		return errors.New(errMsg)
	}

	mch, err := api.MchService.GetAppId(orderTx.Mch)
	if err != nil {
		return fmt.Errorf("没有这个商户：%s", orderTx.Mch)
	}
	if applyOrder.AppId != mch.Id {
		return fmt.Errorf("商户校验异常：%s", orderTx.Mch)
	}

	orders, err := dao.FcOrderFindByOutNo(orderTx.OuterOrderNo)
	if err != nil {
		return fmt.Errorf("根据outOrderNo(%s)获取数据失败 %v", orderTx.OuterOrderNo, err)
	}

	if len(orders) != 1 { // 多地址出账，在fc_order表同一outerOrderNo只能有一条记录
		return fmt.Errorf("outerOrderNo(%s) 在订单表记录条数(%d)不为一", orderTx.OuterOrderNo, len(orders))
	}

	order := orders[0]

	if err = checkIfOrderAndTxsAmountNotEqual(order); err != nil {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("walletServer回调blockchainsgo：%s", err.Error()))
		return fmt.Errorf("检查订单金额是否与交易总金额相等失败 %v", err)
	}

	log.Infof("回调的订单交易状态为：%d", orderTx.Status)
	if orderTx.Status == entity.OtxBroadcastSuccess {
		if err = processMulAddrTrxForSuccess(applyOrder, order, orderTx); err != nil {
			msg := fmt.Sprintf("walletServer回调处理订单交易(%s-%s)失败:%v", order.OuterOrderNo, orderTx.SeqNo, err)
			log.Infof(msg)
			dingding.ErrTransferDingBot.NotifyStr(msg)
		}

		if err = watchSameWayBack(order.OuterOrderNo, orderTx.SeqNo, orderTx.Chain, orderTx.CoinCode, orderTx.FromAddress, orderTx.ToAddress); err != nil {
			msg := fmt.Sprintf("walletServer回调处理订单交易(%s-%s)，设置防原路退回失败:%v", order.OuterOrderNo, orderTx.SeqNo, err)
			log.Infof(msg)
			dingding.ErrTransferDingBot.NotifyStr(msg)
		}

		return nil
	}

	err = api.OrderService.SendApplyFail(applyOrder.Id)
	if err != nil {
		msg := fmt.Sprintf("walletServer回调处理订单(%s)修改transfers_apply状态出错: %v", order.OuterOrderNo, err)
		dingding.ErrTransferDingBot.NotifyStr(msg)
	}

	// 订单交易处理失败，钉钉通知值班人员
	txs, _ := dao.FindOrderTxsByOuterOrderNo(order.OuterOrderNo)
	successCount := 0
	failureCount := 0
	processingCount := 0
	for _, t := range txs {
		if t.IsChainFailure() {
			continue
		}
		if t.IsCompleted() {
			successCount += 1
		} else if t.IsProcessing() {
			processingCount += 1
		} else {
			failureCount += 1
		}
	}

	var sb strings.Builder
	sb.WriteString("walletServer处理订单交易失败:\n")
	sb.WriteString(fmt.Sprintf("订单：%s 状态：%d\n", order.OuterOrderNo, order.Status))
	sb.WriteString(fmt.Sprintf("交易：%s 状态：%d\n", orderTx.SeqNo, orderTx.Status))
	sb.WriteString(fmt.Sprintf("错误信息：%s\n", orderTx.ErrMsg))
	sb.WriteString("该订单下交易：\n")
	sb.WriteString(fmt.Sprintf("成功广播：%d笔\n", successCount))
	sb.WriteString(fmt.Sprintf("正在处理：%d笔\n", processingCount))
	sb.WriteString(fmt.Sprintf("处理失败：%d笔\n", failureCount))
	dingding.ErrTransferDingBot.NotifyStr(sb.String())

	return nil
}

func checkIfOrderAndTxsAmountNotEqual(order *entity.FcOrder) error {
	txs, err := dao.FindOrderTxsByOuterOrderNo(order.OuterOrderNo)
	if err != nil {
		return err
	}

	// 订单的金额
	orderAmount, _ := decimal.NewFromString(order.TotalAmount)
	// 订单下交易的总金额（排除链上失败交易）
	txsTotalAmount := decimal.Zero
	// 已广播成功的交易总金额
	txsSuccessTotalAmount := decimal.Zero
	for _, tx := range txs {
		if tx.IsChainFailure() {
			// 排除掉链上失败交易
			continue
		}
		if tx.IsCanceled() {
			// 排除已取消的交易
			continue
		}

		fs, _ := decimal.NewFromString(tx.Amount)
		if tx.IsCompleted() {
			// 统计已广播成功的交易金额
			txsSuccessTotalAmount = txsSuccessTotalAmount.Add(fs)
		}

		// 统计有效的交易金额
		txsTotalAmount = txsTotalAmount.Add(fs)
	}
	if order.Status == int(status.BroadcastStatus) {
		if !orderAmount.Equal(txsSuccessTotalAmount) {
			return fmt.Errorf("订单(%s)状态为已广播成功，但订单金额(%s)与已广播成功的交易总金额(%s)不相等", order.OuterOrderNo, orderAmount.String(), txsSuccessTotalAmount.String())
		}
	}

	if !orderAmount.Equal(txsTotalAmount) {
		return fmt.Errorf("订单金额(%s)与交易总金额(%s)不相等，当前订单状态为:%s", orderAmount.String(), txsTotalAmount.String(), status.StatusDesc[status.OrderStatus(order.Status)])
	}
	return nil
}

func processMulAddrTrxForSuccess(apply *entity.FcTransfersApply, order *entity.FcOrder, orderTx *entity.FcOrderTxs) error {
	// 如果订单已经完成，扭转transfers_apply的状态为已完成
	// 前面已经校验了已完成订单与它底下的交易总金额是否匹配
	if order.Status == int(status.BroadcastStatus) {
		err := api.OrderService.SendApplyTransferSuccess(apply.Id)
		if err != nil {
			errMsg := fmt.Sprintf("接收到walletServer回调信息，订单交易%s-%s 广播成功，但修改applyTransfers状态失败", order.OuterOrderNo, orderTx.SeqNo)
			dingding.ErrTransferDingBot.NotifyStr(errMsg)
			return errors.New(errMsg)
		}
		dao.UpdatePriorityCompletedIfExist(apply.Id)
	}

	//通知商户
	return api.OrderService.NotifyToMchV2(apply, order, orderTx)
}

func watchSameWayBack(outerOrder, seqNo, chain, coinCode, from, to string) error {
	key := redis.GetWatchSameWayBackCacheKey(chain, coinCode, from, to)
	log.Infof("订单:%s 交易流水号:%s 生成的防原路返回缓存key:%s", outerOrder, seqNo, key)
	return redis.Client2.Set(key, outerOrder, time.Hour*24*32) // 保存32天
}
