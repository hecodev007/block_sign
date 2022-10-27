package api

import (
	"github.com/group-coldwallet/mtrserver/service"
	"github.com/group-coldwallet/mtrserver/service/chain"
)

var (
	ChainService map[string]service.ChainService
)

func InitApiService() {
	ChainService = make(map[string]service.ChainService)
	ChainService["mtr"] = chain.NewMtrService()
}
