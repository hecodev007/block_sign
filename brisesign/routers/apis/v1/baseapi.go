package v1

import (
	"brisesign/conf"
	"brisesign/model"
	"brisesign/services"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"strings"
)

type BaseApi struct {
	Srv services.IService
}

func NewBaseApi(ctx context.Context) *BaseApi {
	ba := new(BaseApi)
	ba.Srv = services.GetIService(ctx)
	return ba
}
func (ba *BaseApi) CreateAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqCreateAddressParamsV2
		resp *model.RespCreateAddressParams
		err  error
	)

	// 解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse create address post data error")
		return
	}
	if req.Count > 50000 {
		respFailDataReturn(c, fmt.Sprintf("Create address nums must be less than 50000,Num=%d", req.Count))
		return
	}
	if req.Mch == "" {
		respFailDataReturn(c, "Mch id is null")
		return
	}
	if strings.ToLower(req.CoinCode) != strings.ToLower(conf.Config.CoinType) {
		respFailDataReturn(c, fmt.Sprintf("Coin name is not %s", strings.ToLower(conf.Config.CoinType)))
		return
	}

	// 调用service
	resp, err = ba.Srv.CreateAddressService(&req)
	//var arr = []util.AddrInfo{ //用于水龙头测试账号的密钥插入
	//	{
	//		Address: "0xc2bba9ae55bc6a7d13797e3fd393c406702aaf60",
	//		PrivKey: "837b1cb134e5c6d4d08d56e1fb5000b0de8cc6bb62b8445c60f69c9e425ff1ed",
	//	},
	//}
	//util.CreateAddrCsv("./file", "hoo", "test", "op", arr)
	if err != nil {
		respFailDataReturn(c, "Create address error,Err="+err.Error())
		return
	}

	// 成功发送
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
	// 解析json数据
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
	// 成功发送
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
	// 成功发送

	var (
		req interface{}
		err error
	)
	// 解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, fmt.Sprintf("Parse transfer post data error,Err=[%v]", err))
		return
	}
	// 调用transfer服务
	data, err := ba.Srv.TransferService(req)

	if err != nil {
		respFailDataReturn(c, fmt.Sprintf("transfer error ,Err=[%v]", err))
		return
	}
	// 成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    string(data),
	})
}

func (ba *BaseApi) TransferCollect(c *gin.Context) {
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
	data, err2 := ba.Srv.TransferCollectService(req)
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
	log.Println(message)
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
	// 解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse get balance post data error")
		return
	}
	// 判断必要的参数
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
	// 调用service
	resp, err = ba.Srv.GetBalance(&req)
	if err != nil {
		respFailDataReturn(c, "get balance error,Err="+err.Error())
		return
	}
	// 成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    resp,
	})
}

func (ba *BaseApi) ValidAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.ReqValidAddressParams
		err error
	)
	// 解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse valid address post data error")
		return
	}
	// 判断必要的参数
	if req.Address == "" {
		respFailDataReturn(c, "address is null")
		return
	}
	// 调用service
	err = ba.Srv.ValidAddress(req.Address)
	if err != nil {
		respFailDataReturn(c, err.Error())
		return
	}
	// 成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    true,
	})
}

func (ba *BaseApi) DelKey(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req model.ReqDelKeyParams
		err error
	)
	// 解析json数据
	if err = c.BindJSON(&req); err != nil {
		respFailDataReturn(c, "Parse post data error")
		return
	}

	// 判断必要的参数
	if req.OrderId == "" {
		respFailDataReturn(c, "orderId require")
		return
	}

	err = ba.Srv.DelKey(req.OrderId)
	if err != nil {
		respFailDataReturn(c, err.Error())
		return
	}

	// 成功发送
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    true,
	})
}
