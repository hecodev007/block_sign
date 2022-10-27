package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"xorm.io/builder"
)

type FcFeeCount struct {
	Count int64 `json:"count" xorm:"not null INT"`
}

type FcAddressFeeCount struct {
	Address string `json:"address" xorm:"not null VARCHAR(256)"`
	Count   int64  `json:"count" xorm:"not null INT"`
}

type FcAddressReceive struct {
	Address string `json:"address" xorm:"not null VARCHAR(256)"`
	Amount  string `json:"amount" xorm:"not null INT"`
}

// 获取指定时间的手续费笔数
func FcOrderTotalFeeCount(beginTime int64, chainName string) (*FcFeeCount, error) {
	fee := &FcFeeCount{}
	has, err := db.Conn.SQL("SELECT  COUNT(1) AS count FROM fc_order where outer_order_no LIKE 'FEE%' AND coin_name = ? and create_at > ? and status = 4", chainName, beginTime).Get(fee)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return fee, nil
}

// 获取指定时间前20条打手续费条数大于10的地址
func FcOrderAddressFeeCount(beginTime int64, chainName string, limitCount int, limit int) ([]*FcAddressFeeCount, error) {
	results := make([]*FcAddressFeeCount, 0)
	err := db.Conn.SQL("SELECT * FROM (SELECT to_address AS address, COUNT(1) as count FROM fc_order where outer_order_no LIKE 'FEE%'  AND coin_name = ?  and create_at > ? and status = 4  group by to_address) as t  where t.count > ?  ORDER BY t.count DESC", chainName, beginTime, limitCount).Limit(limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// 查看给定地址在一定时间内是否有主链币入账金额情况
func FcOrderFindReceive(beginTime int64, chainName string, addrs []string) ([]*FcAddressReceive, error) {
	results := make([]*FcAddressReceive, 0)
	err := db.Conn.Select("to_address AS address,SUM(amount) AS amount").Table("fc_tx_transaction_new").
		Where(builder.Eq{
			"coin":      chainName,
			"coin_type": chainName,
			"tx_type":   1,
			"create_at": beginTime,
		}, builder.In("to_address", addrs)).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcOrderGetByTxid(txid string, status int) (*entity.FcOrder, error) {
	order := &entity.FcOrder{}
	has, err := db.Conn.Where("tx_id = ? and status = ?", txid, status).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}

/*
func: 根据txid以及out_order_no查询Order表
auth： flynn
date： 2020-07-03
*/
func FcOrderGetByOutOrderNoAndTxid(outOrderNo, txid string, status int) (*entity.FcOrder, error) {
	order := &entity.FcOrder{}
	has, err := db.Conn.Where("outer_order_no = ? and tx_id = ? and status = ?", outOrderNo, txid, status).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}

func FcOrderGetByOutOrderNo(outOrderNo string, status int) (*entity.FcOrder, error) {
	order := &entity.FcOrder{}
	has, err := db.Conn.Where("outer_order_no = ? and status = ?", outOrderNo, status).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}

func FcOrderHaveByOutOrderNo(outOrderNo string, status int) bool {
	has, err := db.Conn.Table("fc_order").Where("outer_order_no = ? and status <= ?", outOrderNo, status).Count()
	if err != nil {
		return false
	}
	if has > 0 {
		return true
	}

	return false
}

//根据状态查询订单
func FcOrderFindByNoAndStatus(outOrderNo string, status status.OrderStatus) ([]*entity.FcOrder, error) {
	results := make([]*entity.FcOrder, 0)
	err := db.Conn.Where("outer_order_no = ? and status = ?", outOrderNo, status).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//根据外部订单号查询订单
func FcOrderFindByOutNo(outOrderNo string) ([]*entity.FcOrder, error) {
	results := make([]*entity.FcOrder, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).Desc("create_at").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//根据状态查询订单
func FcOrderFindByOrderId(orderId string) (*entity.FcOrder, error) {
	result := new(entity.FcOrder)
	has, err := db.Conn.Where("order_no = ?", orderId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

//根据状态外部订单和商户名字
func FcOrderFindListByOutNo(outOrderNo string) ([]*entity.FcOrder, error) {
	result := make([]*entity.FcOrder, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).Find(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func FcOrderFindSuccessOrder(outOrderNo string) (*entity.FcOrder, error) {
	results := new(entity.FcOrder)
	has, err := db.Conn.Where("outer_order_no = ? and status = ?", outOrderNo, status.BroadcastStatus).Get(results)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return results, nil
}

//根据内部订单号修改状态
func FcOrderUpdateState(orderNo string, status int) error {
	var (
		affected int64 = 0
	)
	res, err := db.Conn.Exec("update fc_order set status = ? where order_no = ? ", status, orderNo)
	if err != nil {
		return err
	}
	affected, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return err
	}
	return err
}

func FcOrderUpdateTxIdAndStatus(id int, txId string) error {
	var (
		affected int64 = 0
	)
	res, err := db.Conn.Exec("update fc_order set status = ?,tx_id = ? where id = ? LIMIT 1", status.BroadcastStatus, txId, id)
	if err != nil {
		return err
	}
	affected, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return err
	}
	return err
}

func FcOrderUpdateState2(outOrderNo, txid string, status int) error {
	var (
		affected int64 = 0
	)
	res, err := db.Conn.Exec("update fc_order set status = ? where outer_order_no = ? and tx_id = ?", status, outOrderNo, txid)
	if err != nil {
		return err
	}
	affected, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return err
	}
	return err
}

func FcOrderInsert(order *entity.FcOrder) error {
	_, err := db.Conn.InsertOne(order)
	return err
}
