package routers

import (
	"github.com/gin-gonic/gin"
	"ytaDataServer/api"
	"ytaDataServer/common"
	"ytaDataServer/services"
)

func InitRouter(name, runmode string, w services.WatchControl, p common.Processor) *gin.Engine {
	r := gin.New()
	r.Use(corsMiddleware())

	if runmode == "prod" {
		gin.SetMode("release")
	}

	api.NewMController(p, w).Router(r, name)
	//r.Use(gin.Logger())
	//r.Use(gin.Recovery())
	return r
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
