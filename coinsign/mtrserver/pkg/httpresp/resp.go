package httpresp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code HttpCode    `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func HttpRespOk(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Msg:  GetMsg(SUCCESS),
		Data: nil,
	})
}

func HttpRespOkByMsg(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Msg:  msg,
		Data: data,
	})
}

func HttpRespError(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code: FAIL,
		Msg:  GetMsg(FAIL),
		Data: nil,
	})
}

//httpCode http状态码
//errCode  业务执行码
//msg  返回消息
//data 返回数据
func HttpRespErrorByMsg(ctx *gin.Context, code HttpCode, errMsg string, data interface{}) {
	if code == SUCCESS {
		code = FAIL
	}
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  errMsg,
		Data: data,
	})
}
