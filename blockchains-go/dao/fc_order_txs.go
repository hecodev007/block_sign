package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"time"
	"xorm.io/builder"
)

func InsertOrderTxs(order *entity.FcOrderTxs) error {
	_, err := db.Conn.InsertOne(order)
	return err
}

//
//func GetOrderTxBySeqNoLimit1() (*entity.FcOrderTxs, error) {
//	order := &entity.FcOrderTxs{}
//	has, err := db.Conn.Where("id = 186").Get(order)
//	if err != nil {
//		return nil, err
//	}
//	if !has {
//		return nil, errors.New("Not Fount!")
//	}`
//	return order, nil
//}

func UpdateOrderTxsStatus(seqNo string, newStatus entity.OrderTxStatus) error {
	_, err := db.Conn.
		Table("fc_order_txs").
		Where("seq_no = ? ", seqNo).
		Update(map[string]interface{}{"status": newStatus, "update_time": time.Now()})
	if err != nil {
		return err
	}
	return nil
}

func UpdateOrderTxToCancelAndUnlockFreeze(tx *entity.FcOrderTxs) error {
	session := db.Conn.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	coinType := tx.Chain
	if tx.CoinCode != "" {
		coinType = tx.CoinCode
	}
	session.Exec("update fc_order_txs set freeze_unlock = 1,status = ?,update_time = ? where id = ? limit 1", entity.OtxCanceled, time.Now(), tx.Id)
	session.Exec("update fc_address_amount set forzen_amount = forzen_amount - ? where coin_type = ? and address = ? limit 1",
		tx.Amount, coinType, tx.FromAddress)
	return session.Commit()
}

func GetOrderTxBySeqNo(seqNo string) (*entity.FcOrderTxs, error) {
	tx := &entity.FcOrderTxs{}
	has, err := db.Conn.Where("seq_no = ?", seqNo).Get(tx)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return tx, nil
}

//
//func GetOrderTxBySeqNo(seqNo string) (*entity.FcOrderTxs, error) {
//	result := &entity.FcOrderTxs{}
//	exist, err := db.Conn.Table("fc_order_txs").
//		Where("seq_no = ?", seqNo).
//		Get(result)
//	if err != nil {
//		return nil, err
//	}
//	if !exist {
//		return nil, nil
//	}
//	return result, nil
//}

func FindOrderTxsByOuterOrderNo(outOrderNo string) ([]entity.FcOrderTxs, error) {
	results := make([]entity.FcOrderTxs, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).Desc("sort").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindOrderTxsCompletedByOuterOrderNo(outOrderNo string) ([]entity.FcOrderTxs, error) {
	results := make([]entity.FcOrderTxs, 0)
	err := db.Conn.Where("outer_order_no = ?", outOrderNo).And(builder.In("status", []int{4, 12})).Desc("sort").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindOrderTxsValid(outOrderNo string) ([]entity.FcOrderTxs, error) {
	results := make([]entity.FcOrderTxs, 0)
	err := db.Conn.Where("outer_order_no = ? AND status != ?", outOrderNo, entity.OtxTrxOnChainFailure).Desc("sort").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindOrderTxsChainFailure(outOrderNo string) ([]entity.FcOrderTxs, error) {
	results := make([]entity.FcOrderTxs, 0)
	err := db.Conn.Where("outer_order_no = ? AND status = ?", outOrderNo, entity.OtxTrxOnChainFailure).Desc("sort").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindOrderTxsNotChainFailure(outOrderNo string) ([]entity.FcOrderTxs, error) {
	results := make([]entity.FcOrderTxs, 0)
	err := db.Conn.Where("outer_order_no = ? AND status != ?", outOrderNo, entity.OtxTrxOnChainFailure).Desc("sort").Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
