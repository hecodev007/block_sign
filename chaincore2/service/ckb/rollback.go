package ckb

import (
	dao "github.com/group-coldwallet/chaincore2/dao/daockb"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块交易
	dao.DeleteFromBlockTX(height)

	// 块vin
	dao.DeleteFromBlockTXVin(height)

	// 块vou
	dao.DeleteFromBlockTXVout(height)

	// 块
	dao.DeleteBlockInfo(height)
}
