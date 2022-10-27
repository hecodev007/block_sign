package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type SatRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewSatRegisterService() service.RegisterService {
	return &SatRegisterService{
		CoinName: "satcoin",
	}
}

func (b *SatRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/satcoin/importaddrs"
	log.Infof("发送地址：%s", url)
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
