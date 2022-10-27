package ksm

import (
	"bncsync/common/log"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/JFJun/go-substrate-crypto/ss58"
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
	AccountDisplay struct{
		Address string `json:"address"`
	} `json:"account_display"`
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
type ExtingParams struct{
	Name string `json:"name"`
	Type string `json:"type"`
	Type_name string `json:"type_name"`
	Value interface{} `json:"value"`
}
func (block *Block) ToTransaction(i int) (tx *Transaction,err error){
	extrinsic := block.Extrinsics[i]
	if extrinsic.Success == false{
		return nil,errors.New("失败的交易")
	}
	//log.Info(xutils.String(extrinsic))
	mod_func := extrinsic.CallModule+extrinsic.CallModuleFunction
	//log.Info(mod_func)
	if mod_func== "currenciestransfer_native_currency" {
		tx = &Transaction{
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
		_=tx
		var paramValues []ExtingParams
		if err := json.Unmarshal([]byte(extrinsic.Params),&paramValues);err != nil {
			log.Info(err.Error())
			return nil, err
		}
		if len(paramValues) !=2 {
			log.Info(len(paramValues))
			return nil, errors.New("交易不确定有没有问题,联系开发")
		}
		for _,v :=range paramValues{
			if v.Type_name == "Address"{
				valueBytes,_ :=json.Marshal(v.Value)
				//log.Info(string(valueBytes))
				valueMap := make(map[string]string,0)
				json.Unmarshal(valueBytes,&valueMap)
				//log.Info(strings.TrimPrefix(valueMap["Id"],"0x"))
				if tx.To,err = HexToAddress(strings.TrimPrefix(valueMap["Id"],"0x"));err != nil {
					log.Info(err.Error())
					return nil, err
				}
				//log.Info(tx.To)
			} else if  v.Type_name == "BalanceOf"{
				value,err := decimal.NewFromString(v.Value.(string))
				if err != nil {
					log.Info(err.Error())
					return nil, err
				}
				tx.Value = value.Shift(-12).String()
				//log.Info(tx.Value)
			} else {
				return nil, errors.New("交易不确定有没有问题,联系开发")
			}
		}
		//log.Info(xutils.String(tx))
		return tx,nil
	} else if mod_func == "balancestransfer" || mod_func == "balancestransfer_keep_alive"{
		tx = &Transaction{
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
			}
		}
		return tx,nil
	}
	return nil, nil
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
	return resp.Data.Account.Balance.Shift(12).IntPart(),resp.Data.Account.Nonce,nil
}
var prefix = []byte{0x06}

func HexToAddress(hex string)(addr string,err error){
	return ss58.EncodeByPubHex(hex, prefix)
}