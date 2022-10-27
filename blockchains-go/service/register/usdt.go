package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type UsdtRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewUsdtRegisterService() service.RegisterService {
	return &UsdtRegisterService{
		CoinName: "usdt",
	}
}

func (b *UsdtRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	//btc服务也需要注册
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/btc/importaddrs"
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	util.PostJson(url, mapData)

	url = conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/usdt/importaddrs"
	return util.PostJson(url, mapData)
}
