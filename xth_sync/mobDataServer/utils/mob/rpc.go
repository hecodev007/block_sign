package mob

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type ErrorCode int

var ErrNullResult = errors.New("result is null")

type Error struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */
	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */
	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *Error) Error() string {
	return e.Message
}

//RPC 响应返回数据结构
type Response struct {
	ID      string          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}

//RPC 请求参数数据结构
type request struct {
	ID      string        `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//包装的RPC-HTTP 客户端
type RpcClient struct {
	client      *http.Client
	url         string
	log         *log.Logger
	Debug       bool
	mutex       *sync.Mutex
	Credentials string //访问权限认证的 base58编码
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string, options ...func(rpc *RpcClient)) *RpcClient {
	rpc := &RpcClient{
		client:      http.DefaultClient,
		url:         url,
		log:         log.New(os.Stderr, "", log.LstdFlags),
		mutex:       &sync.Mutex{},
		Credentials: "",
	}
	if username != "" {
		rpc.Credentials = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}
	for _, option := range options {
		option(rpc)
	}

	return rpc
}

//获取RPC服务URL
func (rpc *RpcClient) URL() string {
	return rpc.url
}

func (rpc *RpcClient) Urlfetch(ctx context.Context, seconds ...int) {

	if len(seconds) > 0 {
		ctx, _ = context.WithDeadline(
			ctx,
			time.Now().Add(time.Duration(1000000000*seconds[0])*time.Second),
		)
	}

	rpc.client = urlfetch.Client(ctx)
}

//没有权限认证的RPC请求。
func (rpc *RpcClient) CallNoAuth(method string, target interface{}, params ...interface{}) error {
	result, err := rpc.call(method, "", params...)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(result, target)
}

//需要权限认证的RPC请求。
func (rpc *RpcClient) CallWithAuth(method, credentials string, target interface{}, params ...interface{}) error {
	result, err := rpc.call(method, credentials, params...)
	if err != nil {
		fmt.Println("CallWithAuth", method, err.Error())
		return err
	}

	if target == nil {
		return nil
	}
	//fmt.Println("CallWithAuth",method)
	//fmt.Println("resp  ", string(result))
	return json.Unmarshal(result, target)
}

// Call returns raw response of method call
func (rpc *RpcClient) call(method, credentials string, params ...interface{}) (json.RawMessage, error) {
	request := request{
		ID:      "curltest",
		JSONRPC: "1.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("NewRequest %v ", string(body))
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if credentials != "" {
		req.Header.Add("Authorization", "Basic "+credentials)
	}

	//rpc.log.Infof("rpc.client.Do %v \n", req)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}

	//rpc.log.Println(fmt.Sprintf("%s\nResponse: %s\n", method, data))
	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, fmt.Errorf("resp: %v , err: %v", resp, err)
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Result, nil

}

// RawCall returns raw response of method call (Deprecated)
func (rpc *RpcClient) RawCall(method, credentials string, params ...interface{}) (json.RawMessage, error) {
	return rpc.call(method, credentials, params...)
}

type EntropyRep struct{
	Entropy string `json:"entropy"`
}
func (rpc *RpcClient) Entropy() (string,error){
	rep := new(EntropyRep)
	result,err := http.Post(rpc.url+"/entropy","",nil)
	if err != nil {
		return "",err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return "",err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return "", err
	}
	return rep.Entropy,nil
}

func (rpc *RpcClient)GetBlockCount() (int64,error) {
	return rpc.Info()
}
type TxOutResponse struct{
	BlockIndex decimal.Decimal `json:"block_index"`
}
func (rpc *RpcClient)GetBlockIndexByTxid(txhash string)(int64,error){
	rep := new(TxOutResponse)
	result,err := http.Get(rpc.url+"/tx-out/"+txhash+"/block-index")
	if err != nil {
		return 0,err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return 0,err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return 0, errors.New(string(data))
	}
	return rep.BlockIndex.IntPart(),nil
}

type InfoResponse struct{
	BlockCount decimal.Decimal `json:"block_count"`
}

func (rpc *RpcClient) Info() (int64,error){
	rep := new(InfoResponse)
	result,err := http.Get(rpc.url+"/ledger/local")
	//println(rpc.url+"/ledger/local")
	if err != nil {
		return 0,err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return 0,err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return 0, errors.New(string(data))
	}
	return rep.BlockCount.IntPart(),nil
}
type BlockResponse struct {
	Txs []*Txo `json:"tx_outs"`
}
type Txo struct{
	MonitorId string `json:"monitor_id"`
	Subaddress_index int64 `json:"subaddress_index"`
	Public_key string `json:"public_key"`
	Key_image string `json:"key_image"`
	Value decimal.Decimal `json:"value"`
	Direction string `json:"direction"` //received,spent
	Address string `json:"address"`
}
func (rpc *RpcClient) ProcessBlock(monitorid string,blockheight int64)(*BlockResponse,error){
	rep := new(BlockResponse)
	url := fmt.Sprintf("%s/monitors/%s/processed-block/%d",rpc.url,monitorid,blockheight)
	result,err := http.Get(url)
	if err != nil {
		return  nil,err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return nil,err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return nil, errors.New(string(data))
	}
	return rep,nil
}
type GenPriRes struct{
	ViewPrivateKey string `json:"view_private_key"`
	SpendPrivatePey string `json:"spend_private_key"`
}
func (rpc *RpcClient) GenPri(entropy string) (view_pri string,spend_pri string,err error){
	rep := new(GenPriRes)
	result,err := http.Get(rpc.url+"/entropy/"+entropy)
	if err != nil {
		return "","",err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return "","",err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return "","", err
	}
	return rep.ViewPrivateKey,rep.SpendPrivatePey,nil
}
type AddMonitorResponse struct{
	MonitorId string `json:"monitor_id"`
}
type GetMonitorResponse struct{
	First_subaddress int64 `json:"first_subaddress"`
	Num_subaddresses int64 `json:"num_subaddresses"`
	First_block int64 `json:"first_block"`
	Next_block int64 `json:"next_block"`
}
func (rpc *RpcClient)DelMonitor(monitorid string) error{
	//rep := new(GetMonitorResponse)
	//result,err := http.DefaultClient(rpc.url+"/monitors/"+monitorid)

	req, err := http.NewRequest("DELETE", rpc.url+"/monitors/"+monitorid, nil)
	if err != nil{
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
func (rpc *RpcClient)GetMonitor(monitorid string) (first_subaddress,num_subaddresses ,first_block,next_block int64,err error){
	rep := new(GetMonitorResponse)
	result,err := http.Get(rpc.url+"/monitors/"+monitorid)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	println(string(data))
	if err != nil{
		return 0, 0, 0, 0, err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return 0, 0, 0, 0, err
	}
	println(rep.First_subaddress,rep.Num_subaddresses,rep.Next_block,rep.Num_subaddresses)
	return rep.First_subaddress,rep.Num_subaddresses,rep.Next_block,rep.Num_subaddresses,nil
}

func (rpc *RpcClient)AddMonitor(view_private_key string,spend_private_key string,num int64) (string,error){
	params := make(map[string]interface{},0)
	account_key:= make(map[string]interface{},0)
	account_key["view_private_key"] = view_private_key
	account_key["spend_private_key"] = spend_private_key
	params["account_key"] = account_key
	params["first_subaddress"] = 0
	params["num_subaddresses"] = num
	params_json,_ := json.Marshal(params)

	//fmt.Println(string(params_json))
	//fmt.Println(rpc.url+"/monitors")
	rep := new(AddMonitorResponse)
	result,err := http.Post(rpc.url+"/monitors","application/json",bytes.NewBuffer(params_json))
	if err != nil {
		return "",err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return "",err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return "", err
	}
	return rep.MonitorId,nil
}
type GetBalanceResponse struct {
	Balance decimal.Decimal `json:"balance"`
}

func (rpc *RpcClient)GetBalance(monitorid string,num int64) (balance int64 ,err error){
	rep := new(GetBalanceResponse)
	result,err := http.Get(fmt.Sprintf("%s/monitors/%v/subaddresses/%d/balance",rpc.url,monitorid,num))
	if err != nil {
		return 0, err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return  0, err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return 0, err
	}
	return rep.Balance.IntPart(),nil
}

type GetAddressResponse struct{
	iew_public_key string `json:"view_public_key"`
	Spend_public_key string `json:"spend_public_key"`
	B58_address_code string `json:"b58_address_code"`
}
func (rpc *RpcClient)GetAddress(monitorid string,num int64) (address string ,err error){
	rep := new(GetAddressResponse)
	result,err := http.Get(fmt.Sprintf("%s/monitors/%v/subaddresses/%d/public-address",rpc.url,monitorid,num))
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	//fmt.Println(string(data))
	if err != nil{
		return  "", err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return "", err
	}
	return rep.B58_address_code,nil
}
type SendTransactionResponse struct {
	Sender_tx_receipt struct{
		Key_images []string `json:"key_images"`
		Tombstone int64 `json:"tombstone"`

	} `json:"sender_tx_receipt"`
	Receiver_tx_receipt_list []struct{
		Recipient struct{
			View_public_key string `json:"view_public_key"`
			Spend_public_key string `json:"spend_public_key"`
			Fog_report_url string `json:"fog_report_url"`

			Fog_authority_sig string `json:"fog_authority_sig"`
			Fog_report_id string `json:"fog_report_id"`

		} `json:"recipient"`
		Tx_public_key string `json:"tx_public_key"`
		Tx_out_hash string `json:"tx_out_hash"`
		Tombstone int64 `json:"tombstone"`
		Confirmation_number string `json:"confirmation_number"`
	}
}
func (rpc *RpcClient)SendTransaction(fromMonitorid string,value int64,toB58Address string,memo string)(txhash string,err error){
	params := make(map[string]interface{},0)
	params["receiver_b58_address_code"] = toB58Address
	params["value"] = strconv.FormatInt(value,10)
	if memo != "" {
		params["memo"] = memo
	}
	params_json,_ := json.Marshal(params)


	rep := new(SendTransactionResponse)
	url := fmt.Sprintf("%s/monitors/%s/subaddresses/%d/pay-address-code",rpc.url,fromMonitorid,0)
	result,err := http.Post(url,"application/json",bytes.NewBuffer(params_json))
	if err != nil {
		return "",err
	}
	defer result.Body.Close()
	data, err := ioutil.ReadAll(result.Body)
	if err != nil{
		return "",err
	}
	if err = json.Unmarshal(data,rep);err != nil{
		return "", err
	}
	return rep.Receiver_tx_receipt_list[0].Tx_public_key,nil
}
