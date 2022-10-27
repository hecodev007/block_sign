package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/group-coldwalle/coinsign/qieusdtserver/db"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service/usdtfile"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	cfg                    *config.GlobalConfig
	transactionInputSignal int32 = 0
	signInputSignal        int32 = 0
	pushInputSignal        int32 = 0
)

func SetConfig(c *config.GlobalConfig) {
	cfg = c
}

func Gas(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		gasInput  *models.GasInput
		gasOutput *models.GasOutput
		gasInputD []byte
	)

	if gasInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(gasInputD) <= 0 {
		HttpError(w, []byte("no gas_input"), nil)
		return
	}
	if gasInput, err = util.DecodeGasInput(gasInputD); err != nil {
		HttpError(w, []byte("decode gas input error: "+err.Error()), nil)
		return
	}
	gasOutput, err = service.GetGasInput(gasInput)
	if err != nil {
		HttpError(w, []byte("gas error:"+err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"gas": gasOutput})
}

//获取sign in 模板
func GetTxInput(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		txInput  *models.AddrTxin
		txInputD []byte
		txin     *models.TxInputNew
	)
	if txInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(txInputD) <= 0 {
		HttpError(w, []byte("no_input"), nil)
		return
	}
	if txInput, err = util.DecodeAddrTxin(txInputD); err != nil {
		HttpError(w, []byte("decode  input error: "+err.Error()), nil)
		return
	}
	if txInput.Gas == 0 {
		txin, err = util.MakeCollectionTxInputAutoFee(txInput.Addrs, txInput.Toaddr, txInput.Changeaddress)
	} else {
		txin, err = util.MakeCollectionTxInput(txInput.Addrs, txInput.Toaddr, txInput.Changeaddress, txInput.Gas)
	}

	if err != nil {
		HttpError(w, []byte("get input error: "+err.Error()), nil)
		return
	}

	HttpOK(w, map[string]interface{}{"gas": txInput.Gas, "result": txin})
}

//获取sign in 模板
func GetTxInputUseFee(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		txInput  *models.AddrTxinUseFee
		txInputD []byte
	)
	if txInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(txInputD) <= 0 {
		HttpError(w, []byte("no_input"), nil)
		return
	}
	if txInput, err = util.DecodeAddrTxinFee(txInputD); err != nil {
		HttpError(w, []byte("decode  input error: "+err.Error()), nil)
		return
	}

	txin, err := util.MakeCollectionTxInputUseFee(txInput.Addr, txInput.Toaddr, txInput.Changeaddress, txInput.Feeaddress, txInput.Gas)
	if err != nil {
		HttpError(w, []byte("get input error: "+err.Error()), nil)
		return
	}

	HttpOK(w, map[string]interface{}{"gas": txInput.Gas, "result": txin})
}

//获取sign in 模板
func GetSignInput(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		transInput  *models.TxInput
		transInputD []byte
	)
	if transInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(transInputD) <= 0 {
		HttpError(w, []byte("no transcation_input"), nil)
		return
	}
	if transInput, err = util.DecodeTxInput(transInputD); err != nil {
		HttpError(w, []byte("decode SignInput input error: "+err.Error()), nil)
		return
	}
	signInput, err := service.GetSignInput(transInput)
	if err != nil {
		HttpError(w, []byte("get SignInput error: "+err.Error()), nil)
		return
	}
	//HttpOK(w, map[string]interface{}{"raw_tx": signInput.Raw, "txins": signInput.Txins})
	HttpOKData(w, signInput)
}

//获取sign in 模板
func GetSignInputNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var (
		err         error
		transInputs *models.TxInputNew
		transInputD []byte
		signInputs  []*models.SignInput
	)
	if transInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(transInputD) <= 0 {
		HttpError(w, []byte("创建交易错误，缺少传入参数"), nil)
		return
	}
	if transInputs, err = util.DecodeTxInputNew(transInputD); err != nil {
		log.Error("创建交易错误，解析参数错误，请仔细检查传入结构:", err.Error())
		HttpError(w, []byte("创建交易错误，解析参数错误，请仔细检查传入结构"), nil)
		return
	}
	//signInput, err := service.GetSignInput(transInput)
	//fmt.Println("GetSignInputNew")
	//fmt.Println(fmt.Sprintf("%+v", transInputs))
	signInputs, err = service.GetSignInputNew(transInputs)
	if err != nil {
		log.Error("创建交易错误:", err.Error())
		HttpError(w, []byte("创建交易错误: "+err.Error()), nil)
		return
	}

	for i, k := range signInputs {
		jsonByte, _ := json.Marshal(k)
		fmt.Println(string(jsonByte))
		hash := util.Md5HashString(jsonByte)
		signInputs[i].Hash = hash
	}
	//HttpOK(w, map[string]interface{}{"raw_tx": signInput.Raw, "txins": signInput.Txins})
	HttpOKData(w, signInputs)
}

//获取sign in 模板
func GetSignInputNewOne(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var (
		err         error
		transInputs *models.TxInputNew
		transInputD []byte
		signInputs  []*models.SignInput
	)

	if transInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	log.Infof("接收内容：%s", string(transInputD))

	if len(transInputD) <= 0 {
		HttpError(w, []byte("创建交易错误，缺少传入参数"), nil)
		return
	}
	if transInputs, err = util.DecodeTxInputNew(transInputD); err != nil {
		log.Error("创建交易错误，解析参数错误，请仔细检查传入结构:", err.Error())
		HttpError(w, []byte("创建交易错误，解析参数错误，请仔细检查传入结构"), nil)
		return
	}
	//signInput, err := service.GetSignInput(transInput)
	//fmt.Println("GetSignInputNew")
	//fmt.Println(fmt.Sprintf("%+v", transInputs))
	signInputs, err = service.GetSignInputNew(transInputs)
	if err != nil {
		log.Error("创建交易错误:", err.Error())
		HttpError(w, []byte("创建交易错误: "+err.Error()), nil)
		return
	}

	for i, k := range signInputs {
		jsonByte, _ := json.Marshal(k)
		fmt.Println(string(jsonByte))
		hash := util.Md5HashString(jsonByte)
		signInputs[i].Hash = hash
	}
	//HttpOK(w, map[string]interface{}{"raw_tx": signInput.Raw, "txins": signInput.Txins})
	HttpOKData(w, signInputs[0])
}

func Sign(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		signInput  *models.SignInput
		pushInput  *models.PushInput
		signInputD []byte
	)

	if signInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(signInputD) <= 0 {
		log.Error("sign 交易签名错误，缺少需要签名的参数结构")
		HttpError(w, []byte("交易签名错误，缺少需要签名的参数结构"), nil)
		return
	}
	if signInput, err = util.DecodeSignInput(signInputD); err != nil {
		log.Error("sign 交易签名错误，解析结构异常，请检查传入结构:", err.Error())
		HttpError(w, []byte("交易签名错误，解析结构异常，请检查传入结构"), nil)
		return
	}

	pushInput, err = service.SignTranscation(signInput)
	if err != nil {
		log.Error("sign 交易签名错误:", err.Error())
		HttpError(w, []byte("交易签名错误"), nil)
		return
	}
	HttpOKData(w, pushInput)
}

func SignNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	var (
		err        error
		signInputs []*models.SignInput
		signInputD []byte
	)

	if signInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(signInputD) <= 0 {
		log.Error("sign_new 交易签名错误，缺少需要签名的参数结构")
		HttpError(w, []byte("交易签名错误，缺少需要签名的参数结构"), nil)
		return
	}

	if signInputs, err = util.DecodeSignInputNew(signInputD); err != nil {
		log.Error("sign_new 交易签名错误，解析结构异常，请检查传入结构:", err.Error())
		HttpError(w, []byte("交易签名错误，解析结构异常，请检查传入结构"), nil)
		return
	}

	//==============校验hash=====================
	for _, k := range signInputs {
		hash := k.Hash //临时存储hash
		k.Hash = ""
		//反序列化byte
		jsonbyte, _ := json.Marshal(k)
		checkhash := util.Md5HashString(jsonbyte)
		if hash != checkhash {
			fmt.Println("hash校验不通过")
			HttpError(w, []byte("signtx hash error"), nil)
			return
		}
	}
	//==============校验hash=====================

	pushInputs := make([]*models.PushInput, 0)
	for k, v := range signInputs {
		pushInput, err := service.SignTranscation(v)
		if err != nil {
			log.Error("sign_new 错误的下标:"+strconv.Itoa(k)+"，交易签名错误: ", err.Error())
			HttpError(w, []byte("错误的下标:"+strconv.Itoa(k)+"，交易签名错误: "+err.Error()), nil)
			return
		}
		pushInput.MchInfo = v.MchInfo
		pushInputs = append(pushInputs, pushInput)
	}
	HttpOKData(w, pushInputs)
}

func SignNewOne(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	var (
		err        error
		signInput  *models.SignInput
		signInputD []byte
	)

	if signInputD, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(signInputD) <= 0 {
		log.Error("sign_new 交易签名错误，缺少需要签名的参数结构")
		HttpError(w, []byte("交易签名错误，缺少需要签名的参数结构"), nil)
		return
	}

	if signInput, err = util.DecodeSignInputNewOne(signInputD); err != nil {
		log.Error("sign_new 交易签名错误，解析结构异常，请检查传入结构:", err.Error())
		HttpError(w, []byte("交易签名错误，解析结构异常，请检查传入结构"), nil)
		return
	}

	//

	//==============校验hash=====================
	hash := signInput.Hash //临时存储hash
	signInput.Hash = ""
	//反序列化byte
	jsonbyte, _ := json.Marshal(signInput)
	fmt.Println(fmt.Sprintf("signOne 传入data：%s", string(jsonbyte)))
	checkhash := util.Md5HashString(jsonbyte)
	if hash != checkhash {
		fmt.Println("hash校验不通过")
		HttpError(w, []byte("signtx hash error"), nil)
		return
	}
	//==============校验hash=====================
	pushInput, err := service.SignTranscation(signInput)
	if err != nil {
		log.Error("交易签名错误: ", err.Error())
		HttpError(w, []byte("交易签名错误: "+err.Error()), nil)
		return
	}
	pushInput.MchInfo = signInput.MchInfo
	HttpOKData(w, pushInput)
}

//推送交易
func PushTransaction(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	var (
		data      []byte
		err       error
		pushInput *models.PushInput
		trxId     string
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		HttpError(w, []byte("no push_input"), nil)
	}
	log.Println("PushTransaction data:", string(data))
	pushInput, err = util.DecodePushInput(data)
	if err != nil {
		HttpError(w, []byte("decode PushInput error:"+err.Error()), nil)
		return
	}
	trxId, err = service.PushTranscation(pushInput)
	if err != nil {
		HttpError(w, []byte("PushTransaction error"+err.Error()), nil)
		return
	}
	log.Println("PushTransaction txid:", trxId)
	HttpOKData(w, models.PushResult{
		Txid:    trxId,
		MchInfo: pushInput.MchInfo,
	})
}

//多个推送交易
func PushTransactionMore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	var (
		data       []byte
		err        error
		pushInputs []*models.PushInput
		trxId      string
	)
	type pushresult struct {
		code    int
		message string
		data    interface{}
	}

	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		HttpError(w, []byte("no push_input"), nil)
		return
	}
	pushInputs, err = util.DecodePushInputNew(data)
	if err != nil {
		HttpError(w, []byte("decode PushInput error:"+err.Error()), nil)
		return
	}
	if len(pushInputs) <= 0 {
		HttpError(w, []byte("decode pushInputs length is 0 "), nil)
		return
	}

	result := make(map[int]pushresult)
	for k, v := range pushInputs {
		trxId, err = service.PushTranscation(v)
		if err != nil {
			pr := pushresult{
				code:    -1,
				message: err.Error(),
				data:    nil,
			}
			result[k] = pr
		} else {
			pr := pushresult{
				code:    0,
				message: "ok",
				data:    map[string]interface{}{"txid": trxId},
			}
			result[k] = pr
		}
	}
	HttpOKData(w, result)
}

//签名--广播
func TransactionCreate(w http.ResponseWriter, r *http.Request) {
	var (
		data       []byte
		err        error
		transInput *models.TxInput
		signInput  *models.SignInput
		pushInput  *models.PushInput
		trxId      string
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if len(data) <= 0 {
		HttpError(w, []byte("no transcation_input"), nil)
		return
	}
	if transInput, err = util.DecodeTxInput(data); err != nil {
		HttpError(w, []byte("decode transcation input error: "+err.Error()), nil)
		return
	}
	signInput, err = service.GetSignInput(transInput)
	if err != nil {
		HttpError(w, []byte("SignInput input error: "+err.Error()), nil)
		return
	}
	pushInput, err = service.SignTranscation(signInput)
	if err != nil {
		HttpError(w, []byte("Signtranscation input error: "+err.Error()), nil)
		return
	}
	trxId, err = service.PushTranscation(pushInput)
	if err != nil {
		HttpError(w, []byte("PushTransaction error"+err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"transaction_id": trxId})
}

//创建地址,追加进去文件
func CreateAddress(w http.ResponseWriter, r *http.Request) {
	var (
		data []byte
		err  error
		//addressinput  *models.AddressInput
		addressoutput *models.AddressOutPut
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		HttpError(w, []byte("no createaddress_input"), nil)
		return
	}
	//if addressinput, err = util.DecodeCreateAddress(data); err != nil {
	//	HttpError(w, []byte("decode createaddress input error: "+err.Error()), nil)
	//	return
	//}
	if addressoutput, err = service.CreateNewAddress(); err != nil {
		HttpError(w, []byte("createaddress error: "+err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"addressInfo": addressoutput})
}

//批量创建地址
//post
//批量创建私钥并保存到文件
func BatchCreateAddress(w http.ResponseWriter, r *http.Request) {
	var (
		data         []byte
		err          error
		addressinput *models.BatchAddressInput
		usdtAFile    string
		usdtBFile    string
		usdtCFile    string
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		HttpError(w, []byte("no gen num"), nil)
		return
	}

	if addressinput, err = util.DecodeBatchCreateAddress(data); err != nil {
		HttpError(w, []byte("decode batchCreateAddress input error: "+err.Error()), nil)
		return
	}

	if usdtAFile, usdtBFile, usdtCFile, err = usdtfile.GenRylinkFile(addressinput); err != nil {
		HttpError(w, []byte("GenRylinkFile error: "+err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"usdt_a": usdtAFile, "usdt_b": usdtBFile, "usdt_c": usdtCFile})
}

//单线程批量创建地址
//post
//批量创建私钥并保存到文件
func BatchCreateAddressBySingleThread(w http.ResponseWriter, r *http.Request) {
	var (
		data         []byte
		err          error
		addressinput *models.BatchAddressInput
		usdtAFile    string
		usdtBFile    string
		usdtCFile    string
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if len(data) <= 0 {
		addressinput = &models.BatchAddressInput{Num: 1}

	} else {
		if addressinput, err = util.DecodeBatchCreateAddress(data); err != nil {
			HttpError(w, []byte("decode BatchCreateAddressBySingleThread input error: "+err.Error()), nil)
			return
		}
	}
	if usdtAFile, usdtBFile, usdtCFile, err = usdtfile.GenRylinkFileBySingleThread(addressinput); err != nil {
		HttpError(w, []byte("GenRylinkFile error: "+err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"usdt_a": usdtAFile, "usdt_b": usdtBFile, "usdt_c": usdtCFile})
}

//批量生成签名参数
func UploadTransactionInput(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		inputs        []*models.TxInput
		signInputFile string
	)
	if !atomic.CompareAndSwapInt32(&transactionInputSignal, 0, 1) {
		log.Errorln("the last upload file is processing")
		HttpError(w, []byte("the last upload file is processing"), nil)
		return
	}
	defer atomic.StoreInt32(&transactionInputSignal, 0)
	if err = r.ParseMultipartForm(51200000); err != nil {
		log.Errorln("upload file size error: ", err)
		HttpError(w, []byte("upload file size error: "+err.Error()), nil)
		return
	}
	if inputs, err = readTransactionInputFile(r); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if signInputFile, err = writeSignInputFile(inputs); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"sign_input_file": signInputFile})
}

func readTransactionInputFile(r *http.Request) ([]*models.TxInput, error) {
	const name = "tx_input"
	var (
		f      multipart.File
		data   []byte
		err    error
		inputs []*models.TxInput
	)
	if f, _, err = r.FormFile(name); err != nil {
		err = errors.New("pase " + name + " file error: " + err.Error())
		return nil, err
	}
	defer f.Close()
	data, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.New("read " + name + " file error: " + err.Error())
		return nil, err
	}
	if len(data) == 0 {
		err = errors.New("read " + name + " file empty error")
		return nil, err
	}
	if inputs, err = util.DecodeTxInputs(data); err != nil {
		err = errors.New("decode DecodeTransactionInputs error: " + err.Error())
		return nil, err
	}
	if len(inputs) == 0 {
		err = errors.New("DecodeTransactionInputs empty error")
		return nil, err
	}
	return inputs, nil
}

//批量签名
func writeSignInputFile(inputs []*models.TxInput) (string, error) {
	const (
		baseFile = "sign_input_file"
	)
	var (
		signInput     *models.SignInput
		signInputs    []*models.SignInput
		data          []byte
		file          *os.File
		err           error
		signInputFile string
	)
	if len(inputs) == 0 {
		return "", errors.New("writeSignInputFile inputs data empty")
	}
	signInputs = make([]*models.SignInput, 0, 300)
	for _, input := range inputs {
		signInput, err = service.GetSignInput(input)
		if err != nil {
			return "", errors.New("writeSignInputFile GetSignInput error: " + err.Error())
		}
		signInputs = append(signInputs, signInput)
	}
	if len(signInputs) == 0 {
		return "", errors.New("writeSignInputFile GetSignInput empty error")
	}
	if data, err = util.EncodeSignInputs(signInputs); err != nil {
		return "", errors.New("writeSignInputFile EncodeSignInputs error: " + err.Error())
	}
	//写入文件
	signInputFile = path.Join(cfg.GenFilePath, baseFile+time.Now().Format("20060102150405"))
	fi, err := os.Stat(signInputFile)
	if err == nil {
		if fi.IsDir() {
			return "", errors.New("writeSignInputFile error: " + signInputFile + " is dir")
		}
	}
	file, err = os.OpenFile(signInputFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = errors.New("writeSignInputFile Open " + signInputFile + " file error: " + err.Error())
		return "", err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	w.Write(data)
	w.Flush()
	return signInputFile, nil
}

//批量生成广播交易
func UploadSignInput(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		inputs              []*models.SignInput
		pushTransactionFile string
	)
	if !atomic.CompareAndSwapInt32(&signInputSignal, 0, 1) {
		log.Errorln("the last upload file is processing")
		HttpError(w, []byte("the last upload file is processing"), nil)
		return
	}
	defer atomic.StoreInt32(&signInputSignal, 0)
	if err = r.ParseMultipartForm(51200000); err != nil {
		log.Errorln("upload file size error: ", err)
		HttpError(w, []byte("upload file size error: "+err.Error()), nil)
		return
	}
	if inputs, err = readSignInputFile(r); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}

	if pushTransactionFile, err = writePushTransactionFile(inputs); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"push_transaction_file": pushTransactionFile})
}

func readSignInputFile(r *http.Request) ([]*models.SignInput, error) {
	const name = "sign_input"
	var (
		f      multipart.File
		data   []byte
		err    error
		inputs []*models.SignInput
	)
	if f, _, err = r.FormFile(name); err != nil {
		err = errors.New("pase " + name + " file error: " + err.Error())
		return nil, err
	}
	defer f.Close()
	data, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.New("read " + name + " file error: " + err.Error())
		return nil, err
	}
	if len(data) == 0 {
		err = errors.New("read " + name + " file empty error")
		return nil, err
	}
	if inputs, err = util.DecodeSignInputs(data); err != nil {
		err = errors.New("decode DecodeSignInputs error: " + err.Error())
		return nil, err
	}
	if len(inputs) == 0 {
		err = errors.New("DecodeSignInputs empty error")
		return nil, err
	}
	return inputs, nil
}

func writePushTransactionFile(inputs []*models.SignInput) (string, error) {
	const (
		baseFile = "push_transaction_file"
	)
	var (
		pushInput           *models.PushInput
		pushInputs          []*models.PushInput
		data                []byte
		file                *os.File
		err                 error
		pushTransactionFile string
		//result              *models.OmniImportprivkeyResult
	)
	if len(inputs) == 0 {
		return "", errors.New("writePushTransactionFile inputs data empty")
	}
	pushInputs = make([]*models.PushInput, 0, 300)
	for i, input := range inputs {
		//加入自动导入私钥流程-------------------------------------
		//privKey, ok := db.GetPrivKey(input.Txins[i].Address)
		privKey, ok := db.GetPrivKey(input.Txins[i].Address)
		//if !ok {
		//	data = []byte("have not public_key: \""+transInput.SignPubKey+"\" match private key")
		//	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		//	w.Write(data)
		//	return
		//}加载文件。。。
		//导入私钥到客户端
		if !ok {
			return "", errors.New("miss prvkey ")
		}
		//导入私钥到客户端
		//service.Importprivkey(string(privKey))
		_, err = service.Importprivkey(privKey)
		if err != nil {
			return "", errors.New("writePushTransactionFile EncodeSignInputs error: " + err.Error())
		}

		//------------------------------------------------------
		pushInput, err = service.SignTranscation(input)
		if err != nil {
			return "", errors.New("writePushTransactionFile SignTranscation error: " + err.Error())
		}
		pushInputs = append(pushInputs, pushInput)
	}
	if len(pushInputs) == 0 {
		return "", errors.New("writePushTransactionFile SignTranscation empty error")
	}
	if data, err = util.EncodePushInputs(pushInputs); err != nil {
		return "", errors.New("writePushTransactionFile EncodePushInputs error: " + err.Error())
	}
	//写入文件
	pushTransactionFile = path.Join(cfg.GenFilePath, baseFile+time.Now().Format("20060102150405"))
	fi, err := os.Stat(pushTransactionFile)
	if err == nil {
		if fi.IsDir() {
			return "", errors.New("writePushTransactionFile error: " + pushTransactionFile + " is dir")
		}
	}
	file, err = os.OpenFile(pushTransactionFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = errors.New("writePushTransactionFile Open " + pushTransactionFile + " file error: " + err.Error())
		return "", err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	w.Write(data)
	w.Flush()
	return pushTransactionFile, nil
}

//批量生成广播交易
func UploadPushInput(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		inputs         []*models.PushInput
		pushResultFile string
	)
	if !atomic.CompareAndSwapInt32(&pushInputSignal, 0, 1) {
		log.Errorln("the last upload file is processing")
		HttpError(w, []byte("the last upload file is processing"), nil)
		return
	}
	defer atomic.StoreInt32(&pushInputSignal, 0)
	if err = r.ParseMultipartForm(51200000); err != nil {
		log.Errorln("upload file size error, message: ", err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if inputs, err = readPushInputFile(r); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	if pushResultFile, err = writePushResultFile(inputs); err != nil {
		log.Errorln(err)
		HttpError(w, []byte(err.Error()), nil)
		return
	}
	HttpOK(w, map[string]interface{}{"push_result_file": pushResultFile})
}

func readPushInputFile(r *http.Request) ([]*models.PushInput, error) {
	const name = "push_input"
	var (
		f      multipart.File
		data   []byte
		err    error
		inputs []*models.PushInput
	)
	if f, _, err = r.FormFile(name); err != nil {
		err = errors.New("pase " + name + " file error: " + err.Error())
		return nil, err
	}
	defer f.Close()
	data, err = ioutil.ReadAll(f)
	if err != nil {
		err = errors.New("read " + name + " file error: " + err.Error())
		return nil, err
	}
	if len(data) == 0 {
		err = errors.New("read " + name + " file empty error")
		return nil, err
	}
	if inputs, err = util.DecodePushInputs(data); err != nil {
		err = errors.New("decode DecodePushInputs error: " + err.Error())
		return nil, err
	}
	if len(inputs) == 0 {
		err = errors.New("DecodePushInputs empty error")
		return nil, err
	}
	return inputs, nil
}

//批量广播交易并将结果写入文件
func writePushResultFile(inputs []*models.PushInput) (string, error) {
	const (
		baseFile = "push_result_file"
	)
	var (
		//data []byte
		txid           string
		file           *os.File
		err            error
		pushResultFile string
	)
	if len(inputs) == 0 {
		return "", errors.New("writePushResultFile inputs data empty")
	}
	//写入文件
	pushResultFile = path.Join(cfg.GenFilePath, baseFile+time.Now().Format("20060102150405"))
	fi, err := os.Stat(pushResultFile)
	if err == nil {
		if fi.IsDir() {
			return "", errors.New("writePushTransactionFile error: " + pushResultFile + " is dir")
		}
	}
	file, err = os.OpenFile(pushResultFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = errors.New("writePushTransactionFile Open " + pushResultFile + " file error: " + err.Error())
		return "", err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, input := range inputs {

		txid, err = service.PushTranscation(input)
		if err != nil {
			log.Errorln("writePushResultFile PushTranscation error: ", err)
		}
		w.WriteString(input.Hex)
		w.WriteString(": \n")
		w.Write([]byte(txid))
	}
	w.Flush()
	return pushResultFile, nil
}

//导入地址到客户端监控
func ImportAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	type addrs struct {
		Addrs []string `json:"addrs"`
	}
	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		HttpError(w, []byte(err.Error()), nil)
	}
	if len(data) <= 0 {
		log.Errorf("no datas")
		HttpError(w, []byte("no datas"), nil)
	}
	ads := &addrs{}
	err = json.Unmarshal(data, ads)
	if err != nil {
		log.Errorf("decode importAddress error:" + err.Error())
		HttpError(w, []byte("decode importAddress error:"+err.Error()), nil)
		return
	}
	if len(ads.Addrs) <= 0 {
		log.Errorf("empty addrs")
		HttpError(w, []byte("empty addrs"), nil)
	}

	err = service.ImportAddr(ads.Addrs)
	if err != nil {
		log.Errorf("importAddress error::" + err.Error())
		HttpError(w, []byte("importAddress error"+err.Error()), nil)
		return
	}
	HttpOKData(w, nil)
}

//2018年11月23日改版接口
//
//func CreateTransfer(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
//	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
//	w.Header().Set("content-type", "application/json")             //返回数据格式是json
//
//}
