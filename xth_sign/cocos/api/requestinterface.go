package api

import "github.com/gin-gonic/gin"

type HttpRequest interface {
	CreateAccount(c *gin.Context)
	Transfer(c *gin.Context)
	GetBalance(c *gin.Context)
	VaildAddress(c *gin.Context)
}
