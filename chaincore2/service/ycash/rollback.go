package ycash

import (
	"github.com/group-coldwallet/chaincore2/dao/daoycash"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块
	daoycash.DeleteBlockInfo(height)

	// 块交易
	daoycash.DeleteBlockTX(height)

	// 块vin
	daoycash.DeleteBlockTXVin(height)

	// 块vou
	daoycash.DeleteBlockTXVout(height)
}
