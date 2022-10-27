package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/services"
	v1 "github.com/group-coldwallet/wallet-sign/services/v1"
	log "github.com/sirupsen/logrus"
	"strings"
)

type BaseApi struct {
	Srv services.IService
}

func (ba *BaseApi) ValidAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.ReqValidAddressParams
		err error
	)
	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse valid address post data error")
		return
	}
	//判断必要的参数
	if req.Address == "" {
		respFailDataReturn(c, "address is null")
		return
	}
	//调用service
	err = ba.Srv.ValidAddress(req.Address)
	if err != nil {
		respFailDataReturn(c, err.Error())
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    true,
	})
}

func (ba *BaseApi) GetBalance(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqGetBalanceParams
		resp interface{}
		err  error
	)
	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse get balance post data error")
		return
	}
	//判断必要的参数
	if req.CoinName == "" {
		respFailDataReturn(c, "coin name  is null")
		return
	}
	if strings.ToLower(req.CoinName) != strings.ToLower(conf.Config.CoinType) {
		respFailDataReturn(c, fmt.Sprintf("Coin name is not %s", strings.ToLower(conf.Config.CoinType)))
		return
	}
	if req.Address == "" {
		respFailDataReturn(c, "address is null")
		return
	}
	//调用service
	resp, err = ba.Srv.GetBalance(&req)
	if err != nil {
		respFailDataReturn(c, "get balance error,Err="+err.Error())
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    resp,
	})
}

func NewBaseApi() *BaseApi {
	ba := new(BaseApi)
	ba.Srv = v1.GetIService()
	return ba
}
func (ba *BaseApi) CreateAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqCreateAddressParams
		resp *model.RespCreateAddressParams
		err  error
	)

	//解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse create address post data error")
		return
	}

	if req.Num > 50000 || req.Num <= 0 {
		respFailDataReturn(c, fmt.Sprintf("Create address nums is less than zero or more than 50000,Num=%d", req.Num))
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

	if strings.ToLower(req.CoinName) != strings.ToLower(conf.Config.CoinType) {
		respFailDataReturn(c, fmt.Sprintf("Coin name is not %s", strings.ToLower(conf.Config.CoinType)))
		return
	}

	//调用service
	resp, err = ba.Srv.CreateAddressService(&req)

	if err != nil {
		respFailDataReturn(c, "Create address error,Err="+err.Error())
		return
	}

	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    resp,
	})
}
func (ba *BaseApi) Sign(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqSignParams
		resp model.RespSignParams
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
	data, err2 := ba.Srv.SignService(&req)
	if err2 != nil {
		respFailDataReturn(c, fmt.Sprintf("sign error,Err=%v", err2))
		return
	}

	resp.ReqBaseParams = req.ReqBaseParams
	resp.Result = data.(string)
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    resp,
	})
}
func (ba *BaseApi) Transfer(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	//成功发送

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
	data, err2 := ba.Srv.TransferService(req)
	if err2 != nil {
		respFailDataReturn(c, fmt.Sprintf("transfer error ,Err=[%v]", err2))
		return
	}
	//成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func respFailDataReturn(c *gin.Context, message string) {
	c.JSON(200, gin.H{
		"code":    1,
		"message": message,
	})
	log.Errorf("发送失败 %s", message)
}
