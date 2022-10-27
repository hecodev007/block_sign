package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//查询block信息列表
func FcTxTransactionFindList(mchId, txType int, coinType, dateStart, dateEnd string) ([]*entity.FcTxTransaction, int64, error) {
	results := make([]*entity.FcTxTransaction, 0)
	count, err := db.Conn.Where("mch_id = ? and tx_type = ? and coin_type = ? and contrast_time between ? and ?",
		mchId, txType, coinType, dateStart, dateEnd).Desc("contrast_time").FindAndCount(&results)
	if err != nil {
		return nil, 0, err
	}
	return results, count, nil
}


