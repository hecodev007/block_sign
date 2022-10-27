package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func FindTxBreakByChain(chain string) ([]entity.FcTxBreak, error) {
	results := make([]entity.FcTxBreak, 0)
	err := db.Conn.Where("chain = ?", chain).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func DeleteTxBreakById(txId string) error {
	result := new(entity.FcTxBreak)
	_, err := db.Conn.Where("tx_id = ?", txId).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
