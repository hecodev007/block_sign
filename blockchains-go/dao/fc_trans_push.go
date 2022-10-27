package dao

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"strings"
)

//查询1000个可用的utxo
func FcTransPushFindVaildUtxo(appId int, coinName string) ([]entity.FcTransPush, error) {
	results := make([]entity.FcTransPush, 0)
	err := db.Conn.Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).Desc("amount").Limit(1000).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcTransPushFindVaildUtxo2(appId int, coinName string, limit int, ascOrDescByAmount string) ([]entity.FcTransPush, error) {
	results := make([]entity.FcTransPush, 0)
	var err error
	if ascOrDescByAmount == "desc" {
		err = db.Conn.Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).Desc("amount").Limit(limit).Find(&results)
	} else if ascOrDescByAmount == "asc" {
		err = db.Conn.Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).Asc("amount").Limit(limit).Find(&results)
	} else {
		err = db.Conn.Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).Limit(limit).Find(&results)
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

//冻结指定utxo
//func FcTransPushFreezeUtxo(ids []int64) error {
//	_, err := db.Conn.Exec("update fc_trans_push set is_spent = 2 where id in (?)", ids)
//	return err
//}

//冻结指定utxo
func FcTransPushFreezeUtxo(txid string, index int, address string) error {
	_, err := db.Conn.Exec("update fc_trans_push set is_spent = 2 where transaction_id = ? and trx_n = ? and address = ?", txid, index, address)
	return err
}

//解动指定utxo
func FcTransPushUnFreezeUtxo(coin, txid string, index int, address string) error {
	if strings.ToLower(coin) == "btm" {
		// params.address is mux_id
		_, err := db.Conn.Exec("update fc_trans_push set is_spent = 0 where  transaction_id = ? and trx_n = ? and mux_id = ? and is_spent = 2", txid, index, address)
		return err
	}
	_, err := db.Conn.Exec("update fc_trans_push set is_spent = 0 where  transaction_id = ? and trx_n = ? and address = ? and is_spent = 2", txid, index, address)
	return err
}

//params appId: 	商户ID
//params coinName: 	币种名
//params limit: 	需要的记录数
//params ascOrDescByAmount: 升序或者降序查询，升序从小到大查找出账UTXO，降序 从大到小查询
func FcTransPushAddressValidUTXO(appId int64, coinName string, limit int, ascOrDescByAmount string) ([]*entity.FcTransPush, error) {
	results := make([]*entity.FcTransPush, 0)
	var err error
	if ascOrDescByAmount == "asc" {
		err = db.Conn.Table("fc_trans_push").
			Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).
			Asc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "desc" {
		err = db.Conn.Table("fc_trans_push").
			Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).
			Desc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "" {
		err = db.Conn.Table("fc_trans_push").
			Where("app_id  = ? and coin_type = ? and is_in = 1 and is_spent = 0 and confirmations > 0", appId, coinName).
			Limit(limit).
			Find(&results)
	} else {
		err = fmt.Errorf("error ascOrDesc type: %s", ascOrDescByAmount)
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}
