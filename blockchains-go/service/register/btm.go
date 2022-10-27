package register

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
)

type BtmRegisterService struct {
	CoinName string `json:"coinName"`
}

func NewBtmRegisterService() service.RegisterService {
	return &BtmRegisterService{
		CoinName: "btm",
	}
}

func (b *BtmRegisterService) RegisterToNode(addrs []string) ([]byte, error) {
	url := conf.Cfg.CoinServers[b.CoinName].Url + "/api/v1/btm/importaddrs"
	log.Infof("发送地址：%s", url)
	mapData := make(map[string][]string, 0)
	mapData["addrs"] = addrs
	return util.PostJson(url, mapData)
}
