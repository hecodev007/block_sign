package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/heco-sign/conf"
	"github.com/group-coldwallet/heco-sign/model"
	"github.com/group-coldwallet/heco-sign/services"
	v1 "github.com/group-coldwallet/heco-sign/services/v1"
	"github.com/group-coldwallet/heco-sign/util"
	log "github.com/sirupsen/logrus"
	"strings"
)

type BaseApi struct {
	Srv services.IService
}

func NewBaseApi() *BaseApi {
	ba := new(BaseApi)
	ba.Srv = v1.GetIService()
	return ba
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
func (ba *BaseApi) CreateAddress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("content-type", "application/json")
	var (
		req  model.ReqCreateAddressParamsV2
		resp *model.RespCreateAddressParams
		err  error
	)

	//解析json数据
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
	//开启 参数验证
	if conf.Config.IsStartValid {
		err = ba.validParams(req)
		if err != nil {
			respFailDataReturn(c, fmt.Sprintf("验证参数签名错误，Err=[%v]", err))
			return
		}
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

func respFailDataReturn(c *gin.Context, message string) {
	c.JSON(200, gin.H{
		"code":    1,
		"message": message,
	})
	log.Error(message)
}

func (ba *BaseApi) validParams(req interface{}) error {
	reqData, ok := req.(map[string]interface{})
	if !ok {
		return fmt.Errorf("req data is not map[string]interface{}: %v", req)
	}
	var (
		sign                       string
		currentTime                string
		from, to, amount, contract string
	)
	// 1. 获取签名
	sign, ok = reqData["sign"].(string)
	if !ok {
		return fmt.Errorf("req data sign is null: %v", reqData["sign"])
	}
	currentTime, ok = reqData["current_time"].(string)
	if !ok {
		return fmt.Errorf("req data current time is null: %v", reqData["current_time"])
	}
	from, ok = reqData["from_address"].(string)
	if !ok {
		return fmt.Errorf("req data from address is null: %v", reqData["from_address"])
	}
	to, ok = reqData["to_address"].(string)
	if !ok {
		return fmt.Errorf("req data to address is null: %v", reqData["to_address"])
	}
	amount, ok = reqData["amount"].(string)
	if !ok {
		return fmt.Errorf("req data sign is null: %v", reqData["amount"])
	}
	contract, ok = reqData["contract_address"].(string)
	if !ok {
		contract = ""
	}
	newSig, err := ba.createSign(from, to, amount, contract, currentTime)
	if err != nil {
		return err
	}
	if sign != newSig {
		return fmt.Errorf("签名不正确：%s", newSig)
	}
	return nil
}

func (ba *BaseApi) createSign(from, to, amount, contract, currentTime string) (string, error) {
	secret := fmt.Sprintf("%sHOO_WALLET_TRANSFER_SERVICE%s", currentTime, currentTime)
	params := fmt.Sprintf("from=%s&to=%s&amount=%s&contract=%s&time=%s", from, to, amount, contract, currentTime)
	key := sha256.Sum256([]byte(secret))
	sig, err := util.AesBase64Crypt([]byte(params), key[:], true)
	if err != nil {
		return "", fmt.Errorf("加密req data error:%v", err)
	}
	crypt := sha256.Sum256(sig)
	return "0x" + hex.EncodeToString(crypt[:]), err
}
