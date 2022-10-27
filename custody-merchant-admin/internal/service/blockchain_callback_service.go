package service

import (
	"custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

//充值回调

func InComeCallBack(back *domain.InComeBack) error {
	log.Errorf("InComeCallBack %+v\n", back)
	var err error
	//查找 地址对应的商户
	// 新增充值订单和交易记录
	confirmTime := time.Unix(back.ConfirmTime, 0)
	if TimeIsNull(confirmTime) {
		confirmTime = time.Now().Local()
	}
	height := int(back.BlockHeight)
	minerFee := back.Fee
	if back.Coin == "" {
		log.Error("币种名不能为空")
		return errors.New("币种名不能为空")
	}
	if back.Txid == "" {
		log.Error("交易TXID不能为空")
		return errors.New("交易TXID不能为空")
	}
	//判断txid 是否已经处理
	key := fmt.Sprintf("%s-custody:in-txid:%v", config.Conf.Mod, back.Txid)
	var v string
	err = cache.GetRedisClientConn().Get(key, &v)
	if v != "" { //txid 已经处理
		log.Error("交易TXID已经处理1\n")
		return nil
	}
	//判断流水记录有没有处理
	billDao := new(bill.BillDetail)
	bil, err := billDao.GetBillByTxIdState(back.Txid, 1)
	if bil.Id != 0 {
		log.Error("交易TXID已经处理2\n")
		return nil
	}
	coinName := back.CoinType
	if coinName == "" {
		coinName = back.Coin
	}
	wallet := domain.MqWalletInfo{
		FromAddress: back.FromAddress,
		ToAddress:   back.ToAddress,
		TxId:        back.Txid,
		SerialNo:    xkutils.Generate("HF", time.Now().Local()),
		CoinName:    coinName,
		Memo:        back.Memo,
		Nums:        decimal.NewFromFloat(back.Amount),
		Height:      height,
		ConfirmNums: int(back.Confirmations),
		ConfirmTime: confirmTime,
		RealNums:    decimal.NewFromFloat(back.Amount),
		Destroy:     decimal.Zero,
		BurnFee:     decimal.Zero,
		MinerFee:    decimal.NewFromFloat(minerFee),
	}
	log.Errorf("InComeCallBack wallet %+v\n", wallet)
	err = ReceiveBillDetail(wallet)
	if err != nil {
		log.Errorf("InComeCallBack ReceiveBillDetail err %+v\n", err)
		log.Error(err.Error())
		return err
	}
	return nil
}
