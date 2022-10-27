package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"xorm.io/builder"
)

func FcOrderTxsPushInsert(order *entity.FcOrderTxsPush) error {
	_, err := db.Conn.InsertOne(order)
	return err
}

func FindOrderTxsPushByOrderTxIds(orderTxIds []int64) ([]entity.FcOrderTxsPush, error) {
	results := make([]entity.FcOrderTxsPush, 0)
	err := db.Conn.Where(builder.In("order_txs_id", orderTxIds)).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func GetOrderTxsPushOrderTxsId(orderTxsId int64) (*entity.FcOrderTxsPush, error) {
	tx := &entity.FcOrderTxsPush{}
	has, err := db.Conn.Where("order_txs_id = ?", orderTxsId).Get(tx)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return tx, nil
}

func FcOrderTxsPushUpdate(orderTxsId int64, blockHeight int64, confirmation int, confirmTime int64) error {
	model := &entity.FcOrderTxsPush{BlockHeight: blockHeight, Confirmation: confirmation, ConfirmTime: confirmTime}
	_, err := db.Conn.Where("orderTxsId = ?", orderTxsId).Update(model)
	if err != nil {
		return err
	}
	return nil
}
