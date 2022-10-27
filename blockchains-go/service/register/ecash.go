package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type XecRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewXecRegisterService() service.RegisterService {
	return &XecRegisterService{
		CoinName: "xec",
	}
}

func (b *XecRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/xec/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
