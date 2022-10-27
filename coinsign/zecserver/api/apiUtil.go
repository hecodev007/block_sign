package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/zecserver/api/e"
	"github.com/group-coldwallet/zecserver/util"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
	Hash string      `json:"hash,omitempty"`
}

//httpCode http状态码
//errCode  业务执行码
//data 返回数据

func HttpResponse(ctx *gin.Context, httpCode, errCode int, data interface{}) {
	hash := ""
	if data != nil {
		byteData, err := json.Marshal(data)
		if err != nil {
			logrus.Error(err)
		}
		hash = util.Md5HashString(byteData)
	}
	ctx.JSON(httpCode, Response{
		Code: errCode,
		Msg:  e.GetMsg(errCode),
		Data: data,
		Hash: hash,
	})
}

//httpCode http状态码
//errCode  业务执行码
//msg  返回消息
//data 返回数据
func HttpResponseByMsg(ctx *gin.Context, httpCode, errCode int, msg string, data interface{}) {
	hash := ""
	if data != nil {
		byteData, err := json.Marshal(data)
		if err != nil {
			logrus.Error(err)
		}
		hash = util.Md5HashString(byteData)
	}
	ctx.JSON(httpCode, Response{
		Code: errCode,
		Msg:  msg,
		Data: data,
		Hash: hash,
	})
}
