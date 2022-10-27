package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

func HttpRespOnlyOK(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Msg:  GetMsg(SUCCESS),
		Data: nil,
	})
}

//httpCode http状态码
//errCode  业务执行码
//data 返回数据
func HttpRespOK(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Msg:  GetMsg(SUCCESS),
		Data: data,
	})
}

func HttpRespErrorOnly(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code: ERROR,
		Msg:  GetMsg(ERROR),
		Data: nil,
	})
}

//httpCode http状态码
//errCode  业务执行码
//msg  返回消息
//data 返回数据
func HttpRespError(ctx *gin.Context, errMsg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: ERROR,
		Msg:  errMsg,
		Data: data,
	})
}

func HttpRespCommon(ctx *gin.Context, code int, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
