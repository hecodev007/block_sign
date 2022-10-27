package eos

import (
	dao "github.com/group-coldwallet/chaincore2/dao/daoeos"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块交易
	dao.DeleteFromBlockTX(height)

	// 块
	dao.DeleteBlockInfo(height)
}
