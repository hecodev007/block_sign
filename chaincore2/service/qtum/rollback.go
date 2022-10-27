package qtum

import (
	dao "github.com/group-coldwallet/chaincore2/dao/daoqtum"
	"github.com/group-coldwallet/common/log"
)

// 回滚到指定高度
func Rollback(height int64) {
	log.Debug("Rollback.........", height)

	// 块交易
	dao.DeleteBlockTX(height)

	// 块vin
	dao.DeleteBlockTXVin(height)

	// 块vou
	dao.DeleteBlockTXVout(height)

	// contract
	dao.DeleteContractTX(height)

	// 块
	dao.DeleteBlockInfo(height)
}
