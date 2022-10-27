package api

import (
	"dotsign/api/controller"
	"github.com/gin-gonic/gin"
)

func InitRouter(name, runmode string) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	if runmode != "prod" {
		r.Use(gin.Logger())
	}
	gin.SetMode(runmode)

	//new(controller.ZcashController).Router(r)
	new(controller.Controller).Router(r)
	return r, nil
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
