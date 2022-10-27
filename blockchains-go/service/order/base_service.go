package order

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/runtime/wechat"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type OrderBaseService struct {
}

func (srv *OrderBaseService) SendApplyRollback(applyId int) error {
	result, err := dao.FcTransfersApplyById(applyId)
	if err != nil {
		log.Errorf("SendApplyRollback error:%s", err.Error())
		return fmt.Errorf("回滚交易状态异常：%s", err.Error())
	}
	if result.Status != int(entity.ApplyStatus_Auditing) {
		return fmt.Errorf("订单：%s,可回滚状态40，当前订单状态：%s", result.OutOrderid, result.Status)
	}
	walletType := global.WalletType(result.CoinName, result.AppId)
	err = dao.FcTransfersApplyAbandoned(result.OutOrderid, result.OrderId, walletType)

	//return dao.FcTransfersApplyUpdateStatusById(applyId, int(entity.ApplyStatus_Ignore))
	return err
}

func (srv *OrderBaseService) AbandonedOrder(outOrderId string) error {
	apply, err := dao.FcTransfersApplyByOutOrderNo(outOrderId)
	if err != nil {
		return fmt.Errorf("查询异常:%s", err.Error())
	}
	if apply.Status != int(entity.ApplyStatus_TransferOk) {
		return fmt.Errorf("订单非完成状态，不允许废弃，当前状态%d", apply.Status)
	}
	walletType := global.WalletType(apply.CoinName, apply.AppId)
	err = dao.FcTransfersApplyAbandoned(apply.OutOrderid, apply.OrderId, walletType)
	return err

}

//重试一次
func (srv *OrderBaseService) SendApplyRetryOnce(applyId int, errNum int64) error {
	data, err := dao.FcTransfersApplyById(applyId)
	if err != nil {
		return err
	}

	log.Infof("[SendApplyRetryOnce] 准备执行FcTransfersApplyUpdateErrNum ")
	err = dao.FcTransfersApplyUpdateErrNum(int64(applyId), errNum)
	if err != nil {
		return err
	}
	log.Infof("[SendApplyRetryOnce] 执行 FcTransfersApplyUpdateErrNum 完成")
	//查询交易地址
	addrInfos, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(applyId, "to")
	if err != nil {
		return err
	}
	log.Infof("[SendApplyRetryOnce] 执行 FcTransfersApplyCoinAddressFindAddrInfo 完成")
	if len(addrInfos) == 0 {
		return errors.New("empty address info ")
	}

	orderNo, err := util.GetOrderId(data.OutOrderid, addrInfos[0].Address, addrInfos[0].ToAmount)
	if err != nil {
		return err
	}
	log.Infof("[SendApplyRetryOnce] 执行 GetOrderId 完成")

	//return dao.FcTransfersApplyUpdateStatusAddErr(applyId, int(entity.ApplyStatus_CreateRetry), orderNo)
	return dao.FcTransfersApplyUpdateOrderIdAndStatus(applyId, int(entity.ApplyStatus_AuditOk), orderNo, 2)
}

func (srv *OrderBaseService) SendAuditFail(applyId int) error {
	result, err := dao.FcTransfersApplyById(applyId)
	if err != nil {
		log.Errorf("SendAuditFail error:%s", err.Error())
		return fmt.Errorf("取消交易状态异常：%s", err.Error())
	}
	if result.Status != int(entity.ApplyStatus_Auditing) {
		return fmt.Errorf("订单：%s,非待审核状态，当前订单状态：%s", result.OutOrderid, result.Status)
	}
	return dao.FcTransfersApplyUpdateStatusById(applyId, int(entity.ApplyStatus_AuditFail))
}

func (srv *OrderBaseService) GetApplyOrderByOrderNo(orderNo string) (*entity.FcTransfersApply, error) {
	return dao.FcTransfersApplyByOrderNo(orderNo)
}

func (srv *OrderBaseService) SendApplyTransferSuccess(applyId int) error {
	return dao.FcTransfersApplyUpdateStatusById(applyId, int(entity.ApplyStatus_TransferOk))
}

func (srv *OrderBaseService) GetApplyOrder(outOrderId string) (*entity.FcTransfersApply, error) {
	return dao.FcTransfersApplyByOutOrderNo(outOrderId)
}

func (srv *OrderBaseService) IsAllowRepush(outOrderId string) error {
	if outOrderId == "" {
		return errors.New("miss outOrderId")
	}
	applyOrder, _ := dao.FcTransfersApplyByOutOrderNo(outOrderId)
	if applyOrder == nil {
		return fmt.Errorf("没有找到订单：%s", outOrderId)
	}

	if global.WalletType(applyOrder.CoinName, applyOrder.AppId) == status.WalletType_Cold {
		order, err := dao.FcOrderFindByOutNo(outOrderId)
		if err != nil {
			if err != nil {
				err = fmt.Errorf("订单：%s,冷钱包当前状态不允许重推,查询异常:%s", outOrderId, err.Error())
			}
			return err
		}

		//冷钱包
		for _, v := range order {
			if v.Status <= status.BroadcastStatus.Int() {
				return fmt.Errorf("c:订单：%s,冷钱包当前状态不允许重推：status:%d", outOrderId, v.Status)
			}
		}

		if len(order) > 0 {
			if order[0].Status == status.BroadcastErrorStatus.Int() || order[0].Status == status.UnknowErrorStatus.Int() {
				//7，9需要人工审核
				return fmt.Errorf("订单：%s,冷钱包当前状态不允许重推,需要人工确认,广播错误:%s \n接收地址：%s \n ",
					outOrderId, order[0].ErrorMsg, order[0].ToAddress)
			}
		}

	} else {
		//尝试从orderHot表查询
		orderHot, err := dao.FcOrderHotFindByOutNo(outOrderId)
		if err != nil {
			if err != nil {
				err = fmt.Errorf("订单：%s,热钱包当前状态不允许重推,:%s", outOrderId, err.Error())
			}
			return err
		}
		for _, v := range orderHot {
			if v.Status <= int(status.BroadcastStatus) {
				return fmt.Errorf("h:订单：%s,热钱包当前状态不允许重推：status:%d", outOrderId, v.Status)
			}
		}
		if len(orderHot) > 0 {
			if orderHot[0].Status == status.BroadcastErrorStatus.Int() || orderHot[0].Status == status.UnknowErrorStatus.Int() {
				//7，9需要人工审核
				return fmt.Errorf("订单：%s,热钱包当前状态不允许重推,需要人工确认,广播错误:%s \n接收地址：%s \n",
					outOrderId, orderHot[0].ErrorMsg, orderHot[0].ToAddress)
			}
		}
	}
	//最终检查apply表状态
	if applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		//不是49状态不允许钉钉重推
		return fmt.Errorf("订单：%s,当前状态：%d,不允许重推", outOrderId, applyOrder.Status)
	}
	return nil
}

func (srv *OrderBaseService) NotifyToMchByOutOrderId(outOrderId string) error {
	applyOrder, err := dao.FcTransfersApplyByOutOrderNo(outOrderId)
	if err != nil {
		log.Errorf("FcTransfersApplyByOutOrderNo error:%s", err.Error())
		return err
	}
	walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)

	mch, err := dao.FcMchFindById(applyOrder.AppId)
	if err != nil {
		log.Errorf("查询商户信息失败：outOrderId:%s,mchid:%d", outOrderId, applyOrder.AppId)
		return err
	}

	tac, err := dao.GetApplyAddressByApplyCoinId(int64(applyOrder.Id), "to")
	if err != nil {
		return fmt.Errorf("dao.GetApplyAddressByApplyCoinId 出错 %v", err)
	}

	notify := &model.NotifyOrderToMch{
		Amount:            tac.ToAmount,
		Sfrom:             mch.Platform,
		OutOrderId:        outOrderId,
		OuterOrderNo:      outOrderId,
		Chain:             applyOrder.CoinName,
		CoinType:          applyOrder.CoinName,
		Txid:              "", // 下面补上
		Msg:               "", // 下面补上
		Memo:              applyOrder.Memo,
		ContractAddress:   applyOrder.Eostoken,
		Contract:          applyOrder.Eostoken,
		OrderSplitTxCount: 1,
		IsIn:              model.IsInType_Send,
		BlockHeight:       0,
		Confirmations:     0,
		ConfirmTime:       0,
		FromAddress:       "", // 下面补上
		ToAddress:         "", // 下面补上

		CoinName:  applyOrder.CoinName,
		TokenName: applyOrder.Eoskey,
	}
	if applyOrder.Eoskey != "" {
		notify.CoinType = applyOrder.Eoskey
	}
	switch walletType {
	case status.WalletType_Cold:
		orderCold, err := dao.FcOrderGetByOutOrderNo(outOrderId, int(status.BroadcastStatus))
		if err != nil {
			log.Errorf("FcOrderGetByOutOrderNo error:%s", err.Error())
			return err
		}
		notify.Txid = orderCold.TxId
		notify.Msg = "success"
		notify.FromAddress = orderCold.FromAddress
		notify.ToAddress = orderCold.ToAddress

	case status.WalletType_Hot:
		orderHot, err := dao.FcOrderHotGetByOutOrderNo(outOrderId, int(status.BroadcastStatus))
		if err != nil {
			log.Errorf("FcOrderHotGetByOutOrderNo error:%s", err.Error())
			return err
		}
		notify.Txid = orderHot.TxId
		notify.Msg = "success"
		notify.FromAddress = orderHot.FromAddress
		notify.ToAddress = orderHot.ToAddress
	default:
		err = fmt.Errorf("查询异常，币种类型配置文件缺少相关配置，币种：%s", applyOrder.CoinName)
		log.Errorf("%s", err.Error())
		return err
	}

	innerTxs := make([]model.TxPushInner, 0)
	innerTxs = append(innerTxs, model.TxPushInner{
		SeqNo:         md5Hash(applyOrder.OrderId),
		Status:        model.TxPushNormal,
		Amount:        tac.ToAmount,
		FromAddress:   notify.FromAddress,
		ToAddress:     notify.ToAddress,
		Memo:          applyOrder.Memo,
		Timestamp:     time.Now().Unix(),
		TransactionId: notify.Txid,
	})
	innerTxMs, _ := json.Marshal(innerTxs)
	notify.Txs = string(innerTxMs)

	dataByte, _ := json.Marshal(notify)
	//log.Infof("主动回调outOrderId:%s,Url：%s,发送内容：%s", outOrderId, applyOrder.CallBack, string(dataByte))
	respByte, err := util.PostByteForCallBack(applyOrder.CallBack, dataByte, mch.ApiKey, mch.ApiSecret)
	if err != nil {
		log.Infof("主动回调失败，outOrderId:%s,Url：%s,发送内容：%s，返回结果：%s",
			outOrderId, applyOrder.CallBack, string(dataByte), string(respByte))
		return err
	}
	log.Infof("主动回调成功，outOrderId:%s,Url：%s,发送内容：%s，返回结果：%s",
		outOrderId, applyOrder.CallBack, string(dataByte), string(respByte))

	//resp := model.DecodeBCallbackResp(respByte)
	//if resp.Code != 0 {
	//	log.Infof("主动回调解析code失败，outOrderId:%s,Url：%s,发送内容：%s，返回结果：%s",
	//		outOrderId, applyOrder.CallBack, string(dataByte), string(respByte))
	//	return errors.New("目标地址返回异常")
	//}
	//log.Infof("主动回调成功，outOrderId:%s,Url：%s,发送内容：%s，返回结果：%s",
	//	outOrderId, applyOrder.CallBack, string(dataByte), string(respByte))
	return nil

}

func md5Hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//保存到数据库即可，起一个定时扫描任务
func (srv *OrderBaseService) NotifyToMch(ta *entity.FcTransfersApply) error {
	var (
		err         error
		txid        string
		message     string
		order       *entity.FcOrder
		orderHot    *entity.FcOrderHot
		memoEncrypt string
		fromAddr    string
		toAddr      string
	)

	if ta == nil || ta.CallBack == "" {
		err = fmt.Errorf("订单[%s],缺少回调地址", ta.OutOrderid)
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
		return err
	}

	if global.WalletType(ta.CoinName, ta.AppId) == status.WalletType_Cold {
		order, err = dao.FcOrderFindByOrderId(ta.OrderId)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("订单[%s],保存回调信息失败，err=[%s]", ta.OutOrderid, err.Error()))
			dingding.ErrTransferDingBot.NotifyStr(err.Error())
			return err
		}
		memoEncrypt = order.MemoEncrypt
		if order.Status == status.BroadcastStatus.Int() {
			message = "success"
			txid = order.TxId
		} else {
			message = "fail"
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单[%s],数据异常:%s", ta.OutOrderid, order.ErrorMsg))
			if txid != "" {
				dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单[%s],数据异常,txid存在[%s],但是数据状态异常[%d]",
					ta.OutOrderid, order.TxId, order.Status))
			}
			return errors.New("数据异常")
		}
		fromAddr = order.FromAddress
		toAddr = order.ToAddress
	} else {
		orderHot, err = dao.FcOrderHotFindByOrderId(ta.OrderId)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("订单[%s],保存回调信息失败，err=[%s]", ta.OutOrderid, err.Error()))
			dingding.ErrTransferDingBot.NotifyStr(err.Error())
			return err
		}
		memoEncrypt = orderHot.MemoEncrypt
		if orderHot.Status == status.BroadcastStatus.Int() {
			message = "success"
			txid = orderHot.TxId
		} else {
			message = "fail"
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单[%s],数据异常:%s", ta.OutOrderid, orderHot.ErrorMsg))
			if txid != "" {
				dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单[%s],数据异常,txid存在[%s],但是数据状态异常[%d]",
					ta.OutOrderid, orderHot.TxId, orderHot.Status))
			}
			return fmt.Errorf("数据异常，广播状态为： %d", orderHot.Status)
		}
		fromAddr = orderHot.FromAddress
		toAddr = orderHot.ToAddress
	}

	// 未知的历史遗留问题：BTC、LTC等币种在fc_order表to地址后面多了个空格
	// 这里需要去掉空格
	if len(toAddr) > 0 {
		toAddr = strings.TrimSpace(toAddr)
	}

	txStatus := &entity.FcTransactionsState{
		AppId:       ta.AppId,
		Sfrom:       "",
		CoinName:    ta.CoinName,
		OrderId:     ta.OrderId,
		OutOrderid:  ta.OutOrderid,
		Txid:        txid,
		CallBack:    ta.CallBack,
		Msg:         message,
		Eoskey:      ta.Eoskey,
		Eostoken:    ta.Eostoken,
		Memo:        ta.Memo,
		MemoEncrypt: memoEncrypt,
		Data:        "",
		RetryNum:    0,
		CallbackMsg: "",
		CreateTime:  util.GetChinaTimeNow().Unix(),
		Lastmodify:  util.GetChinaTimeNow(),
		Status:      entity.FcTransactionsStatesWait, //默认等待状态
		PushStatus:  1,
	}

	mch, err := dao.FcMchFindById(ta.AppId)
	if err != nil {
		err = fmt.Errorf("毁掉保存，商户信息查询失败,appid:%d，error=[%s]", ta.AppId, err.Error())
		log.Errorf(err.Error())
		return err
	}

	//填充商户名
	txStatus.Sfrom = mch.Platform

	tac, err := dao.GetApplyAddressByApplyCoinId(int64(ta.Id), "to")
	if err != nil {
		return fmt.Errorf("dao.GetApplyAddressByApplyCoinId 出错 %v", err)
	}

	//组装发送数据
	notify := &model.NotifyOrderToMch{
		Amount:            tac.ToAmount,
		Sfrom:             mch.Platform,
		OutOrderId:        ta.OutOrderid,
		OuterOrderNo:      ta.OutOrderid,
		Chain:             ta.CoinName,
		CoinType:          ta.CoinName,
		Txid:              txid,
		Msg:               message,
		Memo:              ta.Memo,
		MemoEncrypt:       memoEncrypt,
		ContractAddress:   ta.Eostoken,
		Contract:          ta.Eostoken,
		OrderSplitTxCount: 1,
		IsIn:              model.IsInType_Send,
		BlockHeight:       0,
		Confirmations:     0,
		ConfirmTime:       0,
		FromAddress:       fromAddr,
		ToAddress:         toAddr,

		CoinName:  ta.CoinName,
		TokenName: ta.Eoskey,
	}
	if ta.Eoskey != "" {
		notify.CoinType = ta.Eoskey
	}

	// 设置链和币种的全局ID
	chainUnionCoinSet, err := dao.GetFcCoinSetByName(notify.Chain)
	if err != nil {
		return err
	}
	notify.CoinUnionId = chainUnionCoinSet.UnionId

	if notify.Chain != notify.CoinType {
		// 如果是代币交易，去数据库查询代币对应的全局ID
		coinTypeUnionCoinSet, ctErr := dao.GetFcCoinSetByName(notify.CoinType)
		if ctErr != nil {
			return ctErr
		}
		notify.CoinTypeUnionId = coinTypeUnionCoinSet.UnionId
	} else {
		// 如果是主链币交易，直接使用主链的全局ID
		notify.CoinTypeUnionId = chainUnionCoinSet.UnionId
	}

	innerTxs := make([]model.TxPushInner, 0)
	innerTxs = append(innerTxs, model.TxPushInner{
		SeqNo:         md5Hash(ta.OrderId),
		Status:        model.TxPushNormal,
		Amount:        tac.ToAmount,
		FromAddress:   fromAddr,
		ToAddress:     toAddr,
		Memo:          ta.Memo,
		Timestamp:     time.Now().Unix(),
		TransactionId: txid,
	})

	innerTxMs, _ := json.Marshal(innerTxs)
	notify.Txs = string(innerTxMs)

	notufyData, _ := json.Marshal(notify)
	//填充data
	txStatus.Data = string(notufyData)
	return dao.FcTransactionsStateInsert(txStatus)

}

func (srv *OrderBaseService) SendApplyReviewOk(applyId int) error {
	result, err := dao.FcTransfersApplyById(applyId)
	if err != nil {
		log.Errorf("SendApplyReviewOk error:%s", err.Error())
		return fmt.Errorf("重置交易状态异常：%s", err.Error())
	}
	if result.Status != int(entity.ApplyStatus_Auditing) || result.Status != int(entity.ApplyStatus_CreateFail) {
		orderId := fmt.Sprintf("%s_%d", result.OrderId, time.Now().Unix())
		return dao.FcTransfersApplyUpdateOrderIdAndStatus(applyId, int(entity.ApplyStatus_AuditOk), orderId, 2)

	}

	return fmt.Errorf("订单：%s,当前状态不允许审核，状态：%d", result.OutOrderid, result.Status)
}

func (srv *OrderBaseService) SendApplyWait(applyId int) error {
	return dao.FcTransfersApplyUpdateStatusById(applyId, int(entity.ApplyStatus_Creating))
}

func (srv *OrderBaseService) SendApplyCreateSuccess(applyId int) error {
	return dao.FcTransfersApplyUpdateStatusById(applyId, int(entity.ApplyStatus_CreateOk))
}

func (srv *OrderBaseService) SendApplyRetry(applyId int) error {
	data, err := dao.FcTransfersApplyById(applyId)
	if err != nil {
		return err
	}
	//查询交易地址
	addrInfos, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(applyId, "to")
	if err != nil {
		return err
	}
	if len(addrInfos) == 0 {
		return errors.New("empty address info ")
	}

	orderNo, err := util.GetOrderId(data.OutOrderid, addrInfos[0].Address, addrInfos[0].ToAmount)
	if err != nil {
		return err
	}

	//return dao.FcTransfersApplyUpdateStatusAddErr(applyId, int(entity.ApplyStatus_CreateRetry), orderNo)
	//return dao.FcTransfersApplyUpdateStatusAddErr(applyId, int(entity.ApplyStatus_CreateRetry), orderNo)
	return dao.FcTransfersApplyUpdateStatusAddErr(applyId, int(entity.ApplyStatus_CreateFail), orderNo)
}

func (srv *OrderBaseService) SendApplyFail(applyId int) error {
	//data, err := dao.FcTransfersApplyById(applyId)
	//if err != nil {
	//	return err
	//}
	//查询交易地址
	//addrInfos, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(applyId, "to")
	//if err != nil {
	//	return err
	//}
	//if len(addrInfos) == 0 {
	//	return errors.New("empty address info ")
	//}

	//orderNo, err := util.GetOrderId(data.OutOrderid, addrInfos[0].Address, addrInfos[0].ToAmount)
	//if err != nil {
	//	return err
	//}

	return dao.FcTransfersApplyFail(applyId, int(entity.ApplyStatus_CreateFail))
}

func (srv *OrderBaseService) SaveTransferByUtxo(params model.TransferParams, callBackUrl string) (orderId int64, err error) {
	coinName := params.CoinName
	if params.TokenName != "" {
		coinName = params.TokenName
	}

	//币种信息
	coinSet, err := dao.FcCoinSetGetByName(coinName, 1)
	if err != nil {
		return httpresp.UnsupportedToken, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.UnsupportedToken), coinName)
	}

	//商户信息
	mch, err := dao.FcMchFindByPlatform(params.Sfrom)
	if err != nil {
		return httpresp.SFROM_ERROR, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.SFROM_ERROR), params.Sfrom)
	}

	//生成特殊的订单ID
	orderNoStr, err := util.GetOrderId(params.OutOrderId, params.ToAddress, params.Amount.String())
	if err != nil {
		log.Errorf("订单：%s,生成订单no失败,err:%s", params.OutOrderId, err.Error())
		return httpresp.FAIL, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.FAIL), params.Sfrom)

	}
	ta := &entity.FcTransfersApply{
		Username:   "api",
		OrderId:    orderNoStr,
		Applicant:  params.Sfrom,
		AppId:      mch.Id,
		CallBack:   callBackUrl,
		OutOrderid: params.OutOrderId,
		CoinName:   params.CoinName,
		Type:       "cz",
		Memo:       params.Memo,
		Eoskey:     params.TokenName,
		Eostoken:   params.ContractAddress,
		Fee:        params.Fee.String(),
		Status:     int(entity.ApplyStatus_AuditOk),
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     params.ToAddress,
		AddressFlag: "to",
		ToAmount:    params.Amount.String(),
		Lastmodify:  util.GetChinaTimeNow(),
	}

	holdAmount, _ := decimal.NewFromString(coinSet.HoldAmount)
	if holdAmount.Equals(decimal.Zero) {
		//请系统设置拦截金额
		log.Errorf("请系统设置拦截金额，币种：%s", coinSet.Name)
		return 0, errors.New("服务器繁忙")

	}
	log.Infof("holdAmount:%s,params.Amount:%s", holdAmount.String(), params.Amount.String())
	if params.Amount.GreaterThanOrEqual(holdAmount) {
		//金额风险
		ta.Status = int(entity.ApplyStatus_Auditing) //需要审核
	}

	//不存储change，找零需要计算完utxo才知道
	//taChange := &entity.FcTransfersApplyCoinAddress{
	//	ApplyCoinId: coinSet.Id,
	//	Address:     params.ToAddress,
	//	AddressFlag: "change",
	//	ToAmount:    0,
	//}
	orderId, err = dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo})
	if err != nil {
		return 0, err
	}
	if ta.Status == int(entity.ApplyStatus_Auditing) {

		tpl := "金额风控提醒，请客服注意审核:\n" +
			"商户：%s \n" +
			"出账订单：%s \n" +
			"主链币种：%s \n" +
			"出账金额：%s \n" +
			"接收地址：%s \n"

		tpl = fmt.Sprintf(
			tpl,
			mch.Platform,
			params.OutOrderId,
			params.CoinName,
			params.Amount.String(),
			params.ToAddress,
		)
		if params.ContractAddress != "" {
			tpl = tpl + "合约地址：%s \n"
			tpl = fmt.Sprintf(
				tpl,
				params.ContractAddress,
			)
		}
		if params.TokenName != "" {
			tpl = tpl + "代币名称：%s \n"
			tpl = fmt.Sprintf(
				tpl,
				params.TokenName,
			)
		}
		//钉钉通知审核
		dingding.ReviewDingBot.NotifyStr(tpl)
		wechat.SendWarnInfo(tpl)

	}
	return orderId, err

}

func verifyMchBalance(coinName, contractAddress string, transferAmount decimal.Decimal, mchId int) (ok bool, mchBalance decimal.Decimal, err error) {
	log.Infof("start VerifyMchBalance coinName: %s,contractAddress: %s, transferAmount: %s, mchId: %d ", coinName, contractAddress, transferAmount.String(), mchId)
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,查询币种异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	result, err := dao.FcMchAmountGetByACId(mchId, coinResult.Id)
	if err != nil {
		if err.Error() == "Not Fount!" {
			return false, decimal.Zero, fmt.Errorf("商户余额不足，余额0")
		}
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,查询余额异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	balance, err := decimal.NewFromString(result.Amount)
	if err != nil {
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,转换余额异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	if contractAddress == "" {
		if transferAmount.GreaterThanOrEqual(balance) {
			// 主链币
			return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s (尚未扣除手续费)", balance.String(), transferAmount.String())
		}
	} else {
		log.Infof("else VerifyMchBalance coinName: %s,contractAddress: %s, transferAmount: %s, mchId: %d, balance: %s ", coinName, contractAddress, transferAmount.String(), mchId, balance.String())
		if transferAmount.GreaterThan(balance) {
			return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s", balance.String(), transferAmount.String())
		}
	}

	return true, balance, nil

}

func (srv *OrderBaseService) SaveTransferByAccount(params model.TransferParams, callBackUrl string) (applyId int64, err error) {
	log.Infof("SaveTransferByAccount : %+v", params)
	coinName := params.CoinName
	if params.TokenName != "" {
		coinName = params.TokenName
	}

	//币种信息
	coinSet, err := dao.FcCoinSetGetByName(coinName, 1)
	if err != nil {
		return httpresp.UnsupportedToken, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.UnsupportedToken), coinName)
	}

	//商户信息
	mch, err := dao.FcMchFindByPlatform(params.Sfrom)
	if err != nil {
		return httpresp.SFROM_ERROR, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.SFROM_ERROR), params.Sfrom)
	}
	log.Infof("SaveTransferByAccount 根据sfrom查询到mch信息: +%v", mch)

	//查询主链币冷地址
	froms, err := dao.FcGenerateAddressListFindAddresses(1, 2, mch.Id, params.CoinName)
	if err != nil || len(froms) == 0 {
		return httpresp.ADR_NONE, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.ADR_NONE), params.Sfrom)
	}

	if coinName == "welups" {
		ok, _, err := verifyMchBalance(params.CoinName, params.ContractAddress, params.Amount, mch.Id)
		if err != nil {
			log.Errorf("1.验证商户余额异常,币种：%s,商户：%s, error: %s", coinName, params.Sfrom, err.Error())
			return httpresp.MchAmountNotEnough, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.MchAmountNotEnough), params.Sfrom)
		}
		if !ok {
			log.Errorf("2.验证商户余额异常,币种：%s,商户：%s", coinName, params.Sfrom)
			return httpresp.MchAmountNotEnough, fmt.Errorf(httpresp.GetMsg(httpresp.MchAmountNotEnough))
		}
	}

	//redisHelper, _ := util.AllocRedisClient()
	//cache, _ := redisHelper.Get("outcollect")
	//if cache == "1" {
	//	addrAmts, _ := dao.FcAddressAmountByAddrs(froms, mch.Id, coinName)
	//	//if err != nil {
	//	//	return httpresp.ADR_NONE, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.ADR_NONE), params.Sfrom)
	//	//}
	//	log.Infof("[转账前判断余额]查询到addressAmount数据条数为: %d", len(addrAmts))
	//	has := false
	//	for _, addrAmt := range addrAmts {
	//		amtDecimal, _ := decimal.NewFromString(addrAmt.Amount)
	//		if amtDecimal.Cmp(params.Amount) == -1 {
	//			log.Infof("[转账前判断余额] 币种 %s 地址 %s 余额%s 需要转账%s，余额不足", coinName, addrAmt.Address, addrAmt.Amount, params.Amount)
	//		} else {
	//			has = true
	//			break
	//		}
	//	}
	//	if !has {
	//		log.Infof("[转账前判断余额] 币种 %s 需要转账%s，%d个出账地址都余额不足，不能转账", coinName, params.Amount, len(addrAmts))
	//	}
	//}

	//生成特殊的订单ID
	orderNoStr, err := util.GetOrderId(params.OutOrderId, params.ToAddress, params.Amount.String())
	if err != nil {
		log.Errorf("订单：%s,生成订单no失败", params.OutOrderId)
		return httpresp.FAIL, fmt.Errorf("%s:%s", httpresp.GetMsg(httpresp.FAIL), params.Sfrom)

	}

	if coinName == "mdu" {
		//todo mdu签名程序orderNoStr 有数据长度问题，尚未解决
		orderNoStr = util.GetUUID()
	}

	isMultiFromTransaction := IsMultiFromVersion(params.CoinName) && !params.IsPseudCustody
	ta := &entity.FcTransfersApply{
		Username:   "api",
		OrderId:    orderNoStr,
		Applicant:  params.Sfrom,
		AppId:      mch.Id,
		CallBack:   callBackUrl,
		OutOrderid: params.OutOrderId,
		CoinName:   params.CoinName,
		Type:       "cz",
		Memo:       params.Memo,
		Eoskey:     params.TokenName,
		Eostoken:   params.ContractAddress,
		Fee:        params.Fee.String(),
		Status:     int(entity.ApplyStatus_AuditOk),
		Createtime: time.Now().Unix(),
	}
	if isMultiFromTransaction {
		ta.TxType = entity.MultiAddrTx
	} else {
		ta.TxType = entity.SingleAddrTx
	}

	if IsNewCollectVersion(params.CoinName) && !isMultiFromTransaction {
		log.Info("outcollect 命中")
		_, err := CheckIfNeedCollect(mch.Id, params.OutOrderId, params.CoinName, params.TokenName, params.ContractAddress, coinSet.CollectThreshold, params.Amount)
		if err != nil {
			errMsg := fmt.Sprintf("订单=%s 尝试进行出账归集失败，币种[%s %s]：%v", params.OutOrderId, params.CoinName, params.TokenName, err)
			dingding.ErrTransferDingBot.NotifyStr(errMsg)
			log.Error(errMsg)
		}
	}

	holdAmount, _ := decimal.NewFromString(coinSet.HoldAmount)
	if holdAmount.Equals(decimal.Zero) {
		//请系统设置拦截金额
		log.Errorf("请系统设置拦截金额，币种：%s", coinSet.Name)
		return 0, errors.New("服务器繁忙")

	}
	log.Infof("holdAmount:%s,params.Amount:%s", holdAmount.String(), params.Amount.String())
	if params.Amount.GreaterThanOrEqual(holdAmount) {
		//金额风险
		ta.Status = int(entity.ApplyStatus_Auditing) //需要审核
	} else {
		if isMultiFromTransaction {
			log.Info("IsMultiFromVersion 走多地址出账")
			// 直接将状态设置为43，让多地址出账程序拉取处理
			ta.Status = int(entity.ApplyStatus_Creating)
		}
	}

	// 出账不存储from 因为异步的关系 from在之后不一定还够余额
	//tacFrom := &entity.FcTransfersApplyCoinAddress{
	//	ApplyCoinId: coinSet.Id,
	//	Address:     froms[0],
	//	AddressFlag: "from",
	//}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId:    coinSet.Id,
		Address:        params.ToAddress,
		AddressFlag:    "to",
		ToAmount:       params.Amount.String(),
		Lastmodify:     util.GetChinaTimeNow(),
		BanFromAddress: params.BanFromAddress,
	}
	applyId, err = dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo})
	if err != nil {
		return 0, err
	}
	if ta.Status == int(entity.ApplyStatus_Auditing) {
		//钉钉通知审核
		tpl := "金额风控提醒，请客服注意审核:\n" +
			"商户：%s \n" +
			"出账订单：%s \n" +
			"主链币种：%s \n" +
			"出账金额：%s \n" +
			"接收地址：%s \n"

		tpl = fmt.Sprintf(
			tpl,
			mch.Platform,
			params.OutOrderId,
			params.CoinName,
			params.Amount.String(),
			params.ToAddress,
		)
		if params.ContractAddress != "" {
			tpl = tpl + "合约地址：%s \n"
			tpl = fmt.Sprintf(
				tpl,
				params.ContractAddress,
			)
		}
		if params.TokenName != "" {
			tpl = tpl + "代币名称：%s \n"
			tpl = fmt.Sprintf(
				tpl,
				params.TokenName,
			)
		}
		//钉钉通知审核
		dingding.ReviewDingBot.NotifyStr(tpl)
		wechat.SendWarnInfo(tpl)
	}

	if isMultiFromTransaction && ta.Status == int(entity.ApplyStatus_Creating) {
		if err = PushToWaitingList(params.OutOrderId); err != nil {
			log.Infof("尝试将订单%s 推入缓存等待处理失败 %v", params.OutOrderId, err)
		}
	}

	return applyId, err

}

func NewOrderBaseService() service.OrderService {
	return &OrderBaseService{}
}
