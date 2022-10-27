package middleware

import (
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/mch"
	"github.com/group-coldwallet/blockchains-go/service/security"
)

var (
	transferSecurityService service.TransferSecurityService
	mchService              service.MchService
)

func init() {
	transferSecurityService = security.NewSecurityService()
	mchService = mch.NewMchBaseService()
}
