package apis

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/bsc-sign/conf"
	v1 "github.com/bsc-sign/routers/apis/v1"
)

type Apis interface {
	CreateAddress(c *gin.Context)
	Sign(c *gin.Context)
	Transfer(c *gin.Context)
	TransferCollect(c *gin.Context)
	GetBalance(c *gin.Context)
	ValidAddress(c *gin.Context)
	DelKey(c *gin.Context)
}

func CreateApis(ctx context.Context) Apis {
	var apis Apis
	switch conf.Config.Version {
	case "v1":
		apis = v1.NewBaseApi(ctx)
	case "v2":
		// apis = v2.NewBaseApi()
	default:
		// 默认使用v1版本
		apis = v1.NewBaseApi(ctx)
	}
	return apis
}
