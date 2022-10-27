package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/dogeserver/api"
	"github.com/group-coldwallet/dogeserver/conf"
	"github.com/group-coldwallet/dogeserver/model/bo"
	"github.com/group-coldwallet/dogeserver/pkg/dogeutil"
	"github.com/group-coldwallet/dogeserver/pkg/httpresp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// @Summary 交易签名
// @Produce  json
// @param body body string true "sign json"
// @Success 200 {string} json "{"code":0,"message":"ok","data":"123123123"}"
// @Router /v1/transfer [post]
func Transfer(c *gin.Context) {
	tpl := new(dogeutil.DogeTxTpl)
	data, err := c.GetRawData()
	log.Println("data:", string(data))
	if err != nil {
		httpresp.HttpRespError(c)
		return
	}
	json.Unmarshal(data, tpl)

	coinName := strings.ToLower(tpl.CoinName)

	hex, err := api.ChainService[coinName].SignTx(tpl)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	raw := HexRaw{
		Hex: hex,
	}
	rawdata, _ := json.Marshal(raw)
	log.Println("raw:", raw)
	txid, err := sendRaw(rawdata)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	httpresp.HttpRespOkByMsg(c, "ok", txid)
}

// @Summary 生成地址
// @Produce  json
// @param body body string true "{'num':10,'orderId':'123456','mchId':'test','coinName':'btc'}"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/createaddr [post]
func CreateAddrs(c *gin.Context) {

	if conf.GlobalConf.SystemModel != "cold" {
		httpresp.HttpRespError(c)
		return
	}
	params := new(bo.CreateAddrParam)
	c.BindJSON(params)
	//限制5w数量
	if params.Num > 50000 {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "num too big，50000", nil)
		return
	}
	if params.MchId == "" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "empty mchId", nil)
		return
	}
	if params.CoinName != "doge" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "error coinName", nil)
		return
	}
	if params.OrderId == "" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "error orderId", nil)
		return
	}
	filepath := conf.GlobalConf.DogeCfg.CreateAddrPath
	resultVo, err := api.ChainService["doge"].CreateAddr(params, filepath)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	httpresp.HttpRespOkByMsg(c, "ok", resultVo)
}

//=====================================================send======================================================

type HexRaw struct {
	Hex string `json:"hex"`
}

type DogeSendResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Txid string `json:"txid"`
	} `json:"data"`
}

//=====================================================send======================================================

func sendRaw(data []byte) (txid string, err error) {
	//jsonStr :=[]byte(`{ "username": "auto", "password": "auto123123" }`)
	//url := "http://47.244.140.180:9999/api/v1/doge/send"
	if len(conf.GlobalConf.DogeCfg.PushServers) == 0 {
		return "", errors.New("miss push url")
	}
	url := conf.GlobalConf.DogeCfg.PushServers[0]
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}
	result := new(DogeSendResult)
	json.Unmarshal(body, result)
	if result.Code != 0 && result.Data.Txid == "" {
		return "", errors.New(string(body))
	} else {
		txid = result.Data.Txid
		return txid, nil
	}
}
