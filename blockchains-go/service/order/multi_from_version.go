package order

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"strings"
	"time"
)

func PushToWaitingList(outerOrderNo string) error {
	if err := redis.Client.ListRPush(redis.CacheKeyWaitingOrderList, outerOrderNo); err != nil {
		return err
	}
	log.Infof("PushToWaitingList 订单%s推入待处理缓存列表成功", outerOrderNo)
	return nil
}

func IsMultiFromVersion(chain string) bool {
	chain = strings.ToLower(chain)

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Errorf("AllocRedisClient 出错 %v", err)
	}
	defer redisHelper.Close()

	cacheCoins, _ := redisHelper.Get("multiversion")
	split := strings.Split(cacheCoins, ",")
	exist := false
	for _, s := range split {
		if s == strings.ToLower(chain) {
			exist = true
		}
	}

	return exist
}

func (srv *OrderBaseService) NotifyToMchV2(ta *entity.FcTransfersApply, order *entity.FcOrder, orderTx *entity.FcOrderTxs) error {
	var err error
	if ta.CallBack == "" {
		err = fmt.Errorf("订单[%s],缺少回调地址", ta.OutOrderid)
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
		return err
	}

	txStatus := &entity.FcTransactionsState{
		AppId:       ta.AppId,
		Sfrom:       "",
		CoinName:    ta.CoinName,
		OrderId:     ta.OrderId,
		OutOrderid:  ta.OutOrderid,
		Txid:        orderTx.TxId,
		CallBack:    ta.CallBack,
		Msg:         "success",
		Eoskey:      ta.Eoskey,
		Eostoken:    ta.Eostoken,
		Memo:        ta.Memo,
		MemoEncrypt: "",
		Data:        "",
		RetryNum:    0,
		CallbackMsg: "",
		CreateTime:  util.GetChinaTimeNow().Unix(),
		Lastmodify:  util.GetChinaTimeNow(),
		Status:      entity.FcTransactionsStatesWait, //默认等待状态
		PushStatus:  1,
	}

	push := &entity.FcOrderTxsPush{
		OrderTxsId: orderTx.Id,
		TxId:       orderTx.TxId,
		IsIn:       int(model.IsInType_Send),
	}

	dao.FcOrderTxsPushInsert(push)

	allTxs, err := dao.FindOrderTxsValid(order.OuterOrderNo)
	if err != nil {
		return fmt.Errorf("根据outerOrderNo(%s)获取订单交易失败 %v", order.OuterOrderNo, err)
	}

	needPushTxs := make([]entity.FcOrderTxs, 0)
	log.Infof("获取到已完成的订单交易 %+v", allTxs)
	orderTxIds := make([]int64, 0)
	splitCount := 0
	for _, tx := range allTxs {
		// 状态为 13 也需要推送给交易所
		if tx.NeedPush() {
			needPushTxs = append(needPushTxs, tx)
			orderTxIds = append(orderTxIds, tx.Id)
		}
		if !tx.IsChainFailure() && !tx.IsCanceled() {
			splitCount += 1
		}
	}
	log.Infof("splitCount = %d", splitCount)

	pushList, err := dao.FindOrderTxsPushByOrderTxIds(orderTxIds)
	if err != nil {
		return fmt.Errorf("调用dao.FindOrderTxsPushByOrderTxIds失败 %v", err)
	}

	if len(pushList) != len(needPushTxs) {
		return fmt.Errorf("outerOrderNo(%s) 订单交易条数与推送记录条数不一致", order.OuterOrderNo)
	}

	//组装发送数据
	notify := &model.NotifyOrderToMch{
		Msg:               "success",
		Amount:            order.TotalAmount,
		Sfrom:             orderTx.Mch,
		OrderSplitTxCount: splitCount,
		IsIn:              model.IsInType_Send,
		OuterOrderNo:      ta.OutOrderid,
		OutOrderId:        ta.OutOrderid,
		Chain:             ta.CoinName,
		CoinName:          ta.CoinName,
		CoinType:          ta.CoinName,
		Contract:          ta.Eostoken,
		ContractAddress:   ta.Eostoken,
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

	isCareOutSide := len(needPushTxs) == 1
	if isCareOutSide {
		notify.Txid = needPushTxs[0].TxId
		notify.TransactionId = needPushTxs[0].TxId
	}

	innerTxs := make([]model.TxPushInner, 0)
	for _, tx := range needPushTxs {
		sta := model.TxPushNormal
		if tx.IsChainFailure() {
			sta = model.TxPushFailure
		}

		tp := model.TxPushInner{
			SeqNo:         tx.SeqNo,
			Status:        sta,
			Amount:        tx.Amount,
			FromAddress:   tx.FromAddress,
			ToAddress:     tx.ToAddress,
			Timestamp:     time.Now().Unix(),
			TransactionId: tx.TxId,
		}

		p := getOrderTxPush(tx.Id, pushList)
		if p != nil {
			tp.Memo = p.Memo
			tp.ConfirmTime = p.ConfirmTime
			tp.Confirmations = int64(p.Confirmation)
			tp.BlockHeight = p.BlockHeight
			tp.TrxN = p.TrxN
			tp.Fee = p.Fee
		}
		innerTxs = append(innerTxs, tp)
	}

	innerTxsMs, _ := json.Marshal(innerTxs)
	notify.Txs = string(innerTxsMs)

	ms, _ := json.Marshal(notify)
	txStatus.Data = string(ms)
	return dao.FcTransactionsStateInsert(txStatus)
}

func getOrderTxPush(orderTxId int64, pushList []entity.FcOrderTxsPush) *entity.FcOrderTxsPush {
	for _, p := range pushList {
		if p.OrderTxsId == orderTxId {
			return &p
		}
	}
	return nil
}
