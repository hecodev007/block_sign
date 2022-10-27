package httpresp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    HttpCode    `json:"code"`
	Status  HttpCode    `json:"status"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

func HttpRespOk(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Status:  SUCCESS,
		Msg:     GetMsg(SUCCESS),
		Success: true,
		Data:    nil,
	})
}

func HttpRespOkByMsg(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Status:  SUCCESS,
		Msg:     msg,
		Success: true,
		Data:    data,
	})
}

func HttpRespError(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code:    FAIL,
		Status:  FAIL,
		Msg:     GetMsg(FAIL),
		Success: false,
		Data:    nil,
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
		Code:    code,
		Status:  code,
		Msg:     errMsg,
		Success: false,
		Data:    data,
	})
}
