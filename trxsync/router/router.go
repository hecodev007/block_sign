package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/trxsync/api"
	"github.com/group-coldwallet/trxsync/services"
)

func InitRouter(bs *services.BaseService) (*gin.Engine, *api.MController, error) {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	if bs.Cfg.Mode == "prod" {
		gin.SetMode("release")
	}
	r.Use(gin.Logger())
	controller, err := api.NewMController(bs)
	if err != nil {
		return nil, nil, err
	}
	rGroup := r.Group(fmt.Sprintf("/%s", bs.Cfg.Sync.Name))
	{
		rGroup.POST("/rpc", controller.RpcPost)
		rGroup.POST("/insert", controller.InsertWatchAddress)
		rGroup.POST("/remove", controller.RemoveWatchAddress)
		rGroup.POST("/update", controller.UpdateWatchAddress)
		rGroup.POST("/repush", controller.RepushTx)
		rGroup.POST("/insertcontract", controller.InsertWatchContract)
		rGroup.POST("/removecontract", controller.RemoveWatchContract)
	}

	r.GET("/info", controller.Info)

	return r, controller, nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length,access-control-allow-origin, access-control-allow-headers")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
