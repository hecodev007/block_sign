package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type UcaRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewUcaRegisterService() service.RegisterService {
	return &UcaRegisterService{
		CoinName: "uca",
	}
}

func (b *UcaRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/uca/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
