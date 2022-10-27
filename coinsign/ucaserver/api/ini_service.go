package api

import (
	"github.com/group-coldwallet/ucaserver/service"
	"github.com/group-coldwallet/ucaserver/service/chain"
)

var (
	ChainService map[string]service.ChainService
)

func InitApiService() {
	ChainService = make(map[string]service.ChainService)
	ChainService["uca"] = chain.NewUcaService()
}
