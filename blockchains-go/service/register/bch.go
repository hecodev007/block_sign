package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type BchRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewBchRegisterService() service.RegisterService {
	return &BchRegisterService{
		CoinName: "bch",
	}
}

func (b *BchRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/bch/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
