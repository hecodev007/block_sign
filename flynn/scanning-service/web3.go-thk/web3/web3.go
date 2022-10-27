package web3

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/account"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/net"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/thk"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/utils"
)

type Web3 struct {
	DefaultAddress string
	Provider       providers.ProviderInterface
	Thk            *thk.Thk
	Net            *net.Net
	Personal       *account.Personal
	Utils          *utils.Utils
}

func NewWeb3(provider providers.ProviderInterface) *Web3 {
	web3 := new(Web3)
	web3.Provider = provider
	web3.Thk = thk.NewThk(provider)
	web3.Net = net.NewNet(provider)
	web3.Personal = account.NewPersonal(provider)
	web3.Utils = utils.NewUtils(provider)
	return web3
}
