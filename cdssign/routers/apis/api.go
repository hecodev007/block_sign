package apis

import (
	"github.com/eth-sign/conf"
	v1 "github.com/eth-sign/routers/apis/v1"
	"github.com/gin-gonic/gin"
)

type Apis interface {
	CreateAddress(c *gin.Context)
	Sign(c *gin.Context)
	Transfer(c *gin.Context)
	TransferWithNonce(c *gin.Context)
	GetBalance(c *gin.Context)
	ValidAddress(c *gin.Context)
}

func CreateApis() Apis {
	var apis Apis
	switch conf.Config.Version {
	case "v1":
		apis = v1.NewBaseApi()
	case "v2":
		// apis = v2.NewBaseApi()
	default:
		// 默认使用v1版本
		apis = v1.NewBaseApi()
	}
	return apis
}
