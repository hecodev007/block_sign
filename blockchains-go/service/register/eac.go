package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type EacRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewEacRegisterService() service.RegisterService {
	return &EacRegisterService{
		CoinName: "eac",
	}
}

func (b *EacRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/v1/eac/importaddrs"
	log.Infof("发送地址：%s", url)
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
