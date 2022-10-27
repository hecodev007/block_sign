package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Gin struct {
	C *gin.Context
}

func (g *Gin) Response(httpCode, errCode int, data interface{}) {
	g.C.JSON(http.StatusOK, gin.H{
		"code":    errCode,
		"message": "",
		"data":    data,
	})
	return
}

func (g *Gin) ResponseMsg(httpCode, errCode int, Msg string) {
	g.C.JSON(http.StatusOK, gin.H{
		"code":    errCode,
		"message": Msg,
		"data":    nil,
	})
	return
}
