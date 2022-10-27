package hc

import (
	dao "github.com/group-coldwallet/chaincore2/dao/daohc"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func RollbackFromHeight(height int64) {
	log.Debug("RollbackFromHeight.........", height)

	// 块交易
	dao.DeleteFromBlockTX(height)

	// 块vin
	dao.DeleteFromBlockTXVin(height)

	// 块vou
	dao.DeleteFromBlockTXVout(height)

	// 块
	dao.DeleteFromBlockInfo(height)
}

// 回滚指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块交易
	dao.DeleteBlockTX(height)

	// 块vin
	dao.DeleteBlockTXVin(height)

	// 块vou
	dao.DeleteBlockTXVout(height)

	// 块
	dao.DeleteBlockInfo(height)
}
