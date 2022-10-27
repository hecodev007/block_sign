package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model/status"
)

func FcOrderHotGetByTxid(txid string, status int) (*entity.FcOrderHot, error) {
	order := &entity.FcOrderHot{}
	has, err := db.Conn.Where("tx_id = ? and status = ?", txid, status).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}
func FcOrderHotGetByTxid2(txid string) (*entity.FcOrderHot, error) {
	order := &entity.FcOrderHot{}
	has, err := db.Conn.Where("tx_id = ? ", txid).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Found!")
	}
	return order, nil
}

func FcOrderHotGetByOutOrderNo(outOrderNo string, status int) (*entity.FcOrderHot, error) {
	order := &entity.FcOrderHot{}
	has, err := db.Conn.Where("outer_order_no = ? and status = ?", outOrderNo, status).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}

func FcOrderHotHaveByOutOrderNo(outOrderNo string, status int) bool {
	has, err := db.Conn.Table("fc_order_hot").Where("outer_order_no = ? and status <= ?", outOrderNo, status).Count()
	if err != nil {
		return false
	}
	if has > 0 {
		return true
	}

	return false
}

//根据状态查询订单
func FcOrderHotFindByNoAndStatus(outOrderNo string, status status.OrderStatus) ([]*entity.FcOrderHot, error) {
	results := make([]*entity.FcOrderHot, 0)
	err := db.Conn.Where("outer_order_no = ? and status = ?", outOrderNo, status).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//根据外部订单号查询订单
func FcOrderHotFindByOutNo(outOrderNo string) ([]*entity.FcOrderHot, error) {
	results := make([]*entity.FcOrderHot, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).Desc("create_at").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcOrderHotFindTxId(txId string) (*entity.FcOrderHot, error) {
	order := &entity.FcOrderHot{}
	has, err := db.Conn.Where("tx_id = ?", txId).Get(order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return order, nil
}

//根据状态查询订单
func FcOrderHotFindListByOutNo(outOrderNo string) ([]*entity.FcOrderHot, error) {
	result := make([]*entity.FcOrderHot, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).Find(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//根据状态查询订单
func FcOrderHotFindByOrderId(orderId string) (*entity.FcOrderHot, error) {
	result := new(entity.FcOrderHot)
	has, err := db.Conn.Where("order_no = ?", orderId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcOrderHotInsert(order *entity.FcOrderHot) error {
	_, err := db.Conn.InsertOne(order)
	//
	//applyOrder, _ := FcTransfersApplyByOutOrderNo(order.OuterOrderNo)
	//tac, _ := GetApplyAddressByApplyCoinId(int64(applyOrder.Id), "to")
	//orderTx := &entity.FcOrderTxs{
	//	SeqNo:        util.GetSeqNo(),
	//	TxId:         order.TxId,
	//	OuterOrderNo: order.OuterOrderNo,
	//	InnerOrderNo: order.OrderNo,
	//	Mch:          applyOrder.Applicant,
	//	Chain:        applyOrder.CoinName,
	//	CoinCode:     applyOrder.Eoskey,
	//	Contract:     applyOrder.Eostoken,
	//	FromAddress:  order.FromAddress,
	//	ToAddress:    order.ToAddress,
	//	Amount:       tac.ToAmount,
	//	Sort:         0,
	//	CreateTime:   time.Now(),
	//	UpdateTime:   time.Now(),
	//}
	//if order.Status == int(status.BroadcastStatus) {
	//	orderTx.Status = entity.OtxBroadcastSuccess
	//} else {
	//	orderTx.Status = entity.OtxBroadcastFailure
	//}
	//if errInsertTx := InsertOrderTxs(orderTx); err != nil {
	//	log.Infof("插入orderhot 同时插入订单交易失败 %v", errInsertTx)
	//} else {
	//	push := &entity.FcOrderTxsPush{
	//		OrderTxsId: orderTx.Id,
	//
	//	}
	//}

	return err
}

func FcOrderHotUpdateFee(outOrderNo string, fee string) error {
	_, err := db.Conn.
		Table("fc_order_hot").
		Where("outer_order_no = ? ", outOrderNo).
		Update(map[string]interface{}{"fee": fee})
	if err != nil {
		return err
	}
	return nil
}

//根据内部订单号修改状态
func FcOrderHotUpdateState(orderNo string, status int) error {
	var (
		affected int64 = 0
	)
	res, err := db.Conn.Exec("update fc_order_hot set status = ? where order_no = ? ", status, orderNo)
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


func FcOrderHotUpdateTxIdAndStatus(id int, txId string) error {
	var (
		affected int64 = 0
	)
	res, err := db.Conn.Exec("update fc_order_hot set status = ?,tx_id = ? where id = ? LIMIT 1", status.BroadcastStatus, txId, id)
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