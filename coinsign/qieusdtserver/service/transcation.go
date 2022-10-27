package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/group-coldwalle/coinsign/qieusdtserver/db"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service/client"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

var (
	defaultTrans = &Transcation{}
)

type Transcation struct {
	RpcClient *client.OmniClient
}

func InitTranscation(cfg *config.GlobalConfig) {
	defaultTrans.RpcClient = client.NewOmniClient(&util.RpcConnConfig{
		Host: cfg.UsdtRpcCfg.Host,
		User: cfg.UsdtRpcCfg.User,
		Pass: cfg.UsdtRpcCfg.Password,
	})
}

func GetGasInput(input *models.GasInput) (*models.GasOutput, error) {
	return defaultTrans.GetGasInput(input)
}

func GetSignInput(input *models.TxInput) (*models.SignInput, error) {
	return defaultTrans.GetSignInput(input)
}

func GetSignInputNew(input *models.TxInputNew) ([]*models.SignInput, error) {
	return defaultTrans.GetSignInputNew(input)
}

func SignTranscation(input *models.SignInput) (*models.PushInput, error) {
	return defaultTrans.SignTranscation(input)
}

func SignTranscation2(input *models.SignInput) (*models.PushInput, error) {
	return defaultTrans.SignTranscation2(input)
}

func PushTranscation(input *models.PushInput) (string, error) {
	return defaultTrans.PushTranscation(input)
}

func CreateNewAddress() (*models.AddressOutPut, error) {
	return defaultTrans.CreateNewAddress()
}

func ImportAddr(addrs []string) error {
	return defaultTrans.ImportAddr(addrs)
}

func Importprivkey(privkey string) (*models.OmniImportprivkeyResult, error) {
	return defaultTrans.RpcImportprivkey(privkey)
}

func DumpPrivKey(addr string) (*models.OmniDumpprivkeyResult, error) {
	return defaultTrans.DumpPrivKey(addr)
}

func (t *Transcation) GetSignInput(input *models.TxInput) (*models.SignInput, error) {

	btcStr := ""
	//默认btc接收值
	var err error
	//参数校验
	if input.Fee.IsZero() {
		return nil, errors.New("Error:Insufficient fee ")
	}
	feeF, _ := input.Fee.Float64()
	if feeF > 0.1 {
		return nil, errors.New("Error: The fee is too large ")
	}
	if input.ChangeAddress == "" {
		return nil, errors.New("Error:Missing changeAddress")
	}
	if input.ToAddress == "" {
		return nil, errors.New("Error:Missing toAddress")
	}
	if input.ToAmount.IsZero() {
		return nil, errors.New("Error:Missing toAmount")
	}
	if input.ToBtc.IsZero() {
		btcStr = ""
	} else {
		if input.ToBtc.LessThan(decimal.NewFromFloat(models.DEFAULT_SEND_BTC)) {
			return nil, errors.New("Error:toBtc min 0.00000546")
		}
		if input.ToBtc.GreaterThanOrEqual(decimal.NewFromFloat(models.DEFAULT_MAX_BTC)) {
			return nil, errors.New("Error:toBtc max 0.01")
		}
		btcStr = input.ToBtc.String()
	}

	if len(input.Txins) == 0 {
		return nil, errors.New("Error:Missing unspents")
	}
	//addr := input.Txins[0].Address
	for i, v := range input.Txins {
		if v.Txid == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss txid", i))
		}
		if v.Vout < 0 {
			return nil, errors.New(fmt.Sprintf("index %d, Error vout", i))
		}
		if v.ScriptPubKey == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss scriptPubKey", i))
		}
		if v.Amount.IsZero() {
			return nil, errors.New(fmt.Sprintf("index %d, Error amount", i))
		}
		if v.Address == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss address", i))
		}
	}

	//流程 前提步骤 需要知道指定地址的utxo btc命令 listunspent 查询地址http://chainquery.com/bitcoin-api/listunspent
	//step1： API:omni_createpayload_simplesend， 创建一个简单的发送模板，指定代币类型和代币数量
	//step2： API:createrawtransaction， 构建交易基本类型 ，指定从哪个源地址（UTXO中txid和vout）转移比特币,参考btc createrawtransaction
	//step3： API:omni_createrawtx_opreturn 在交易数据中加上omni代币数据 参数对应 step2 和 step1的数据
	//step4： API:omni_createrawtx_reference 在交易数据上加上接收地址
	//step5： API:omni_createrawtx_change 在交易数据上指定矿工费用,指定矿工费用和UTXO数据（txid、vout、scriptPubkey、value)

	//step1
	simpleSendRaw, err := t.RpcClient.GetSimpleSend(31, input.ToAmount.String())
	if err != nil {
		return nil, err
	}
	//fmt.Println("simpleSendRaw:", simpleSendRaw)
	changedata := models.OmniChangeTx{} //用于step5
	//step2
	ocx := models.OmniCreateTx{}
	for _, v := range input.Txins {
		simpletxin := models.OmniSimpleTxin{
			Txid: v.Txid,
			Vout: v.Vout,
		}
		ocx.TxIns = append(ocx.TxIns, simpletxin)

		am, _ := v.Amount.Float64()
		changePrvtx := models.OmniChangeTxPrevtxs{
			Txid:         v.Txid,
			Vout:         v.Vout,
			ScriptPubKey: v.ScriptPubKey,
			Value:        am,
		}
		changedata.Prevtxs = append(changedata.Prevtxs, changePrvtx)

	}
	createtxRaw, err := t.RpcClient.GetCreateTx(ocx)
	if err != nil {
		return nil, err
	}
	//fmt.Println("createtxRaw:", createtxRaw)

	//step3
	opreturnResult, err := t.RpcClient.GetOpreturnTx(simpleSendRaw.Result, createtxRaw.Result)
	if err != nil {
		return nil, err
	}
	//fmt.Println("opreturnResult:", opreturnResult)
	//step4
	referenceResult, err := t.RpcClient.GetReferenceTx(opreturnResult.Result, input.ToAddress, btcStr)

	if err != nil {
		return nil, err
	}

	//fmt.Println("referenceResult:", referenceResult)

	//step5
	fee, _ := input.Fee.Float64()
	changedata.Rawtx = referenceResult.Result
	changedata.Fee = fee                         //矿工费
	changedata.Destination = input.ChangeAddress //找零地址
	changeResult, err := t.RpcClient.GetChangeTx(changedata)
	if err != nil {
		return nil, err
	}
	//fmt.Println("changeResult:", changeResult)
	return &models.SignInput{Raw: changeResult.Result, Txins: input.Txins}, nil
}

//新版签名结构生成（支持批量）
func (t *Transcation) GetSignInputNew(txinputs *models.TxInputNew) ([]*models.SignInput, error) {

	//参数校验
	for _, input := range txinputs.Txinputs {
		if input.Fee.IsZero() {
			return nil, errors.New("Error:Insufficient fee ")
		}
		if input.ChangeAddress == "" {
			return nil, errors.New("Error:Missing changeAddress")
		}
		if input.ToAddress == "" {
			return nil, errors.New("Error:Missing toAddress")
		}
		if input.ToAmount.IsZero() {
			return nil, errors.New("Error:Missing toAmount")
		}
		if len(input.Txins) == 0 {
			return nil, errors.New("Error:Missing unspents")
		}
		//address := input.Txins[0].Address
		for i, v := range input.Txins {
			//if v.Address != address {
			//	return nil, errors.New(fmt.Sprintf("index %d,Address Must be consistent", i))
			//}
			if v.Txid == "" {
				return nil, errors.New(fmt.Sprintf("index %d, Miss txid", i))
			}
			if v.Vout < 0 {
				return nil, errors.New(fmt.Sprintf("index %d, Error vout", i))
			}
			if v.ScriptPubKey == "" {
				return nil, errors.New(fmt.Sprintf("index %d, Miss scriptPubKey", i))
			}
			if v.Amount.IsZero() {
				return nil, errors.New(fmt.Sprintf("index %d, Error amount", i))
			}
		}

	}
	signInputs := make([]*models.SignInput, 0)
	for _, txin := range txinputs.Txinputs {
		if si, err := t.CreateTx(txin, txin.MchInfo); err == nil {
			signInputs = append(signInputs, si)
		} else {

			return nil, errors.New(fmt.Sprintf("Error tx : %+v", err))
		}
	}
	return signInputs, nil
}

//签名
func (t *Transcation) SignTranscation(input *models.SignInput) (*models.PushInput, error) {
	if input.Raw == "" {
		return nil, errors.New("Error:Missing Raw ")
	}

	data := models.OmniSigntx{
		Rawtx: input.Raw,
	}
	for i, v := range input.Txins {
		if v.Txid == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss txid", i))
		}
		if v.Vout < 0 {
			return nil, errors.New(fmt.Sprintf("index %d, Error vout", i))
		}
		if v.ScriptPubKey == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss scriptPubKey", i))
		}
		if v.Address == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss address", i))
		}

		privkey, _ := db.GetPrivKey(v.Address)
		if privkey == "" {
			return nil, errors.New(fmt.Sprintf("Misss PrivKey:%s", v.Address))
		}
		//导入地址私钥
		_, err := t.RpcClient.RpcImportprivkey(privkey)
		if err != nil {
			return nil, err
		}
		if v.Amount.IsZero() {
			return nil, errors.New(fmt.Sprintf("index %d, Error amount", i))
		}
		am, _ := v.Amount.Float64()
		data.Prevtxs = append(data.Prevtxs, models.OmniSigntxPrevtxs{
			Txid:         v.Txid,
			Vout:         v.Vout,
			ScriptPubKey: v.ScriptPubKey,
			Value:        am,
			RedeemScript: v.RedeemScript,
		})
	}

	result, err := t.RpcClient.GetSignTx(data)
	if err != nil {
		return nil, err
	}
	return &models.PushInput{Hex: result.Result.Hex, Complete: result.Result.Complete, Error: result.Result.Errors}, nil
}

//签名
func (t *Transcation) SignTranscation2(input *models.SignInput) (*models.PushInput, error) {
	if input.Raw == "" {
		return nil, errors.New("Error:Missing Raw ")
	}
	data := models.OmniSigntx{
		Rawtx: input.Raw,
	}
	for i, v := range input.Txins {
		if v.Txid == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss txid", i))
		}
		if v.Vout < 0 {
			return nil, errors.New(fmt.Sprintf("index %d, Error vout", i))
		}
		if v.ScriptPubKey == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss scriptPubKey", i))
		}
		if v.Address == "" {
			return nil, errors.New(fmt.Sprintf("index %d, Miss address", i))
		}
		if v.Amount.IsZero() {
			return nil, errors.New(fmt.Sprintf("index %d, Error amount", i))
		}
		am, _ := v.Amount.Float64()
		fmt.Println("am3:", am)
		data.Prevtxs = append(data.Prevtxs, models.OmniSigntxPrevtxs{
			Txid:         v.Txid,
			Vout:         v.Vout,
			ScriptPubKey: v.ScriptPubKey,
			Value:        am,
			RedeemScript: v.RedeemScript,
		})
	}

	result, err := t.RpcClient.GetSignTx2(data)
	if err != nil {
		return nil, err
	}
	return &models.PushInput{Hex: result.Result.Hex, Complete: result.Result.Complete, Error: result.Result.Errors}, nil
}

//广播交易
func (t *Transcation) PushTranscation(input *models.PushInput) (string, error) {
	result, err := t.RpcClient.PushTransaction(input.Hex)
	if err != nil {
		log.Errorln("execute pushTransaction error, input:   ", input, "   error message: ", err)
		return "", err
	}
	return result.Result, nil
}

func (t *Transcation) GetGasInput(input *models.GasInput) (*models.GasOutput, error) {
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
	result, err = util.DecodeGasHttpResult(bytes)
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

//生成地址，同时返回私钥 公钥 脚本公钥
func (t *Transcation) CreateNewAddress() (*models.AddressOutPut, error) {
	result, err := t.RpcClient.GetNewAddress()

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Transcation) ImportAddr(ads []string) error {
	if len(ads) > 0 {
		go func(addrs []string) {
			errs := make([]string, 0)
			for _, v := range addrs {
				_, err := t.RpcClient.RpcImportAddrs(v, "hoo", false)
				if err != nil {
					errs = append(errs, v)
					fmt.Println("导入错误，err:", err)
					continue
				}
				fmt.Println("导入成功:", v)
			}
			if len(errs) > 0 {
				arrdata, _ := json.Marshal(errs)
				log.Println(time.Now().String()+"导入错误数量", len(errs), string(arrdata))
			}
			log.Println("完成任务")
		}(ads)
		return nil
	} else {
		return errors.New("empty addrs")
	}
}

//导出私钥
func (t *Transcation) DumpPrivKey(input string) (*models.OmniDumpprivkeyResult, error) {
	result, err := t.RpcClient.Dumpprivkey(input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//导入私钥
func (t *Transcation) RpcImportprivkey(privkey string) (*models.OmniImportprivkeyResult, error) {
	result, err := t.RpcClient.RpcImportprivkey(privkey)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Transcation) CreateTx(input models.TxInput, mchinfo models.MchInfo) (*models.SignInput, error) {
	var btcamount string
	var err error
	if input.ToBtc.IsZero() {
		btcamount = ""
	} else {
		if input.ToBtc.LessThan(decimal.NewFromFloat(models.DEFAULT_SEND_BTC)) {
			return nil, errors.New("Error:toBtc min 0.00000546")
		}
		if input.ToBtc.GreaterThanOrEqual(decimal.NewFromFloat(models.DEFAULT_MAX_BTC)) {
			return nil, errors.New("Error:toBtc max 0.01")
		}
		btcamount = input.ToBtc.String()

	}
	//流程 前提步骤 需要知道指定地址的utxo btc命令 listunspent 查询地址http://chainquery.com/bitcoin-api/listunspent
	//step1： API:omni_createpayload_simplesend， 创建一个简单的发送模板，指定代币类型和代币数量
	//step2： API:createrawtransaction， 构建交易基本类型 ，指定从哪个源地址（UTXO中txid和vout）转移比特币,参考btc createrawtransaction
	//step3： API:omni_createrawtx_opreturn 在交易数据中加上omni代币数据 参数对应 step2 和 step1的数据
	//step4： API:omni_createrawtx_reference 在交易数据上加上接收地址
	//step5： API:omni_createrawtx_change 在交易数据上指定矿工费用,指定矿工费用和UTXO数据（txid、vout、scriptPubkey、value)
	//step1
	simpleSendRaw, err := t.RpcClient.GetSimpleSend(31, input.ToAmount.String())
	if err != nil {
		return nil, err
	}
	//fmt.Println("simpleSendRaw:", simpleSendRaw)
	changedata := models.OmniChangeTx{} //用于step5
	//step2
	ocx := models.OmniCreateTx{}
	for _, v := range input.Txins {
		simpletxin := models.OmniSimpleTxin{
			Txid: v.Txid,
			Vout: v.Vout,
		}
		ocx.TxIns = append(ocx.TxIns, simpletxin)

		am, _ := v.Amount.Float64()
		changePrvtx := models.OmniChangeTxPrevtxs{
			Txid:         v.Txid,
			Vout:         v.Vout,
			ScriptPubKey: v.ScriptPubKey,
			Value:        am,
		}
		changedata.Prevtxs = append(changedata.Prevtxs, changePrvtx)

	}
	createtxRaw, err := t.RpcClient.GetCreateTx(ocx)
	if err != nil {
		return nil, err
	}
	//fmt.Println("createtxRaw:", createtxRaw)

	//step3 组合数据
	opreturnResult, err := t.RpcClient.GetOpreturnTx(simpleSendRaw.Result, createtxRaw.Result)
	if err != nil {
		return nil, err
	}
	//fmt.Println("opreturnResult:", opreturnResult)
	//step4 设置接受者信息
	referenceResult, err := t.RpcClient.GetReferenceTx(opreturnResult.Result, input.ToAddress, btcamount)
	if err != nil {
		return nil, err
	}
	//step5
	fee, _ := input.Fee.Float64()
	//fmt.Println("fee:", fee)
	changedata.Rawtx = referenceResult.Result
	changedata.Fee = fee                         //矿工费
	changedata.Destination = input.ChangeAddress //找零地址
	changeResult, err := t.RpcClient.GetChangeTx(changedata)
	if err != nil {
		return nil, err
	}
	return &models.SignInput{Raw: changeResult.Result, Txins: input.Txins, MchInfo: mchinfo}, nil
}
