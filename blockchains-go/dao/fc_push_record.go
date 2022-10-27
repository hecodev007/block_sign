package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FcPushRecordLast(txid string, coin, cointype string, appId int) (*entity.FcPushRecord, error) {
	order := &entity.FcPushRecord{}
	has, err := db.Conn.Where("tx_id = ? and coin = ? and coin_type = ? and app_id = ? and status = 1", txid, coin, cointype, appId).Desc("confirmations").Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}
