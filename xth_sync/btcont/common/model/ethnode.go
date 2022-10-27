package model

import (
	"btcont/common/conf"
	"btcont/common/log"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var eth_node_url = conf.Cfg.EthNode

type EthNode struct {
	Height int64
	Uptime time.Time
}

func (sc *EthNode) BalanceOf(addr string, contract string) (amount decimal.Decimal, t int64, err error) {
	if contract != "" {
		return decimal.Decimal{}, 0, errors.New("不支持代币")
	}
	amount, err = sc.balanceOf(addr, 0)
	if err != nil {
		log.Info(err.Error())
		return
	}
	bestHeight, err := sc.BlockCount()
	if err != nil {
		log.Info(err.Error())
		return
	}
	num := int64(1000)
	amountbefore, err := sc.balanceOf(addr, bestHeight-num)
	if err != nil {
		log.Info(err.Error())
		return
	}
	if amount.Cmp(amountbefore) == 0 {
		t = time.Now().Unix() - num*15
	} else {
		t = time.Now().Unix()
	}
	return amount, t, nil
}

type EthBlockContResponse struct {
	Result string `json:"result"`
	Error  struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}
}

func (sc *EthNode) BlockCount() (height int64, err error) {
	//time.Sleep(time.Second * 10)
	if time.Since(sc.Uptime) > time.Minute || sc.Height == 0 {
		//log.Info("BlockCount", time.Now().String())
		if height, err = sc.blockCount(); err != nil {
			return 0, err
		}
		sc.Height = height
		sc.Uptime = time.Now()
		return height, err
	} else {
		return sc.Height, nil
	}
}
func (sc *EthNode) blockCount() (h int64, err error) {
	payload := strings.NewReader(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}`)
	resp, err := http.Post(eth_node_url, "application/json", payload)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	ret := new(EthBlockContResponse)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return 0, err
	}
	if ret.Error.Code != 0 {
		return 0, errors.New(ret.Error.Message)
	}
	height, err := strconv.ParseInt(strings.TrimPrefix(ret.Result, "0x"), 16, 64)

	return height, err
}
func (sc *EthNode) balanceOf(addr string, height int64) (amount decimal.Decimal, err error) {
	var tag string
	if height == 0 {
		tag = "pending"
	} else {
		tag = "0x" + strconv.FormatInt(height, 16)
	}

	payload := strings.NewReader(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"%v\",\"%v\"],\"id\":\"22\"}", addr, tag))
	resp, err := http.Post(eth_node_url, "application/json", payload)
	if err != nil {
		log.Info(err.Error())
		return decimal.Decimal{}, err
	}
	if resp.StatusCode != 200 {
		return decimal.Decimal{}, errors.New(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info(err.Error())
		return decimal.Decimal{}, err
	}
	ret := new(EthBlockContResponse)
	err = json.Unmarshal(body, ret)
	if err != nil {
		log.Info(err.Error())
		return decimal.Decimal{}, err
	}
	if ret.Error.Code != 0 {
		log.Info(ret.Error.Message)
		return decimal.Decimal{}, errors.New(ret.Error.Message)
	}
	amountHex := strings.TrimPrefix(ret.Result, "0x")
	if len(amountHex)%2 != 0 {
		amountHex = "0" + amountHex
	}
	//log.Info(amountHex)
	retBytes, err := hex.DecodeString(amountHex)
	if err != nil {
		log.Info(err.Error())
		return decimal.Decimal{}, err
	}
	amount = decimal.NewFromBigInt(big.NewInt(0).SetBytes(retBytes), 0)

	return amount.Shift(-18), err
}
