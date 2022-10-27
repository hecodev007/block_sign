package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

func AmAddBatchAddresses(addrs []entity.Addresses) (int64, error) {
	return db.ConnAddrMgr.Insert(addrs)
}
