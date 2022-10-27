package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//查询block信息列表
func FcTxClearFindList(coinType, txid string) ([]*entity.FcTxClear, int64, error) {
	results := make([]*entity.FcTxClear, 0)
	count, err := db.Conn.Where("coin_type= ? and tx_id = ?", coinType, txid).Desc("id").FindAndCount(&results)
	if err != nil {
		return nil, 0, err
	}
	return results, count, nil
}
