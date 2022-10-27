package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type ZecRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewZecRegisterService() service.RegisterService {
	return &ZecRegisterService{
		CoinName: "zec",
	}
}

func (b *ZecRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/zec/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
