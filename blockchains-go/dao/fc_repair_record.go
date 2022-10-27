package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FcGetRepairRecordByTxId(txId string) (*entity.FcRepairRecord, error) {
	result := new(entity.FcRepairRecord)
	has, err := db.Conn.Where("tx_id = ?", txId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return result, nil
}
