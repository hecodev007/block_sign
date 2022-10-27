package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FcTxTransactionsByTxId(txId string) ([]entity.FcTxTransactionNew, error) {
	results := make([]entity.FcTxTransactionNew, 0)
	err := db.Conn.Where("tx_id = ?", txId).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
