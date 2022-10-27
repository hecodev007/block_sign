package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/celo-sign/conf"
	v1 "github.com/group-coldwallet/celo-sign/routers/apis/v1"
	v2 "github.com/group-coldwallet/celo-sign/routers/apis/v2"
)

type Apis interface {
	CreateAddress(c *gin.Context)
	Sign(c *gin.Context)
	Transfer(c *gin.Context)
	GetBalance(c *gin.Context)
}

func CreateApis() Apis {
	var apis Apis
	switch conf.Config.Version {
	case "v1":
		apis = v1.NewBaseApi()
	case "v2":
		apis = v2.NewBaseApi()
	default:
		//默认使用v1版本
		apis = v1.NewBaseApi()
	}
	return apis
}
