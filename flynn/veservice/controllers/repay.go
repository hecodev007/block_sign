package controllers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/tx"
	"strings"
	"time"
	comm "veservice/common"
)

//代付

type RepayController struct {
	beego.Controller
}

func (c *RepayController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"result": nil,
		"error":  nil,
	}
	set_resp := func(result, err interface{}) {
		resp["result"] = result
		resp["error"] = err
		c.Data["json"] = resp
		c.ServeJSON()

	}
	if c.Ctx.Input.RequestBody == nil {
		set_resp(nil, "request params is null")
		return
	}
	var jsonObj map[string]interface{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
	if err != nil {
		set_resp(nil, err)
		return
	}
	from, to, repay, amount := jsonObj["from"].(string), jsonObj["to"].(string), jsonObj["repay"].(string), jsonObj["amount"].(string)
	if from == "" || to == "" || repay == "" || amount == "" {
		set_resp(nil, fmt.Sprintf("params have null,from=[%s],to=[%s],repay=[%s],amount=[%s]", from, to, repay, amount))
		return
	}
	from = strings.ToLower(from)
	repay = strings.ToLower(repay)
	toThor, _ := thor.ParseAddress(to)
	onlineAmount, _, errAmount := getAddressAmount(from)
	if errAmount != nil {
		set_resp(nil, errAmount)
		return
	}
	_tmpamount, err := decimal.NewFromString(amount)
	if err != nil {
		beego.Debug("convert amount fail !")
		set_resp(nil, "convert amount fail !")
		return
	}
	amountB := _tmpamount.Coefficient()
	if onlineAmount.Cmp(amountB) < 0 {
		set_resp(nil, fmt.Errorf("online amount is less than transfer amount,OnlineAmount=[%d],TransAmount=[%d]", onlineAmount.Int64(), amountB.Int64()))
		return
	}
	_, repayAmount, errRepay := getAddressAmount(repay)
	if errRepay != nil {
		set_resp(nil, errRepay)
		return
	}
	if repayAmount.Cmp(vthoLimit) < 0 {
		set_resp(nil, errors.New("repay address amount is less than 100"))
		return
	}
	var blockNumber uint32
	respData, err := comm.GetJson(fmt.Sprintf("%s/blocks/best", beego.AppConfig.String("url")), nil)
	if err != nil {
		set_resp(nil, err)
		return
	}
	var block map[string]interface{}
	err1 := json.Unmarshal(respData, &block)
	if err1 != nil || block == nil {
		set_resp(nil, err)
		return
	}
	blockNumber = uint32(block["number"].(float64)) + 10
	beego.Debug(blockNumber)
	var feat tx.Features
	chainTag := beego.AppConfig.DefaultString("chaintag", "0x4a")
	tag, _ := hex.DecodeString(removeHex0x(chainTag))
	feat.SetDelegated(true)
	gas := beego.AppConfig.DefaultInt64("gas", 60000)
	trx := new(tx.Builder).ChainTag(tag[0]).BlockRef(tx.NewBlockRef(blockNumber)).
		Expiration(720).
		Clause(tx.NewClause(&toThor).WithValue(amountB)).
		GasPriceCoef(0).
		Gas(uint64(gas)).
		DependsOn(nil).
		Features(feat).
		Nonce(uint64(time.Now().UnixNano())).
		Build()
	var (
		wif, repayWif string
		errA          error
	)
	wif, errA = comm.AesDecrypt(EncryptWifMap[from], []byte(WifKeyListMap[from]))
	if errA != nil {
		set_resp(nil, fmt.Sprintf("do not find private key,address=[%s],Err=[%v]", from, errA))
	}
	privKey, _ := crypto.HexToECDSA(wif)
	if privKey == nil {
		set_resp(nil, "parse private key error")
		return
	}
	repayWif, errA = comm.AesDecrypt(EncryptWifMap[repay], []byte(WifKeyListMap[repay]))
	if errA != nil {
		set_resp(nil, fmt.Sprintf("do not find private key,address=[%s]", repay))
		return
	}
	repayPrivKey, _ := crypto.HexToECDSA(repayWif)
	if repayPrivKey == nil {
		set_resp(nil, "parse repay private key error")
		return
	}
	//from 签名
	sig, err1 := crypto.Sign(trx.SigningHash().Bytes(), privKey)
	if err1 != nil {
		set_resp(nil, "sign error")
		return
	}
	o := crypto.PubkeyToAddress(privKey.PublicKey)
	hash := trx.DelegatorSigningHash(thor.Address(o))
	sigRepay, err2 := crypto.Sign(hash.Bytes(), repayPrivKey)
	if err2 != nil {
		set_resp(nil, "sign repay error")
		return
	}
	sig = append(sig, sigRepay...)
	trx = trx.WithSignature(sig)
	signdata, _ := rlp.EncodeToBytes(trx)
	hextx := hex.EncodeToString(signdata)
	if !strings.HasPrefix(hextx, "0x") {
		hextx = "0x" + hextx
	}
	url := fmt.Sprintf("%s/transactions", beego.AppConfig.String("url"))
	reqbody := map[string]interface{}{
		"raw": hextx,
	}
	respData, err3 := comm.PostJson(url, reqbody)
	if err != nil || len(respData) == 0 {
		resp["error"] = fmt.Sprintf("broadcast error,err=%v", err3)
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	var res map[string]string
	json.Unmarshal(respData, &res)

	resp["result"] = res["id"]
	c.Data["json"] = resp
	c.ServeJSON()
	beego.Debug("repay 交易")
}
