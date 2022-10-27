package api

import (
	"github.com/group-coldwallet/dogeserver/service"
	"github.com/group-coldwallet/dogeserver/service/chain"
)

var (
	ChainService map[string]service.ChainService
)

func InitApiService() {
	ChainService = make(map[string]service.ChainService)
	ChainService["doge"] = chain.NewDogeService()
}
