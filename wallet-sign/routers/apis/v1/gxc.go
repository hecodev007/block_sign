package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/wallet-sign/model"
)

type GxcApi struct {
	*BaseApi
}

func (ga *GxcApi) GetBalance(c *gin.Context) {
	panic("implement me")
}

func NewGxcApi() *GxcApi {
	ga := new(GxcApi)
	ga.BaseApi = NewBaseApi()
	return ga
}

func (ga *GxcApi) CreateAddress(c *gin.Context) {
	ga.BaseApi.CreateAddress(c)
}

func (ga *GxcApi) Sign(c *gin.Context) {
	//ga.BaseApi.Sign(c)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqSignParams
		resp model.RespGxcSignParams
		err  error
	)
	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse sign post data error")
		return
	}
	if req.OrderId == "" {
		respFailDataReturn(c, "Order id is null")
		return
	}

	if req.MchId == "" {
		respFailDataReturn(c, "Mch id is null")
		return
	}
	if req.Data == nil {
		respFailDataReturn(c, "data is null")
		return
	}
	data, err := ga.Srv.SignService(&req)
	if err != nil {
		respFailDataReturn(c, err.Error())
		return
	}
	resp.ReqBaseParams = req.ReqBaseParams
	resp.Hex = data.(string)
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    resp,
	})
}

func (ga *GxcApi) Transfer(c *gin.Context) {
	//ga.BaseApi.Transfer(c)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req interface{}
		err error
	)
	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, fmt.Sprintf("Parse transfer post data error,Err=[%v]", err))
		return
	}
	//调用transfer服务
	data, err := ga.Srv.TransferService(req)
	if err != nil {
		respFailDataReturn(c, fmt.Sprintf("transfer error ,Err=[%v]", err))
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}
