package ksm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
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
	AccountDisplay interface{} `json:"account_display"`
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
		Fee:extrinsic.Fee.Shift(-10).String(),
		Status:extrinsic.Success,
		//To:"",
		//Value:extrinsic
	}
	var params []ExtingArg
	err := json.Unmarshal([]byte(extrinsic.Params),&params)
	if err != nil {
		log.Println(string(extrinsic.Params))
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
			tx.To ,err = HexToAddress(tohex)
			if err != nil {
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
			tx.Value = value_dec.Shift(-10).String()
		default:

			//return nil,errors.New("default 交易解析错误:blockhahs "+tx.BlockHash+" txid:"+tx.Txid)

		}
	}
	return tx,nil
}
func (Api *DotScanApi)Block(blockHash string) (*Block,error){
	resp := new(BlockResponse)
	err := Api.Post("/api/scan/block",map[string]string{"block_hash":blockHash},resp)
	if err != nil {
		return nil, err
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
	return json.Unmarshal(body,ret)
}

var prefix = []byte{0x06}
func HexToAddress(hex string)(addr string,err error){
	return ss58.EncodeByPubHex(hex, prefix)
}