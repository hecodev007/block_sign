package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SYMBOL_SP31 = "SP31" //usdt币种
	SYMBOL_BTC  = "BTC"  //usdt币种
)

type AddrBalance struct {
	Balance []BalanceInfo `json:"balance"` //余额信息
}

type BalanceInfo struct {
	Symbol string `json:"symbol"` //币种类别
	Value  string `json:"value"`  //可用金额
}

type AddrInfo struct {
	Address     string
	UsdtBalance int64
	BtcBalance  int64
}

type AddressResult struct {
	ErrNo  int        `json:"err_no"`
	ErrMsg string     `json:"err_msg"`
	Data   ResultData `json:"data"`
}
type ResultData struct {
	List []ResultDataList `json:"list"`
}
type ResultDataList struct {
	TxHash    string `json:"tx_hash"`
	TxOutputN int    `json:"tx_output_n"`
	Value     int64  `json:"value"`
}

type TxResult struct {
	ErrNo  int          `json:"err_no"`
	ErrMsg string       `json:"err_msg"`
	Data   TxResultData `json:"data"`
}

type TxResultData struct {
	Outputs []TxResultDataOuts `json:"outputs"`
}

type TxResultDataOuts struct {
	Addresses []string `json:"addresses"`
	Value     int64    `json:"value"`
	ScriptHex string   `json:"script_hex"`
}

//单转单归集
//addr：需要归集的地址
//toAddress：需要转账的地址
//changeaddress：找零的的地址
//gas 手续费
func MakeCollectionTxInput(addrs []string, toAddress string, changeaddress string, gas int64) (*models.TxInputNew, error) {
	addrs = StingArrayToRemoveRepeat(addrs)
	var (
		dataMap map[string]AddrBalance
		err     error
		//usdtTotalAmout int64
		//btcTotalAmout  int64
		errtext string
	)

	result := models.TxInputNew{}
	data := make(map[string]AddrInfo)
	action := true

	dataMap, err = GrabUsdtBalance(addrs)
	if err != nil {
		return nil, err
	}

	for _, v := range addrs {

		//第一次循环检测余额合法性
		usdtBalance, btcBalance := "", ""
		for _, vb := range dataMap[v].Balance {
			switch vb.Symbol {
			case SYMBOL_SP31:
				if vb.Value == "0" {
					action = false
					errtext += fmt.Sprintf("address:%v,usdt balance is 0，info：%v", v, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,usdt balance is 0，info：%v", v, vb))
				}
				usdtBalance = vb.Value
			case SYMBOL_BTC:
				if vb.Value == "0" {
					action = false
					errtext += fmt.Sprintf("address:%v,btc balance is 0，info：%v", v, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,btc balance is 0，info：%v", v, vb))
				}
				btcValue, _ := strconv.ParseInt(vb.Value, 10, 64)
				if (gas + 546) > btcValue {
					action = false
					errtext += fmt.Sprintf("1 address:%v,btc balance not enough，gas:%d,btc:%d,info：%v || ", v, gas, btcValue, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,btc fee is not enough，gas：%d,address-btc:%d，info:%v", v, gas.GasFee.HalfHourFee, btcValue, vb))
				}
				btcBalance = vb.Value
			}
		}
		btcAmount, _ := strconv.ParseInt(btcBalance, 10, 64)
		//btcTotalAmout += btcAmount
		usdtAmount, _ := strconv.ParseInt(usdtBalance, 10, 64)
		//usdtTotalAmout += usdtAmount
		fmt.Println(fmt.Sprintf("address: %v,usdt:%v, btc:%v,gas:%d, enough fee btc(gas.GasFee.HalfHourFee+546)：%t", v, usdtBalance, btcBalance, gas, btcAmount > (gas+546)))
		//txins := models.Txins{}
		//fmt.Println("btcTotalAmout====", btcTotalAmout)
		//fmt.Println("usdtTotalAmout====", usdtTotalAmout)
		data[v] = AddrInfo{
			Address:     v,
			UsdtBalance: usdtAmount,
			BtcBalance:  btcAmount,
		}

	}
	if !action {
		return nil, errors.New("Balance is error: " + errtext)
	}

	//第二次循环，获取地址可用的btc有效utxo
	for _, v := range addrs {
		txinsall := make([]models.Txins, 0)
		simpleTxIns := make([]models.OmniSimpleTxin, 0)
		txins := make([]models.Txins, 0)
		addressResult, err := GetTxByAddress(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("address:%v,error:%v", v, err))
		}
		time.Sleep(time.Millisecond * 500)

		if addressResult.ErrNo != 0 {
			return nil, errors.New(fmt.Sprintf("address:%v,addressResult error:%v", v, err))

		}
		if len(addressResult.Data.List) <= 0 {
			fmt.Println("List Error: 可消费余额为0", "  address:", v)
			continue
		}
		for _, va := range addressResult.Data.List {

			//简单的txin用于构造交易
			simpleTxIn := models.OmniSimpleTxin{}
			simpleTxIn.Vout = va.TxOutputN
			simpleTxIn.Txid = va.TxHash
			simpleTxIns = append(simpleTxIns, simpleTxIn)

			//签名txin结构体，需要补充 pubkey
			txin := models.Txins{}
			txin.Address = v
			txin.Txid = va.TxHash
			txin.Vout = va.TxOutputN
			txin.Amount = decimal.NewFromFloat(float64(va.Value) / 100000000)
			txins = append(txins, txin)
		}

		//获取公钥
		for _, vt := range txins {
			txResult, err := GetPybkeyByTx(vt.Txid)
			time.Sleep(time.Millisecond * 500)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Txid:%v,error:%v", vt.Txid, err))
			}
			if txResult.ErrNo != 0 {
				return nil, errors.New(fmt.Sprintf("Txid:%v,ErrNo error:%v", vt.Txid, err))
			}
			if len(txResult.Data.Outputs) <= 0 {
				return nil, errors.New(fmt.Sprintf("Txid:%v,List error:%v", vt.Txid, err))
			}
			for i, vr := range txResult.Data.Outputs {
				if i == vt.Vout {
					if vr.Addresses[0] == vt.Address {
						vtam, _ := vt.Amount.Float64()
						if (float64(vr.Value) / 100000000) == vtam {
							//补充数据
							vt.RedeemScript = ""
							vt.ScriptPubKey = vr.ScriptHex
							txinsall = append(txinsall, vt)
							break
						}
					}
				}
			}
		}

		toAmount := data[v].UsdtBalance

		txInput := models.TxInput{
			ChangeAddress: changeaddress, //找零给目标地址
			ToAmount:      decimal.New(toAmount, -8),
			ToAddress:     toAddress,
			Fee:           decimal.New(gas, -8),
		}
		for _, vt := range txinsall {
			//组装utxo
			txInput.Txins = append(txInput.Txins, vt)
		}
		result.Txinputs = append(result.Txinputs, txInput)
		fmt.Println("address:", v)
		fmt.Println("len(simpleTxIns:", len(simpleTxIns))
		fmt.Println("len(txins:", len(txins))
	}
	fmt.Println(fmt.Sprintf("%+v", result))
	jsonbyte, _ := json.MarshalToString(result)
	fmt.Println(jsonbyte)
	return &result, nil
}

//动态手续费构建
func MakeCollectionTxInputAutoFee(addrs []string, toAddress string, changeaddress string) (*models.TxInputNew, error) {
	addrs = StingArrayToRemoveRepeat(addrs)
	intputNum := 0 //txid数量 用于动态计算fee
	outputNum := 3 //输出数量，用于动态计算fee ，固定为3，因为2个和3个差异不大，考虑3个的原因是有btc找零地址（也许也会没有），btc接收地址，usdt接收地址
	var (
		dataMap map[string]AddrBalance
		err     error
		//usdtTotalAmout int64
		//btcTotalAmout  int64
		btcAmount  int64
		usdtAmount int64
		errtext    string
		gasOutput  *models.GasOutput
	)

	result := models.TxInputNew{}
	data := make(map[string]AddrInfo)
	action := true

	dataMap, err = GrabUsdtBalance(addrs)
	if err != nil {
		return nil, err
	}

	for _, v := range addrs {

		//第一次循环检测余额合法性
		usdtBalance, btcBalance := "", ""
		for _, vb := range dataMap[v].Balance {
			switch vb.Symbol {
			case SYMBOL_SP31:
				if vb.Value == "0" {
					action = false
					errtext += fmt.Sprintf("address:%v,usdt balance is 0，info：%v", v, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,usdt balance is 0，info：%v", v, vb))
				}
				usdtBalance = vb.Value
			case SYMBOL_BTC:
				if vb.Value == "0" {
					action = false
					errtext += fmt.Sprintf("address:%v,btc balance is 0，info：%v", v, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,btc balance is 0，info：%v", v, vb))
				}
				btcValue, _ := strconv.ParseInt(vb.Value, 10, 64)
				//必须有消耗，所以基准值使用1000，而546位接收地址所接收的btc，因此基准值为1546
				if btcValue <= 1546 {
					action = false
					errtext += fmt.Sprintf("address:%v,btc balance not enough，gas:%d,btc:%d,info：%v || ", v, 1546, btcValue, vb)
					//return nil, errors.New(fmt.Sprintf("address:%v,btc fee is not enough，gas：%d,address-btc:%d，info:%v", v, gas.GasFee.HalfHourFee, btcValue, vb))
				}
				btcBalance = vb.Value
			}
		}
		btcAmount, _ = strconv.ParseInt(btcBalance, 10, 64)
		usdtAmount, _ = strconv.ParseInt(usdtBalance, 10, 64)
		fmt.Println("btcAmount", btcAmount)
		fmt.Println("usdtAmount", usdtAmount)
		data[v] = AddrInfo{
			Address:     v,
			UsdtBalance: usdtAmount,
			BtcBalance:  btcAmount,
		}

	}
	if !action {
		return nil, errors.New("Balance is error: " + errtext)
	}

	//txinsall := make([]models.Txins, 0)

	//第二次循环，获取地址可用的btc有效utxo
	for _, v := range addrs {
		simpleTxIns := make([]models.OmniSimpleTxin, 0)
		txins := make([]models.Txins, 0)
		txinsall := make([]models.Txins, 0)
		addressResult, err := GetTxByAddress(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("address:%v,error:%v", v, err))
		}
		time.Sleep(time.Millisecond * 500)

		if addressResult.ErrNo != 0 {
			return nil, errors.New(fmt.Sprintf("address:%v,addressResult error:%v", v, err))

		}
		if len(addressResult.Data.List) <= 0 {
			fmt.Println("List Error: 可消费余额为0", "  address:", v)
			continue
		}
		for _, va := range addressResult.Data.List {

			//简单的txin用于构造交易
			simpleTxIn := models.OmniSimpleTxin{}
			simpleTxIn.Vout = va.TxOutputN
			simpleTxIn.Txid = va.TxHash
			simpleTxIns = append(simpleTxIns, simpleTxIn)

			//签名txin结构体，需要补充 pubkey
			txin := models.Txins{}
			txin.Address = v
			txin.Txid = va.TxHash
			txin.Vout = va.TxOutputN
			txin.Amount = decimal.NewFromFloat(float64(va.Value) / 100000000)
			txins = append(txins, txin)
		}

		//获取公钥

		for _, vt := range txins {
			txResult, err := GetPybkeyByTx(vt.Txid)
			time.Sleep(time.Millisecond * 500)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Txid:%v,error:%v", vt.Txid, err))
			}
			if txResult.ErrNo != 0 {
				return nil, errors.New(fmt.Sprintf("Txid:%v,ErrNo error:%v", vt.Txid, err))
			}
			if len(txResult.Data.Outputs) <= 0 {
				return nil, errors.New(fmt.Sprintf("Txid:%v,List error:%v", vt.Txid, err))
			}
			for i, vr := range txResult.Data.Outputs {
				if i == vt.Vout {
					if vr.Addresses[0] == vt.Address {
						vtam, _ := vt.Amount.Float64()
						if (float64(vr.Value) / 100000000) == vtam {
							//补充数据
							vt.RedeemScript = ""
							vt.ScriptPubKey = vr.ScriptHex
							txinsall = append(txinsall, vt)
							break
						}
					}
				}
			}
		}

		intputNum = len(simpleTxIns)
		//计算公式   (in*148+34*out+10)* X satoshis / byte
		//也就是   (in*148+34*out+10)* X basegas

		gasinput := models.GasInput{
			InNum:  intputNum,
			OutNum: outputNum,
		}
		gasOutput, err = getGasInput(&gasinput)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("get gas error,%s", err.Error()))
		}
		gas := gasOutput.GasFee.HalfHourFee

		usdtAmount = data[v].UsdtBalance
		btcAmount = data[v].BtcBalance
		if btcAmount <= gas+546 {
			return nil, errors.New(fmt.Sprintf("address:%v,btc fee is not enough，gas：%d,toAddressBtc 546,address-btc:%d", v, gas, btcAmount))
		}
		//toamount := fmt.Sprintf("%.8f", float64(usdtAmount)/100000000)
		toamount := decimal.New(usdtAmount, -8)
		txInput := models.TxInput{
			ChangeAddress: changeaddress, //找零给目标地址
			ToAmount:      toamount,
			ToAddress:     toAddress,
			Fee:           decimal.NewFromFloat(float64(gas) / 100000000),
		}
		for _, vt := range txinsall {
			//组装utxo
			txInput.Txins = append(txInput.Txins, vt)
		}
		result.Txinputs = append(result.Txinputs, txInput)
		fmt.Println("address:", v)
		fmt.Println("len(simpleTxIns:", len(simpleTxIns))
		fmt.Println("len(txins:", len(txins))
	}
	fmt.Println(fmt.Sprintf("%+v", result))
	jsonbyte, _ := json.MarshalToString(result)
	fmt.Println(jsonbyte)
	return &result, nil
}

//addr：需要归集的地址
//toAddress：需要转账的地址
//changeaddress：找零的的地址
//feeaddress
//gas 手续费
func MakeCollectionTxInputUseFee(addr, toAddress, changeaddress, feeaddress string, gas int64) (*models.TxInputNew, error) {
	if addr == feeaddress {
		return nil, errors.New("utxo地址不能与代扣地址一致")
	}
	addrs := []string{}
	addrs = append(addrs, addr)
	addrs = append(addrs, feeaddress)
	var (
		dataMap map[string]AddrBalance
		err     error
	)
	result := models.TxInputNew{}
	dataMap, err = GrabUsdtBalance(addrs)
	if err != nil {
		return nil, err
	}

	usdtBalance, usdtbtcBalance, btcBalance := "", "", ""
	//检查usdt余额
	for _, vb := range dataMap[addr].Balance {
		if vb.Symbol == SYMBOL_SP31 {
			if vb.Value == "0" {
				return nil, errors.New(fmt.Sprintf("address:%v,usdt balance is 0，info：%v", addr, vb))
			}
			usdtBalance = vb.Value
		}

		if vb.Symbol == SYMBOL_BTC {
			if vb.Value == "0" {
				return nil, errors.New(fmt.Sprintf("address:%v,must have btc utxo", addr))
			}
			usdtbtcBalance = vb.Value
		}
	}

	//检查手续费btc余额
	for _, vb := range dataMap[changeaddress].Balance {
		if vb.Symbol == SYMBOL_BTC {
			if vb.Value == "0" {
				return nil, errors.New(fmt.Sprintf("address:%v,btc balance is 0，info：%v", changeaddress, vb))
			}
			btcBalance = vb.Value
		}
	}
	ubtc, _ := strconv.ParseFloat(usdtbtcBalance, 64)
	btc, _ := strconv.ParseFloat(btcBalance, 64)
	totalbtc := ubtc + btc
	feeDecimal := float64(gas) / 100000000
	if totalbtc < feeDecimal {
		fmt.Println(fmt.Sprintf("usdt:%d,usdtbtc:%d, btc:%d,gas:%d, enough fee btc(gas.GasFee.HalfHourFee+546)：%t", usdtBalance, usdtbtcBalance, btcBalance, gas, (ubtc+btc) > (float64(gas)/100000000+0.00000546)))
		return nil, errors.New(fmt.Sprintf("usdt:%d,usdtbtc:%d, btc:%d,gas:%d, enough fee btc(gas.GasFee.HalfHourFee+546)：%t", usdtBalance, usdtbtcBalance, btcBalance, gas, (ubtc+btc) > (float64(gas)/100000000+0.00000546)))
	}

	simpleTxIns := make([]models.OmniSimpleTxin, 0)
	txins := make([]models.Txins, 0)
	txinsnew := make([]models.Txins, 0)

	//第二次循环，获取地址可用的btc有效utxo
	for _, v := range addrs {
		addressResult, err := GetTxByAddress(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("address:%v,error:%v", v, err))
		}
		time.Sleep(time.Millisecond * 500)

		if addressResult.ErrNo != 0 {
			return nil, errors.New(fmt.Sprintf("address:%v,addressResult error:%v", v, err))

		}
		if len(addressResult.Data.List) <= 0 {
			fmt.Println("List Error: 可消费余额为0", "  address:", v)
			continue
		}
		for _, va := range addressResult.Data.List {
			//简单的txin用于构造交易
			simpleTxIn := models.OmniSimpleTxin{}
			simpleTxIn.Vout = va.TxOutputN
			simpleTxIn.Txid = va.TxHash
			simpleTxIns = append(simpleTxIns, simpleTxIn)

			//签名txin结构体，需要补充 pubkey
			txin := models.Txins{}
			txin.Address = v
			txin.Txid = va.TxHash
			txin.Vout = va.TxOutputN
			txin.Amount = decimal.NewFromFloat(float64(va.Value) / 100000000)
			txins = append(txins, txin)
		}
	}

	//获取公钥
	for _, vt := range txins {
		txResult, err := GetPybkeyByTx(vt.Txid)
		time.Sleep(time.Millisecond * 500)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Txid:%v,error:%v", vt.Txid, err))
		}
		if txResult.ErrNo != 0 {
			return nil, errors.New(fmt.Sprintf("Txid:%v,ErrNo error:%v", vt.Txid, err))
		}
		if len(txResult.Data.Outputs) <= 0 {
			return nil, errors.New(fmt.Sprintf("Txid:%v,List error:%v", vt.Txid, err))
		}
		for i, vr := range txResult.Data.Outputs {
			if i == vt.Vout {
				if vr.Addresses[0] == vt.Address {
					vtam, _ := vt.Amount.Float64()
					if (float64(vr.Value) / 100000000) == vtam {
						//补充数据
						vt.RedeemScript = ""
						vt.ScriptPubKey = vr.ScriptHex
						//组装utxo
						txinsnew = append(txinsnew, vt)
						break
					}
				}
			}
		}
	}
	//amout, _ := strconv.ParseFloat(usdtBalance, 64)
	amoutStr, _ := decimal.NewFromString(usdtBalance)
	amout := amoutStr.Shift(-8)
	txInput := models.TxInput{
		ChangeAddress: changeaddress, //找零给目标地址
		ToAmount:      amout,
		ToAddress:     toAddress,
		Fee:           decimal.NewFromFloat(float64(gas) / 100000000),
	}
	txInput.Txins = append(txInput.Txins, txinsnew...)
	result.Txinputs = append(result.Txinputs, txInput)
	return &result, nil
}

//找到尚未花费的余额
func GrabUsdtBalance(addrs []string) (map[string]AddrBalance, error) {

	var (
		buf  bytes.Buffer
		data map[string]AddrBalance
	)
	addresssUrl := "https://api.omniwallet.org/v2/address/addr/"
	for i, v := range addrs {
		buf.WriteString("addr=")
		buf.WriteString(v)
		if i < (len(addrs) - 1) {
			buf.WriteString("&")
		}
	}
	req, err := http.NewRequest("POST", addresssUrl, strings.NewReader(buf.String()))
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error in sending request to %s. %s", addresssUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		log.Printf("GrabUsdtBalance Couldn't decode response. %s", err)
		return nil, err
	}
	return data, nil
}

//通过address 获取txid vout  value
func GetTxByAddress(address string) (*AddressResult, error) {
	var result AddressResult
	addresssUrl := "https://chain.api.btc.com/v3/address/%s/unspent"
	addresssUrl = fmt.Sprintf(addresssUrl, address)
	req, err := http.NewRequest("GET", addresssUrl, nil)
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error in sending request to %s. %s", addresssUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Printf("GetTxByAddress Couldn't decode response. %s", err)
		return nil, err
	}
	return &result, nil
}

//抓取交易信息, 主要是抓取公钥
func GetPybkeyByTx(txid string) (*TxResult, error) {
	var result TxResult
	txUrl := "https://chain.api.btc.com/v3/tx/%s?verbose=3"
	txUrl = fmt.Sprintf(txUrl, txid)
	req, err := http.NewRequest("GET", txUrl, nil)
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error in sending request to %s. %s", txUrl, err)
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Printf("GetPybkeyByTx Couldn't decode response. %s", err)
		return nil, err
	}
	return &result, nil
}

func getGasInput(input *models.GasInput) (*models.GasOutput, error) {
	var (
		result *models.GasHttpResult
	)

	if input.InNum <= 0 {
		return nil, errors.New(fmt.Sprintf("Error InNum"))
	}
	if input.OutNum <= 0 {
		return nil, errors.New(fmt.Sprintf("Error OutNum"))
	}

	req, err := http.NewRequest("GET", "https://bitcoinfees.earn.com/api/v1/fees/recommended", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result, err = DecodeGasHttpResult(bytes)
	if err != nil {
		return nil, err
	}
	byteNum := int64(input.InNum*148 + 34*input.OutNum + 10)
	gasFee := models.GasFee{
		FastestFee:  byteNum * result.FastestFee,
		HalfHourFee: byteNum * result.HalfHourFee,
		HourFee:     byteNum * result.HourFee,
	}
	suggested := models.Suggested{
		FastestFee:  int64(math.Ceil(float64(byteNum*result.FastestFee)/1000) * 1000),
		HalfHourFee: int64(math.Ceil(float64(byteNum*result.HalfHourFee)/1000) * 1000),
		HourFee:     int64(math.Ceil(float64(byteNum*result.HourFee)/1000) * 1000),
	}

	return &models.GasOutput{
		HttpResult: *result,
		GasFee:     gasFee,
		Suggested:  suggested,
	}, nil

}
