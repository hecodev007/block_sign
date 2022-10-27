package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/mtrserver/api"
	"github.com/group-coldwallet/mtrserver/conf"
	"github.com/group-coldwallet/mtrserver/model/bo"
	"github.com/group-coldwallet/mtrserver/model/global"
	"github.com/group-coldwallet/mtrserver/model/vo"
	"github.com/group-coldwallet/mtrserver/pkg/httpresp"
	"github.com/group-coldwallet/mtrserver/pkg/mtrutil"
	"github.com/group-coldwallet/mtrserver/pkg/util"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/zwjlink/meterutil/meter"
	"github.com/zwjlink/meterutil/tx"
	"io/ioutil"
	"net/http"
	"strings"
)

// @Summary 交易签名
// @Produce  json
// @param body body string true "sign json"
// @Success 200 {string} json "{"code":0,"message":"ok","data":"123123123"}"
// @Router /v1/transfer [post]
func Transfer(c *gin.Context) {
	var (
		coinName      string
		toAmountFloat decimal.Decimal
		token         tx.TokenType
		feeFloat      = decimal.NewFromFloat(0.0105) //预扣手续费
	)

	params := new(bo.MtrParams)
	data, err := c.GetRawData()
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	json.Unmarshal(data, params)
	err = params.Check()
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	coinName = strings.ToLower(params.CoinName)
	toAmountInt, _ := decimal.NewFromString(params.ToAmountInt64)
	toAmountFloat = toAmountInt.Shift(-18)
	tpl := &bo.TxTpl{mtrutil.MtrTpl{
		ChainTag:        82,
		Expiration:      18,
		Gas:             21000,
		GasPriceCoef:    0,
		BlockRef:        0,
		Outs:            make([]mtrutil.ToTpl, 0),
		FromAddr:        params.FromAddr,
		FromPrivKey:     "",
		FeeAddr:         params.FeeAddr,
		FeePayerPrivKey: "",
	}}
	//填充私钥
	privkey, _ := global.GetValue(params.FromAddr)
	if privkey == "" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, fmt.Sprintf("address:%s,miss key", params.FromAddr), nil)
		return
	}
	tpl.FromPrivKey = privkey

	if params.FeeAddr != "" {
		feekey, _ := global.GetValue(params.FeeAddr)
		if feekey == "" {
			httpresp.HttpRespErrorByMsg(c, http.StatusOK, fmt.Sprintf("address:%s,miss key", params.FeeAddr), nil)
			return
		}
	}

	//设置引用高度
	url := fmt.Sprintf("%s/blocks/best", conf.GlobalConf.MtrCfg.Servers[0])
	resp, err := util.HttpGet(url)
	if err != nil {
		log.Info(url)
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	var block map[string]interface{}
	json.Unmarshal(resp, &block)
	if block["number"] == nil || block["number"].(float64) <= 0 {
		err = fmt.Errorf("get latest block number error,resp=%s", string(resp))
		log.Info(err.Error())
		log.Info(url)
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	blockNumber := block["number"].(float64)
	tpl.BlockRef = uint32(blockNumber) + 10 //延迟10个区块

	//校验from地址金额
	url = fmt.Sprintf("%s/accounts/%s", conf.GlobalConf.MtrCfg.Servers[0], params.FromAddr)
	resp, err = util.HttpGet(url)
	if err != nil {
		log.Info(url)
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	balance := make(map[string]interface{}, 0)
	json.Unmarshal(resp, &balance)
	if balance["balance"] == nil || balance["energy"] == "" {
		err = fmt.Errorf("get balance error,resp=%s", string(resp))
		log.Info(url)
		log.Info(err.Error())
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	mtrBalanceInt, _ := hexutil.DecodeBig(balance["energy"].(string))
	mtrBalanceFloat := decimal.NewFromBigInt(mtrBalanceInt, -18)
	mtrgBalanceInt, _ := hexutil.DecodeBig(balance["balance"].(string))
	mtrgBalanceFloat := decimal.NewFromBigInt(mtrgBalanceInt, -18)

	//因为上面的check方法限制了token，所以只需要判断0和1
	if params.Token == 0 {

		//mtr交易
		token = tx.MeterToken
		if mtrBalanceFloat.Sub(feeFloat).Sub(toAmountFloat).LessThan(decimal.Zero) {
			//余额不足
			err = fmt.Errorf("mtr balance error,address=%s,mtrbalance:%s,fee:%s,toamount:%s",
				params.FromAddr, mtrBalanceFloat.String(), feeFloat.String(), toAmountFloat.String())
			log.Info(err.Error())
			httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
			return
		}
	} else {
		//mtrg交易
		token = tx.MeterGovToken
		if mtrBalanceFloat.LessThan(feeFloat) {
			//手续费不足
			err = fmt.Errorf("mtr balance error,address=%s,mtrbalance:%s,need fee:%s,",
				params.FromAddr, mtrBalanceFloat.String(), feeFloat.String())
			httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
			return
		}

		if mtrgBalanceFloat.LessThan(toAmountFloat) {
			err = fmt.Errorf("mtrg balance error,address=%s,mtrgbalance:%s,toAmount:%s,",
				params.FromAddr, mtrgBalanceFloat.String(), toAmountFloat.String())
			httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
			return
		}
	}

	tpl.Outs = []mtrutil.ToTpl{
		mtrutil.ToTpl{
			ToAddress:   params.ToAddr,
			ToAmountInt: toAmountInt.BigInt(),
			Token:       token,
		},
	}
	hex, err := api.ChainService[coinName].SignTx(tpl)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	raw := HexRaw{
		Raw: hex,
	}
	rawdata, _ := json.Marshal(raw)
	txid, err := sendRaw(rawdata)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	httpresp.HttpRespOkByMsg(c, "ok", txid)
}

// @Summary 生成地址
// @Produce  json
// @param body body string true "{'num':10,'orderId':'123456','mchId':'test','coinName':'mtr'}"
// @Success 200 {string} json "{"code":0,"message":"ok","data":{},"hash":"8978608dad8f150ea142e1c076f6564e"}"
// @Router /v1/createaddr [post]
func CreateAddrs(c *gin.Context) {

	if conf.GlobalConf.SystemModel != "cold" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "not cold", nil)
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
	if params.CoinName != "mtr" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "error coinName", nil)
		return
	}
	if params.OrderId == "" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "error orderId", nil)
		return
	}
	filepath := conf.GlobalConf.MtrCfg.CreateAddrPath
	resultVo, err := api.ChainService["mtr"].CreateAddr(params, filepath)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	httpresp.HttpRespOkByMsg(c, "ok", resultVo)
}

func GetBalance(c *gin.Context) {
	addr, _ := c.GetQuery("addr")
	if addr == "" {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, "empty addr", nil)
		return
	}
	act, err := meter.ParseAddress(addr)
	if err != nil {
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}

	url := fmt.Sprintf("%s/accounts/%s", conf.GlobalConf.MtrCfg.Servers[0], act.String())
	resp, err := util.HttpGet(url)
	if err != nil {
		log.Info(url)
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	balance := make(map[string]interface{}, 0)
	json.Unmarshal(resp, &balance)
	if balance["balance"] == nil || balance["energy"] == "" {
		err = fmt.Errorf("get balance error,resp=%s", string(resp))
		log.Info(url)
		log.Info(err.Error())
		httpresp.HttpRespErrorByMsg(c, http.StatusOK, err.Error(), nil)
		return
	}
	mtrBalanceInt, _ := hexutil.DecodeBig(balance["energy"].(string))
	mtrBalanceFloat := decimal.NewFromBigInt(mtrBalanceInt, -18)
	mtrgBalanceInt, _ := hexutil.DecodeBig(balance["balance"].(string))
	mtrgBalanceFloat := decimal.NewFromBigInt(mtrgBalanceInt, -18)

	resultVo := make([]vo.Mtrbalance, 0)
	resultVo = append(resultVo, vo.Mtrbalance{
		CoinName:     "mtr",
		Decimal:      18,
		BalanceFloat: mtrBalanceFloat.String(),
	})

	resultVo = append(resultVo, vo.Mtrbalance{
		CoinName:     "mtrg",
		Decimal:      18,
		BalanceFloat: mtrgBalanceFloat.String(),
	})
	httpresp.HttpRespOkByMsg(c, "ok", resultVo)
}

//=====================================================send======================================================

type HexRaw struct {
	Raw string `json:"raw"`
}

type MtrSendResult struct {
	Id string `json:"id"`
}

//=====================================================send======================================================

func sendRaw(data []byte) (txid string, err error) {
	//jsonStr :=[]byte(`{ "username": "auto", "password": "auto123123" }`)
	url := conf.GlobalConf.MtrCfg.PushServers[0] + "/transactions"
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

	result := new(MtrSendResult)
	json.Unmarshal(body, result)
	if result.Id == "" {
		return "", errors.New(string(body))
	} else {
		txid = result.Id
		return txid, nil
	}
}
