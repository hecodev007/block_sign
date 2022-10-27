package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/bchserver/api"
	"github.com/group-coldwallet/bchserver/api/e"
	"github.com/group-coldwallet/bchserver/conf"
	"github.com/group-coldwallet/bchserver/model/bo"
	"github.com/group-coldwallet/bchserver/model/global"
	"github.com/group-coldwallet/bchserver/model/vo"
	"github.com/group-coldwallet/bchserver/service/bchservice"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

type API struct {
	bchService bchservice.BasicService
}

//实例化指定版本的service
func NewBchAPI() *API {
	return &API{
		bchService: new(bchservice.BchService),
	}
}

// @Summary 创建签名模板
// @Produce  json
// @param body body string true "sign tpl"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/create [post]
func (a *API) CreateTpl(c *gin.Context) {
	var (
		err          error
		fromAmount   int64
		toAmount     int64
		changeAmount int64
	)

	tpl := new(bo.BchTxTpl)
	txinput := new(bo.TxInput)
	err = c.BindJSON(txinput)
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}
	//校验参数
	if txinput.ChangeAddr == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty changeAddress", nil)
		return
	}
	if len(txinput.Txins) == 0 {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty txinputs", nil)
		return
	}
	if len(txinput.Txouts) == 0 {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty Txouts", nil)
		return
	}
	if txinput.Fee < conf.GlobalConf.BchCfg.MinFee {
		err = fmt.Errorf("Fee too low,min:%d ,now:%d", conf.GlobalConf.BchCfg.MinFee, txinput.Fee)
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}

	if txinput.Fee > conf.GlobalConf.BchCfg.MaxFee {
		err = fmt.Errorf("Fee too high,max:%d ,now:%d", conf.GlobalConf.BchCfg.MaxFee, txinput.Fee)
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}

	tpl.MchId = txinput.MchId

	for i, v := range txinput.Txins {
		if v.Address == "" {
			err = fmt.Errorf("Txins index:%d error,empty address", i)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		if v.Txid == "" {
			err = fmt.Errorf("Txins index:%d ,address:%s,error,empty txid", i, v.Address)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		if v.Amount < 0 {
			err = fmt.Errorf("Txins index:%d ,address:%s,error,error Amount", i, v.Address)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		if v.Vout < 0 {
			err = fmt.Errorf("Txins index:%d ,address:%s,error,error Vout", i, v.Address)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		fromAmount += v.Amount

		tpl.TxIns = append(tpl.TxIns, bo.BchTxInTpl{
			FromAddr:         v.Address,
			FromTxid:         v.Txid,
			FromAmount:       v.Amount,
			FromIndex:        uint32(v.Vout),
			FromRedeemScript: v.RedeemScript,
		})
	}

	for i, v := range txinput.Txouts {
		if v.ToAddress == "" {
			err = fmt.Errorf("Txouts index:%d error,empty address", i)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		//粉尘交易金额
		if v.ToAmount < 546 {
			err = fmt.Errorf("Txouts index:%d，address:%s, error ToAmount,min:546,now:%d", i, v.ToAddress, v.ToAmount)
			api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
			return
		}
		toAmount += v.ToAmount
		tpl.TxOuts = append(tpl.TxOuts, bo.BchTxOutTpl{
			ToAddr:   v.ToAddress,
			ToAmount: v.ToAmount,
		})
	}

	changeAmount = fromAmount - toAmount - txinput.Fee

	if changeAmount < 0 {
		err = fmt.Errorf("amount error ,fromAmount:%d,toAmount:%d,txinput.Fee:%d", fromAmount, toAmount, txinput.Fee)
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}

	if changeAmount < 546 {
		//附加进去手续费 ,不处理即可

	}
	if changeAmount > 546 {
		//发生找零，自动吧changeAddress附加进去txout
		tpl.TxOuts = append(tpl.TxOuts, bo.BchTxOutTpl{
			ToAmount: changeAmount,
			ToAddr:   txinput.ChangeAddr,
		})
	}
	api.HttpResponse(c, http.StatusOK, e.SUCCESS, tpl)
}

// @Summary 模板签名，冷系统有效，热系统无效
// @Produce  json
// @param body body string true "sign json"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/sign [post]
func (a *API) SignTx(c *gin.Context) {
	if conf.GlobalConf.SystemModel != "cold" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "Model is not cold", nil)
		return
	}
	//bind不解析请求头不为json的数据
	//err := c.Bind(tpl)
	//此处使用byte解析
	tpl := new(bo.BchTxTpl)
	data, err := c.GetRawData()
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), vo.SignResult{
			MchInfo: tpl.MchInfo,
		})
		return
	}

	json.Unmarshal(data, tpl)

	hex, err := a.bchService.SignTx(tpl)
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), vo.SignResult{
			MchInfo: tpl.MchInfo,
		})
		return
	}
	fmt.Println(fmt.Sprintf(" response:%+v", vo.SignResult{
		Hex:     hex,
		MchInfo: tpl.MchInfo,
	}))
	api.HttpResponse(c, http.StatusOK, e.SUCCESS, vo.SignResult{
		Hex:     hex,
		MchInfo: tpl.MchInfo,
	})
}

func (a *API) GetPrivkey(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty address", nil)
		return
	}
	privkey, _ := global.GetValue(address)
	if privkey == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty privkey", nil)
		return
	}
	api.HttpResponse(c, http.StatusOK, e.SUCCESS, "has key")
}

// @Summary 广播交易
// @Produce  json
// @param body body string true "push hex"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/push [post]
func (a *API) SendTx(c *gin.Context) {
	var (
		req      *http.Request
		bodyData []byte
	)

	datas, err := c.GetRawData()
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}
	pushParam := new(bo.PushInput)
	err = json.Unmarshal(datas, pushParam)
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}

	if pushParam.Hex == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty hex", nil)
		return
	}
	sendUrl := conf.GlobalConf.BchCfg.PushServers
	if len(sendUrl) == 0 {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty push server", nil)
		return
	}
	result := new(bo.SendTxResult)
	for _, v := range sendUrl {
		req, err = getJsonPostReq(v, datas)
		if err != nil {
			continue
		}
		bodyData, err = respone(req)
		if err != nil {
			continue
		}
		err = json.Unmarshal(bodyData, result)
		if err != nil {
			continue
		}
		if result.Code != 0 {
			err = fmt.Errorf("push error:%s", result.Message)
			continue
		}
		if result.Data.Txid != "" {
			api.HttpResponse(c, http.StatusOK, e.SUCCESS, vo.PushResult{
				MchInfo: pushParam.MchInfo,
				TxID:    result.Data.Txid,
			})
			return
		}
	}
	//如果在for循环中没有返回结果，那么肯定是出错了
	api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)

}

// @Summary 生成地址,冷系统有效，热系统无效
// @Produce  json
// @param body body string true "{'num':10,'orderId':'123456','mchId':'test','coinName':'bch'}"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/addr [post]
func (a *API) CreateAddrs(c *gin.Context) {

	if conf.GlobalConf.SystemModel != "cold" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "Model is not cold", nil)
		return
	}
	params := new(bo.CreateAddrParam)
	c.BindJSON(params)
	//限制100w数量
	if params.Num > 1000000 {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "num too big", nil)
		return
	}
	if params.MchId == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty mchId", nil)
		return
	}
	if params.CoinName != "bch" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "error coinName", nil)
		return
	}
	if params.OrderId == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "empty orderId", nil)
		return
	}
	filepath := conf.GlobalConf.BchCfg.CreateAddrPath
	copyPath := conf.GlobalConf.BchCfg.AddrPath
	resultVo, err := a.bchService.CreateAddr(params, filepath, copyPath)
	if err != nil {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, err.Error(), nil)
		return
	}
	api.HttpResponseByMsg(c, http.StatusOK, e.SUCCESS, e.GetMsg(e.SUCCESS), resultVo)

}

func (a *API) CheckAddr(c *gin.Context) {

	if conf.GlobalConf.SystemModel != "cold" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "Model is not cold", nil)
		return
	}
	return
}

// @Summary 临时导入地址私钥，冷系统有效，热系统失效
// @Produce  multipart/form-data
// @Accept  mpfd
// @Param address formData string true "address"
// @param privkey formData string true "privkey"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/importpk [post]
func (a *API) ImportAddr(c *gin.Context) {
	if conf.GlobalConf.SystemModel != "cold" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "Model is not cold", nil)
		return
	}
	address := c.PostForm("address")
	privkey := c.PostForm("privkey")
	if address == "" || privkey == "" {
		api.HttpResponseByMsg(c, http.StatusOK, e.ERROR, "error params", nil)
		return
	}
	global.SetValue(address, privkey)
	api.HttpResponseByMsg(c, http.StatusOK, e.SUCCESS, e.GetMsg(e.SUCCESS), nil)
}

//==========================================私有方法==========================================

func getJsonPostReq(httpUrl string, body []byte) (*http.Request, error) {
	return getRequest("POST", httpUrl, map[string]string{"Content-Type": "application/json", "Connection": "keep-alive"}, body)
}

func getJsonGetReq(httpUrl string) (*http.Request, error) {
	return getRequest("GET", httpUrl, map[string]string{"Content-Type": "application/json", "Connection": "keep-alive"}, nil)
}

//获取request
func getRequest(method, httpUrl string, reqHeader map[string]string, body []byte) (*http.Request, error) {
	var (
		req *http.Request
	)
	_, err := url.Parse(httpUrl)
	if err != nil {
		return nil, err
	}
	req, _ = http.NewRequest(method, httpUrl, bytes.NewBuffer(body))
	req.Close = true
	if reqHeader != nil && len(reqHeader) > 0 {
		for k, v := range reqHeader {
			req.Header.Set(k, v)
		}
	}
	//req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Connection", "keep-alive")
	return req, nil
}

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 60 * time.Second,
			Timeout:   60 * time.Second,
		}).DialContext,
		IdleConnTimeout: 30 * time.Second,
	},
}

func respone(req *http.Request) ([]byte, error) {
	var (
		resp *http.Response
		err  error
		ds   []byte
	)
	if resp, err = client.Do(req); err != nil {
		return nil, err
	}
	ds, err = ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return nil, err
	}
	//http.StatusOK
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(string(ds))
	}
	return ds, nil
}

//==========================================私有方法==========================================
