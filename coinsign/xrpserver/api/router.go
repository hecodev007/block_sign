package api

import (
	"github.com/gin-gonic/gin"
	"xrpserver/api/controller"
)

func InitRouter(coinname string) *gin.Engine {
	r := gin.New()
	new(controller.Controller).Router(r, coinname)
	//r.Use(gin.Logger())
	//r.Use(gin.Recovery())
	return r
}