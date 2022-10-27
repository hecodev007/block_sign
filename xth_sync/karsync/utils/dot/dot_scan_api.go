package ksm

import (
	"bytes"
	"karsync/common/log"
	"encoding/json"
	"errors"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/onethefour/common/xutils"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strings"
)

type DotScanApi struct{
	Url string
	Key string
	Dec int //精度
}

func NewDotScanApi(url string,key string) *DotScanApi{
	return &DotScanApi{
		Url:url,
		Key:key,
		Dec:12,
	}
}
type MetaDataResponse struct{
	Code int64 `json:"code"`
	Message string `json:"message"`
	Data *MetaData `json:"data"`
}
type MetaData struct{
	BlockNum decimal.Decimal `json:"blockNum"`
	NetworkNode string `json:"networkNode"`
	SpecVersion string `json:"specVersion"`
}
func (Api *DotScanApi)MetaData() (*MetaData,error){
	resp := new(MetaDataResponse)
	err := Api.Post("/api/scan/metadata",nil,resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil,errors.New(resp.Message)
	}
	return resp.Data,nil
}
type BlockResponse struct{
	Code int64 `json:"code"`
	Message string `json:"message"`
	Data *Block `json:"data"`
}
type Block struct {
	BlockTimestamp int64 `json:"block_timestamp"`
	BlockNum int64 `json:"block_num"`
	Hash string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Extrinsics []Extrinsic `json:"extrinsics"`
	Events []Event `json:"events"`
}
type Extrinsic struct{
	AccountId string `json:"account_id"`
	ExtrinsicHash string `json:"extrinsic_hash"`
	BlockTimestamp int64 `json:"block_timestamp"`
	BlockNum int64 `json:"block_num"`
	ExtrinsicIndex string `json:"extrinsic_index"`
	CallModule string `json:"call_module"`
	CallModuleFunction string `json:"call_module_function"`
	Params string `json:"params"`
	Nonce int64 `json:"nonce"`
	Success bool `json:"success"`
	Fee decimal.Decimal `json:"fee"`
	FromHex string `json:"from_hex"`
	Finalized bool `json:"finalized"`
	AccountDisplay interface{} `json:"account_display"`
}
type Event struct{
	EventIndex string `json:"event_index"`
	BlockNum int64 `json:"block_num"`
	ExtrinsicIdx int64 `json:"extrinsic_idx"`
	ModuleId string `json:"module_id"`
	EventID        string `json:"event_id"`
	Params         string `json:"params"`
	EventIdx       int64    `json:"event_idx"`
	ExtrinsicHash  string `json:"extrinsic_hash"`
	Finalized      bool   `json:"final6ized"`
	BlockTimestamp int64    `json:"block_timestamp"`
}
type Transaction struct{
	BlockHash string `json:"block_hash"`
	BlockHeight int64 `json:"block_height"`
	Txid string `json:"txid"`
	Nonce int64 `json:"nonce"`
	From string `json:"from"`
	To string `json:"to"`
	Value string `json:"value"`
	Fee string `json:"fee"`
	Status bool `json:"status"`
}

type ExtingArg struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Value interface{}
}


func (block *Block) ToTransaction(i int) (*Transaction,error){
	extrinsic := block.Extrinsics[i]
	if !(extrinsic.CallModule == "balances" &&
		(extrinsic.CallModuleFunction == "transfer" || extrinsic.CallModuleFunction == "transfer_keep_alive")){
		return nil,nil
	}
	tx := &Transaction{
		BlockHash: block.Hash,
		BlockHeight: block.BlockNum,
		Txid:extrinsic.ExtrinsicHash,
		Nonce: extrinsic.Nonce,
		From:extrinsic.AccountId,
		Fee:extrinsic.Fee.Shift(-12).String(),
		Status:extrinsic.Success,
		//To:"",
		//Value:extrinsic
	}
	var params []ExtingArg
	err := json.Unmarshal([]byte(extrinsic.Params),&params)
	if err != nil {
		log.Info(string(extrinsic.Params))
		return nil,err
	}
	for _,v := range params{
		switch v.Name {
		case "dest":
			value,ok := v.Value.(map[string]interface{})
			if !ok {
				return nil,errors.New("dest 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
			}
			Id ,ok:= value["Id"]
			if !ok {
				return nil,errors.New("dest,Id 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
			}
			tohex,ok := Id.(string)
			if !ok {
				return nil,errors.New("dest,tohex 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
			}
			tx.To ,err = HexToAddress(strings.TrimPrefix(tohex,"0x"))
			if err != nil {
				log.Info(tohex,err.Error())
				return nil,err
			}
		case "value":
			value,ok := v.Value.(string)
			if !ok {
				return nil,errors.New("交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
			}
			value_dec ,err :=decimal.NewFromString(value)
			if err != nil {
				return nil,err
			}
			tx.Value = value_dec.Shift(-12).String()
		default:
			//return nil,errors.New("default 交易解析错误:blockhash "+tx.BlockHash+"\ntxid:"+tx.Txid)

		}
	}
	return tx,nil
}

type BatchAllEvent struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Values []BatchAllEventValue `json:"value"`
}

type BatchAllEventValue struct {
	CallModule string `json:"call_module"`
	CallName string `json:"call_name"`
	Params []ExtingArg `json:"params"`
}

type CurrenciesAllEvent struct {
	Name string `json:"name"`
	Type string `json:"type_name"`
	Value interface{} `json:"value"`
}

type CurrenciesAllEventValue struct {
	CallModule string `json:"call_module"`
	CallName string `json:"call_name"`
	Params []ExtingArg `json:"params"`
}
func (block *Block) ToTransactions(i int) (ret []*Transaction,err error){
	extrinsic := block.Extrinsics[i]
	module_function := extrinsic.CallModule+extrinsic.CallModuleFunction
	if module_function == "utilitybatch_all" {
		return block.Utilitybatch_all(i)
	}
	if module_function == "currenciestransfer" {
		return block.Currenciestransfer(i)
	}
	return ret,nil
}

func (block *Block) Currenciestransfer(i int) (ret []*Transaction,err error){

	if block.Extrinsics[i].Success != true{
		return ret,errors.New("失败的交易")
	}
	extrinsic := block.Extrinsics[i]
	module_function := extrinsic.CallModule+extrinsic.CallModuleFunction
	log.Info(module_function)
	modules := []string{"currenciestransfer"}
	if !IsInArrayStr(module_function,modules){
		return ret,nil
	}

	tmptx := &Transaction{
		BlockHash: block.Hash,
		BlockHeight: block.BlockNum,
		Txid:extrinsic.ExtrinsicHash,
		Nonce: extrinsic.Nonce,
		From:extrinsic.AccountId,
		Fee:extrinsic.Fee.Shift(-12).String(),
		Status:extrinsic.Success,
		//To:"",
		//Value:extrinsic
	}


	var params []CurrenciesAllEvent
	err = json.Unmarshal([]byte(extrinsic.Params),&params)
	if err != nil {
		log.Info(extrinsic.Params)
		//log.Info(extrinsic.Params)
		return ret,err
	}


			tx := new(Transaction)
			*tx = *tmptx
			for _,v := range params{
				switch v.Name {
				case "dest":
					value,ok := v.Value.(map[string]interface{})
					if !ok {
						return nil,errors.New("dest 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					Id ,ok:= value["Id"]
					if !ok {
						return nil,errors.New("dest,Id 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					tohex,ok := Id.(string)
					if !ok {
						return nil,errors.New("dest,tohex 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					tx.To ,err = HexToAddress(strings.TrimPrefix(tohex,"0x"))
					if err != nil {
						log.Info(tohex,err.Error())
						return nil,err
					}
					log.Info(tx.To)
				case "amount":
					value,ok := v.Value.(string)
					if !ok {
						return nil,errors.New("交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					value_dec ,err :=decimal.NewFromString(value)
					if err != nil {
						return nil,err
					}
					tx.Value = value_dec.Shift(-12).String()
				case "currency_id":

				default:
					return nil,errors.New("default 交易解析错误:blockhash "+tx.BlockHash+"\ntxid:"+tx.Txid)

				}
			}
			ret = append(ret,tx)
	return ret,nil
}

func (block *Block) Utilitybatch_all(i int) (ret []*Transaction,err error){

	if block.Extrinsics[i].Success != true{
		return ret,errors.New("失败的交易")
	}
	extrinsic := block.Extrinsics[i]
	module_function := extrinsic.CallModule+extrinsic.CallModuleFunction
	log.Info(module_function)
	modules := []string{"utilitybatch_all"}
	if !IsInArrayStr(module_function,modules){
		return ret,nil
	}

	tmptx := &Transaction{
		BlockHash: block.Hash,
		BlockHeight: block.BlockNum,
		Txid:extrinsic.ExtrinsicHash,
		Nonce: extrinsic.Nonce,
		From:extrinsic.AccountId,
		Fee:extrinsic.Fee.Shift(-12).String(),
		Status:extrinsic.Success,
		//To:"",
		//Value:extrinsic
	}
	var params []BatchAllEvent
	err = json.Unmarshal([]byte(extrinsic.Params),&params)
	if err != nil {
		log.Info(extrinsic.Params)
		log.Info(extrinsic.Params)
		return ret,err
	}

	for _,events := range params{
		for _,valuevent := range events.Values{
			log.Info(xutils.String(valuevent))

			if !(valuevent.CallModule == "Balances" && valuevent.CallName=="transfer"){
				continue
			}
			tx := new(Transaction)
			*tx = *tmptx
			for _,v := range valuevent.Params{
				switch v.Name {
				case "dest":
					value,ok := v.Value.(map[string]interface{})
					if !ok {
						return nil,errors.New("dest 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					Id ,ok:= value["Id"]
					if !ok {
						return nil,errors.New("dest,Id 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					tohex,ok := Id.(string)
					if !ok {
						return nil,errors.New("dest,tohex 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					tx.To ,err = HexToAddress(strings.TrimPrefix(tohex,"0x"))
					if err != nil {
						log.Info(tohex,err.Error())
						return nil,err
					}
				case "value":
					value,ok := v.Value.(string)
					if !ok {
						return nil,errors.New("交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)
					}
					value_dec ,err :=decimal.NewFromString(value)
					if err != nil {
						return nil,err
					}
					tx.Value = value_dec.Shift(-12).String()
				default:
					return nil,errors.New("default 交易解析错误:blockhash "+tx.BlockHash+"\ntxid:"+tx.Txid)

				}
			}
			ret = append(ret,tx)
		}

	}
	return ret,nil
}
func (Api *DotScanApi)Block(blockHash string) (*Block,error){
	resp := new(BlockResponse)
	err := Api.Post("/api/scan/block",map[string]string{"block_hash":blockHash},resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil,errors.New(resp.Message)
	}
	if resp.Message == "API rate limit exceeded"{
		return nil,errors.New("API rate limit exceeded")
	}
	return resp.Data,nil
}
func (Api *DotScanApi)Post(path string,params interface{},ret interface{})error{
	header := "X-API-Key: "+Api.Key
	param_data,err := json.Marshal(params)
	if err != nil {
		return err
	}
	resp,err := http.Post(Api.Url+path,header,bytes.NewBuffer(param_data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(body))
	return json.Unmarshal(body,ret)
}
type AccountInfoResponse struct{
	Code int64 `json:"code"`
	Message string `json:"message"`
	Data struct{
		Account struct{
			Balance decimal.Decimal `json:"balance"`
			Nonce int64 `json:"nonce"`
		} `json:"account"`
	} `json:"data"`
}
func (Api *DotScanApi)AccountInfo(address string)(amount int64,nonce int64,err error){
	resp := new(AccountInfoResponse)
	err = Api.Post("/api/scan/search",map[string]string{"key":address},resp)
	if err != nil {
		return 0,0, err
	}
	if resp.Code != 0 {
		return 0,0,errors.New(resp.Message)
	}
	return resp.Data.Account.Balance.Shift(10).IntPart(),resp.Data.Account.Nonce,nil
}
var prefix = []byte{0x00}

func HexToAddress(hex string)(addr string,err error){
	return ss58.EncodeByPubHex(hex, prefix)
}

func IsInArrayStr(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true

		}
	}
	return false
}