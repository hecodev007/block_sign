package stacks

import (
	dao "github.com/group-coldwallet/chaincore2/dao/daostacks"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块
	dao.DeleteBlockInfo(height)

	// 块交易
	dao.DeleteBlockTX(height)
}
