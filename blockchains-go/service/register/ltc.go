package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type LtcRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewLtcRegisterService() service.RegisterService {
	return &LtcRegisterService{
		CoinName: "ltc",
	}
}

func (b *LtcRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/ltc/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
