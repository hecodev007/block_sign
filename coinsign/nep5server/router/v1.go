package router

import (
	"github.com/gin-gonic/gin"
	v12 "github.com/group-coldwallet/nep5server/api/v1"
	"github.com/group-coldwallet/nep5server/conf"
)

func InitV1(r *gin.Engine, c *conf.Config) {
	nep5Api := v12.NewNep5API(c)

	v1 := r.Group("/v1")
	v1.POST("/createaddr", nep5Api.CreateAddr)
	v1.POST("/sign", nep5Api.Transf)
}
