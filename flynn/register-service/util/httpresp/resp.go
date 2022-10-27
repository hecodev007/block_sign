package httpresp

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Response struct {
	Code    HttpCode    `json:"code"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func HttpRespOK(ctx *gin.Context, msg string, message string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Msg:     GetMsg(SUCCESS),
		Message: message,
		Data:    data,
	})
}

func HttpRespError(ctx *gin.Context, code int, message string) {
	log.Error(message)
	ctx.JSON(http.StatusOK, Response{
		Code:    HttpCode(code),
		Msg:     GetMsg(code),
		Message: message,
	})
}
