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

func HttpRespOkOnly(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Status:  SUCCESS,
		Msg:     GetMsg(SUCCESS),
		Success: true,
		Data:    nil,
	})
}

func HttpRespOK(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Status:  SUCCESS,
		Msg:     msg,
		Success: true,
		Data:    data,
	})
}

func HttpRespErrorOnly(ctx *gin.Context) {
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
func HttpRespError(ctx *gin.Context, code HttpCode, errMsg string, data interface{}) {
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

type ResponseCode struct {
	Code    HttpCode    `json:"code"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func HttpRespCodeOkOnly(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    SUCCESS,
		Msg:     GetMsg(SUCCESS),
		Message: GetMsg(SUCCESS),
		Data:    nil,
	})
}

func HttpRespCodeOkOnlyWithMsg(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    SUCCESS,
		Msg:     msg,
		Message: msg,
		Data:    nil,
	})
}

func HttpRespCodeOK(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    SUCCESS,
		Msg:     msg,
		Message: GetMsg(SUCCESS),
		Data:    data,
	})
}

func HttpRespCodeErrOnly(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    FAIL,
		Msg:     GetMsg(FAIL),
		Message: GetMsg(FAIL),
		Data:    nil,
	})
}

func HttpRespErrWithMsg(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    FAIL,
		Msg:     msg,
		Message: msg,
		Data:    nil,
	})
}

//httpCode http状态码
//errCode  业务执行码
//msg  返回消息
//data 返回数据
func HttpRespCodeError(ctx *gin.Context, code HttpCode, errMsg string, data interface{}) {
	if code == SUCCESS {
		code = FAIL
	}
	ctx.JSON(http.StatusOK, ResponseCode{
		Code:    code,
		Msg:     errMsg,
		Data:    data,
		Message: errMsg,
	})
}
