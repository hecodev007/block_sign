package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type BchaRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewBchaRegisterService() service.RegisterService {
	return &BchaRegisterService{
		CoinName: "bcha",
	}
}

func (b *BchaRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/bcha/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
