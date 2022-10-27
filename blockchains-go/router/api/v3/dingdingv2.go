package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	dingModel "github.com/group-coldwallet/blockchains-go/model/dingding"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"github.com/shopspring/decimal"
	"strings"
	"time"
	"xorm.io/builder"
)

func setSameWayBack(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_SAME_WAY_BACK.ToString(), "", -1)
	params := new(transfer.SameWayBack)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	if params.TxId == "" {
		return fmt.Errorf("txId不能为空")
	}

	txs, err := dao.FcTxTransactionsByTxId(params.TxId)
	if err != nil {
		return fmt.Errorf("dao.FcTxTransactionsByTxId 出错:%v", err)
	}
	if len(txs) == 0 {
		return fmt.Errorf("FcTxTransactionNew没有任何txId=%s 的交易数据", params.TxId)
	}

	var sb strings.Builder
	for _, tx := range txs {
		key := redis.GetWatchSameWayBackCacheKeyWithCoinType(tx.CoinType, tx.ToAddress, tx.FromAddress)
		if err := redis.Client2.Set(key, tx.OuterOrderNo, time.Hour*24*6000); err != nil {
			return fmt.Errorf("设置到redis失败:%v\ntxId:%s\n币种:%s\nfrom地址:%s\nto地址:%s", err, params.TxId, tx.CoinType, tx.FromAddress, tx.ToAddress)
		}
		sb.WriteString(key)
		sb.WriteString("\n")
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("回退交易设置成功，keys:\n%s", sb.String()))
	return nil
}

func orderCollectProcess(outerOrderNo string) error {
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(outerOrderNo))
	}()
	apply, err := dao.FcTransfersApplyByOutOrderNo(outerOrderNo)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%v]", outerOrderNo, err)
	}
	if apply.Status == int(entity.ApplyStatus_TransferOk) {
		return fmt.Errorf("订单：%s 已完成，不可再进行归集", outerOrderNo)
	}

	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id, "address_flag": "to"})
	if err != nil {
		return err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", apply.Id, apply.OutOrderid)
	}
	toEntity := toAddrs[0]
	toAmt, _ := decimal.NewFromString(toEntity.ToAmount)

	coinType := apply.CoinName
	if apply.Eoskey != "" {
		coinType = apply.Eoskey
	}

	coinSet, err := dao.FcCoinSetGetByName(coinType, 1)
	if err != nil {
		return fmt.Errorf("dao.FcCoinSetGetByName err:%v", err)
	}
	msg, err := order.ManualCollect(apply.AppId, outerOrderNo, coinType, coinSet.Token, coinSet.CollectThreshold, toAmt)
	if err != nil {
		return fmt.Errorf("订单归集（%s）出错: %v", outerOrderNo, err)
	}
	dingding.ReviewDingBot.NotifyStr(msg)
	return nil
}

func cancelOrderForMultiAddr(apply *entity.FcTransfersApply) error {
	req := struct {
		OuterOrderNo string `json:"outerOrderNo"`
	}{OuterOrderNo: apply.OutOrderid}
	callUrl := fmt.Sprintf("%s/v2/cancelorder", conf.Cfg.Walletserver.Url)
	return callWalletServer(callUrl, req)
}

func forceCancelTx(content string) error {
	seqNo := strings.Replace(content, dingModel.DING_FORCE_CANCEL_TXS.ToString(), "", -1)
	seqNo = strings.TrimSpace(seqNo)
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(seqNo))
	}()
	if err := rePushTryLock(seqNo); err != nil {
		return err
	}
	callUrl := fmt.Sprintf("%s/v2/forcecanceltx", conf.Cfg.Walletserver.Url)
	cancelTxForMultiFromCore(callUrl, seqNo)
	return nil
}

func cancelTx(content string) error {
	seqNo := strings.Replace(content, dingModel.DING_CANCEL_TXS.ToString(), "", -1)
	seqNo = strings.TrimSpace(seqNo)
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(seqNo))
	}()
	if err := rePushTryLock(seqNo); err != nil {
		return err
	}
	callUrl := fmt.Sprintf("%s/v2/canceltx", conf.Cfg.Walletserver.Url)
	cancelTxForMultiFromCore(callUrl, seqNo)
	return nil
}

func replaceFailureTxs(content string) error {
	outOrderId := strings.Replace(content, dingModel.DING_REPLACE_FAILURE_TXS.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	log.Infof("替换失败交易 %s", outOrderId)
	err := tm.ReplaceTxs(outOrderId)
	if err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单:%s 替换失败交易出错: %v", outOrderId, err))
	} else {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单:%s 替换失败交易请求成功", outOrderId))
	}
	return nil
}

func DiscardAndRePush(content string, auth *dingModel.DingRoleAuth) error {
	// 后缀跟的是订单号
	outOrderId := strings.Replace(content, dingModel.DING_DISCARD_REPUSH_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)

	if err := rePushTryLock(outOrderId); err != nil {
		return err
	}
	return discardAndRePushProcess(outOrderId, auth)
}

func discardAndRePushProcess(outOrderId string, auth *dingModel.DingRoleAuth) error {
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(outOrderId))
	}()
	tx, err := dao.GetOrderTxBySeqNo(outOrderId)
	if err != nil {
		return err
	}
	log.Infof("获取到的txs %v", tx)
	if tx != nil {
		// 多地址出账t
		rePushForceForMultiFrom(tx)
		return nil
	}

	applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
	log.Infof("获取applyOrder信息完成 %s", outOrderId)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}
	log.Infof("HaveCoin完成 %s", outOrderId)

	txBreakList, err := dao.FindTxBreakByChain(applyOrder.CoinName)
	if err != nil {
		return fmt.Errorf("订单：%s,FindTxBreakByChain,error=[%s]", outOrderId, err.Error())
	}

	if len(txBreakList) > 0 {
		return fmt.Errorf("主链：%s 存在未处理的中断交易，需要进行关联\n订单：%s\ntxId：%s", applyOrder.CoinName, txBreakList[0].OutOrderNo, txBreakList[0].TxId)
	}

	rollbackFunc, err := checkIfCanDiscardAndUpdateOrderStatus(outOrderId, status.AbandonedTransaction, applyOrder)
	if err != nil {
		return err
	}
	log.Infof("checkIfCanDiscardAndUpdateOrderStatus完成 %s", outOrderId)

	if rollbackFunc == nil {
		return errors.New(fmt.Sprintf("%s 订单无相关执行记录", outOrderId))
	}

	if err = rePush(outOrderId, applyOrder); err != nil {
		log.Infof("重推订单[%s]失败,回滚 checkIfCanDiscardAndUpdateOrderStatus 所修改的状态", outOrderId)
		rollbackFunc()
		return err
	}
	return nil
}

// DiscardAndRollback 废弃回滚
func DiscardAndRollback(content string, auth *dingModel.DingRoleAuth) error {
	// 后缀跟的是订单号
	outOrderId := strings.Replace(content, dingModel.DING_DISCARD_ROLLBACK_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	if err := rePushTryLock(outOrderId); err != nil {
		return err
	}
	return discardAndRollbackProcess(outOrderId, auth)
}

func discardAndRollbackProcess(outOrderId string, auth *dingModel.DingRoleAuth) error {
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(outOrderId))
	}()
	applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}

	txBreakList, err := dao.FindTxBreakByChain(applyOrder.CoinName)
	if err != nil {
		return fmt.Errorf("订单：%s,FindTxBreakByChain,error=[%s]", outOrderId, err.Error())
	}

	if len(txBreakList) > 0 {
		return fmt.Errorf("主链：%s 存在未处理的中断交易，需要进行关联\n订单：%s\ntxId：%s", applyOrder.CoinName, txBreakList[0].OutOrderNo, txBreakList[0].TxId)
	}

	if applyOrder.TxType == entity.MultiAddrTx {
		log.Infof("废弃回滚 订单%s 为多地址出账类型订单", outOrderId)
		if err = cancelOrderForMultiAddr(applyOrder); err != nil {
			return err
		}
		_, err = dao.FcTransfersApplyUpdateStatusForRollbackForMultiAddr(applyOrder.Id, int(entity.ApplyStatus_Rollback))
		if err != nil {
			log.Infof("更新transferApply[%s]状态失败,回滚 checkIfCanDiscardAndUpdateOrderStatus 所修改的状态", outOrderId)
			return err
		}

	} else {
		rollbackFunc, err := checkIfCanDiscardAndUpdateOrderStatus(outOrderId, status.RollbackTransaction, applyOrder)
		if err != nil {
			return err
		}

		rowsAffected, err := dao.FcTransfersApplyUpdateStatusForRollback(applyOrder.Id, int(entity.ApplyStatus_Rollback))
		if err != nil {
			log.Infof("更新transferApply[%s]状态失败,回滚 checkIfCanDiscardAndUpdateOrderStatus 所修改的状态", outOrderId)
			if rollbackFunc != nil {
				rollbackFunc()
			}
			return err
		}
		if rowsAffected == 0 {
			log.Infof("transferApply[%s]状态 受影响行数为0,回滚 checkIfCanDiscardAndUpdateOrderStatus 所修改的状态", outOrderId)
			if rollbackFunc != nil {
				rollbackFunc()
			}
			return errors.New("操作失败：更新transferApply状态 受影响行数为0")
		}
	}
	dao.UpdatePriorityCompletedIfExist(applyOrder.Id)
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单 %s 已设置为回滚状态", outOrderId))

	return nil
}

func rePush(outOrderId string, applyOrder *entity.FcTransfersApply) error {
	// 查询是否允许重推
	err := api.OrderService.IsAllowRepush(outOrderId)
	log.Infof("IsAllowRepush完成 %s", outOrderId)

	if err != nil {
		return err
	}
	// 特殊处理链上回滚订单

	// 设置重新出账,重试一次，设置错误次数为3即可
	err = api.OrderService.SendApplyRetryOnce(applyOrder.Id, global.RetryNum-1)
	log.Infof("SendApplyRetryOnce完成 %s", outOrderId)

	if err != nil {
		return fmt.Errorf("重推订单：%s，异常:%s", outOrderId, err.Error())
	}

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，异常:%s", outOrderId, err.Error()))
		return err
	}
	defer redisHelper.Close()
	interceptKey := fmt.Sprintf("%d_%s", applyOrder.AppId, applyOrder.OutOrderid)
	// 清除rediskey
	_ = redisHelper.Del(interceptKey)

	cltMsg := ""
	if order.NewCollectEnable() && order.IsNewCollectVersion(applyOrder.CoinName) {
		if err = tryCollect(applyOrder); err != nil {
			cltMsg = "，出账归集失败:" + err.Error()
		}
	}

	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，成功%s", outOrderId, cltMsg))
	log.Infof("NotifyStr完成 %s", outOrderId)

	return nil
}

func checkIfCanDiscardAndUpdateOrderStatus(outOrderId string, newStatus status.OrderStatus, applyOrder *entity.FcTransfersApply) (func(), error) {
	if applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		return nil, fmt.Errorf("订单：%s，apply 状态不允许废弃", outOrderId)
	}

	// 检查币种类型
	walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)
	switch walletType {
	case status.WalletType_Cold:
		orders, _ := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
		if len(orders) == 0 {
			return nil, nil
		}

		for _, v := range orders {
			if v.Status == status.BroadcastStatus.Int() {
				return nil, fmt.Errorf("订单：%s，已经广播，不允许废弃", outOrderId)
			}
		}

		var orderId string
		var oriStatus int
		for _, v := range orders {
			if v.Status == status.BroadcastErrorStatus.Int() || v.Status == status.SignErrorStatus.Int() {
				orderId = v.OrderNo
				oriStatus = v.Status
				break
			}
		}
		if orderId == "" {
			return nil, fmt.Errorf("订单：%s，不存在异常，不允许废弃", outOrderId)
		}

		rollbackFunc := func() {
			api.WalletOrderService.UpdateColdOrderState(orderId, oriStatus)
		}
		log.Infof("准备更新 UpdateColdOrderState %s", outOrderId)
		return rollbackFunc, api.WalletOrderService.UpdateColdOrderState(orderId, newStatus.Int())
	case status.WalletType_Hot:
		orders, _ := api.WalletOrderService.GetHotOrder(applyOrder.OutOrderid)
		if len(orders) == 0 {
			return nil, nil
		}

		for _, v := range orders {
			if v.Status == status.BroadcastStatus.Int() {
				return nil, fmt.Errorf("订单：%s，已经广播，不允许废弃", outOrderId)
			}
		}

		var orderId string
		var oriStatus int
		for _, v := range orders {
			if v.Status == status.BroadcastErrorStatus.Int() || v.Status == status.SignErrorStatus.Int() {
				orderId = v.OrderNo
				oriStatus = v.Status
				break
			}
		}
		if orderId == "" {
			return nil, fmt.Errorf("订单：%s，不存在异常，不允许废弃", outOrderId)
		}
		rollbackFunc := func() {
			api.WalletOrderService.UpdateHotOrderState(orderId, oriStatus)
		}
		log.Infof("准备更新 UpdateHotOrderState %s", outOrderId)
		return rollbackFunc, api.WalletOrderService.UpdateHotOrderState(orderId, newStatus.Int())
	default:
		return nil, errors.New("error walletType")
	}
}

func FixCollectAddressAsset(content string) error {
	chainName := "eth"
	limitCount := 10
	limit := 20

	req := &model.FixAddressReq{}
	jsonDataStr := strings.Replace(content, dingModel.DING_FIX_ADDR_AMOUNT_ALL0C.ToString(), "", -1)
	if err := json.Unmarshal([]byte(jsonDataStr), req); err != nil {
		return err
	}

	eachFee := req.FeeAmount
	days := req.Days

	currentTime := time.Now()
	oldTime := currentTime.AddDate(0, 0, -days).Unix()

	feeCount, err := dao.FcOrderTotalFeeCount(oldTime, chainName)
	if err != nil {
		return err
	}
	if feeCount.Count == 0 {
		return errors.New(fmt.Sprintf("[固定地址金额] 获取到 %d 天大手续费总笔数为0", days))
	}

	addressFeeCountList, err := dao.FcOrderAddressFeeCount(oldTime, chainName, limitCount, limit)
	if err != nil {
		return err
	}
	l1 := fmt.Sprintf("[固定地址金额] 获取前%d条打手续费数量大于%d的记录为 %d", limit, limitCount, len(addressFeeCountList))
	log.Info(l1)
	if len(addressFeeCountList) == 0 {
		return errors.New(l1)
	}
	var addrs []string
	for _, fc := range addressFeeCountList {
		addrs = append(addrs, fc.Address)
	}

	receiveList, err := dao.FcOrderFindReceive(oldTime, chainName, addrs)
	if err != nil {
		return err
	}
	newAddressCountList := make([]*dao.FcAddressFeeCount, 0)
	if len(receiveList) > 0 {
		for _, fc := range addressFeeCountList {
			pass := true
			for _, r := range receiveList {
				amt, _ := decimal.NewFromString(r.Amount)
				// 指定条件下主链币充值金额大于1，禁止固定地址金额
				if amt.Cmp(decimal.NewFromInt(1)) == 1 {
					log.Infof("[固定地址金额] 地址%s 指定条件下主链币充值金额为 %s，禁止固定地址金额", r.Address, amt.String())
					pass = false
					break
				}
			}
			if pass {
				newAddressCountList = append(newAddressCountList, fc)
			}
		}
	} else {
		newAddressCountList = addressFeeCountList
		log.Info("[固定地址金额] 接收所有地址")
	}

	if err = dao.UpdateFcFixAddressDiscard(); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("固定地址金额：\n")
	models := make([]entity.FcFixAddress, 0)
	totalFee := decimal.Zero
	for _, item := range newAddressCountList {
		count := decimal.NewFromInt(item.Count)
		fee := count.Mul(eachFee)
		models = append(models, entity.FcFixAddress{
			ChainName:  chainName,
			Address:    item.Address,
			Amount:     fee.String(),
			Status:     1,
			Payed:      0,
			CreateTime: currentTime,
		})
		totalFee = totalFee.Add(fee)
		sb.WriteString(fmt.Sprintf("%s %s \n", item.Address, fee.String()))
	}
	sb.WriteString("本次总固定金额为 " + totalFee.String() + "\n")
	if req.TotalAmount.Div(decimal.NewFromInt(2)).Cmp(totalFee) == -1 {
		// 如果本次需要分配的固定金额比总金额的一半还多，终止本次操作
		sb.WriteString("本次需要分配的固定金额比总金额的一半还多，终止分配操作")
		dingding.ReviewDingBot.NotifyStr(sb.String())
		return nil
	}
	dao.InsertFcFixAddressBatch(models)
	dingding.ReviewDingBot.NotifyStr(sb.String())
	return nil
}

func OrderTxLink(content string) error {
	req := &model.OrderTxLinkReq{}
	jsonDataStr := strings.Replace(content, dingModel.DING_ORDER_TX_LINK.ToString(), "", -1)
	if err := json.Unmarshal([]byte(jsonDataStr), req); err != nil {
		return err
	}

	if req.TxId == "" || req.OrderNo == "" {
		return errors.New("txId和orderNo不能为空")
	}

	applyOrder, err := api.OrderService.GetApplyOrder(req.OrderNo)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", req.OrderNo, err.Error())
	}

	if applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		return fmt.Errorf("订单：%s，当前状态：%d 状态不合法（必须为49）", req.OrderNo, applyOrder.Status)
	}

	// 检查币种类型
	walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)

	switch walletType {
	case status.WalletType_Cold:
		orders, _ := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
		if len(orders) == 0 {
			return fmt.Errorf("在fc_order表没有找到该订单 %s", req.OrderNo)
		}

		for _, v := range orders {
			if v.Status == status.BroadcastStatus.Int() {
				return fmt.Errorf("订单：%s，已经广播，不允许废弃", req.OrderNo)
			}
		}

		orderId := 0
		for _, v := range orders {
			if v.Status == status.BroadcastErrorStatus.Int() {
				orderId = v.Id
				break
			}
		}
		if orderId == 0 {
			return fmt.Errorf("订单：%s，不存在异常(没有发现状态7的订单)，不能执行该操作", req.OrderNo)
		}

		err = dao.FcOrderUpdateTxIdAndStatus(orderId, req.TxId)
	case status.WalletType_Hot:
		orders, _ := api.WalletOrderService.GetHotOrder(applyOrder.OutOrderid)
		if len(orders) == 0 {
			return fmt.Errorf("在fc_order_hot表没有找到该订单 %s", req.OrderNo)
		}

		for _, v := range orders {
			if v.Status == status.BroadcastStatus.Int() {
				return fmt.Errorf("订单：%s，已经广播，不允许废弃", req.OrderNo)
			}
		}

		orderId := 0
		for _, v := range orders {
			if v.Status == status.BroadcastErrorStatus.Int() {
				orderId = v.Id
				break
			}
		}
		if orderId == 0 {
			return fmt.Errorf("订单：%s，不存在异常(没有发现状态7的订单)，不能执行该操作", req.OrderNo)
		}
		err = dao.FcOrderHotUpdateTxIdAndStatus(orderId, req.TxId)
	default:
		return errors.New("error walletType")
	}

	if err != nil {
		return fmt.Errorf("%s 更新订单状态和txid失败:%v", req.OrderNo, err)
	}
	log.Infof("%s 更新订单表状态和txId完成", req.OrderNo)

	err = dao.FcTransfersApplyUpdateStatusById(applyOrder.Id, int(entity.ApplyStatus_TransferOk))
	if err != nil {
		return fmt.Errorf("%s 更新transfersApply状态失败:%v", req.OrderNo, err)
	}
	log.Infof("%s 更新transfersApply状态完成", req.OrderNo)

	err = dao.DeleteTxBreakById(req.TxId)
	if err != nil {
		return fmt.Errorf("%s 删除tx_break数据失败:%v", req.OrderNo, err)
	}

	err = api.OrderService.NotifyToMchByOutOrderId(req.OrderNo)
	if err != nil {
		return fmt.Errorf("%s 重新推送txid到交易所失败:%v", req.OrderNo, err)
	}

	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单:%s 关联成功，请对 %s 执行一次补数据", req.OrderNo, req.TxId))
	return nil
}
