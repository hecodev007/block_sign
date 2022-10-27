package fil

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"filSign/common/log"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
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
	ID      int64           `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}
type GetNonceResponse struct {
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Message string `json:"message"`
	Result  int64  `json:"result"`
	Error   *Error `json:"error"`
}
type GetBalanceResponse struct {
	ID      int64           `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  decimal.Decimal `json:"result"`
	Error   *Error          `json:"error"`
}

//RPC 请求参数数据结构
type request struct {
	ID      int64         `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//包装的RPC-HTTP 客户端
type RpcClient struct {
	client      *http.Client
	url         string
	Debug       bool
	mutex       *sync.Mutex
	Credentials string //访问权限认证的 base58编码
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string, options ...func(rpc *RpcClient)) *RpcClient {
	credentials := ""
	if username != "" {
		credentials = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}
	rpc := &RpcClient{
		client:      http.DefaultClient,
		url:         url,
		mutex:       &sync.Mutex{},
		Credentials: credentials,
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
		return err
	}

	if target == nil {
		return nil
	}
	err = json.Unmarshal(result, target)
	if err != nil {
		log.Info(method, string(result), err.Error())
	}
	return err
}

// Call returns raw response of method call
func (rpc *RpcClient) call(method, credentials string, params ...interface{}) (json.RawMessage, error) {
	if len(params) == 0 {
		params = make([]interface{}, 0)
	}
	request := request{
		ID:      10086,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	//log.Infof(rpc.url+" NewRequest: %v ", string(body))

	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if credentials != "" {
		log.Info(credentials)
		req.Header.Add("Authorization", "Basic "+credentials)
	}

	//log.Printf("rpc.client.Do %v \n", req)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return nil, err
	}

	if response.Status != "200 OK" {
		re, _ := json.Marshal(request)
		log.Infof("NewRequest %v ", string(re))
		log.Info("response code:", response.Status, method, string(body))
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil { //&& len(data)
		log.Infof("NewRequest %v ", request)
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}
	//if method == "Filecoin.SyncState" {
	//println(method, string(data))
	//}
	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, fmt.Errorf("resp: %v , err: %v", string(data), err)
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

func (rpc *RpcClient) SendRawTransaction(msg interface{}) (txid string, err error) {
	ret := struct {
		Txid string `json:"/"`
	}{}
	err = rpc.CallWithAuth("Filecoin.MpoolPush", rpc.Credentials, &ret, msg)
	if err != nil {
		return "", err
	}
	return ret.Txid, nil
}

func (rpc *RpcClient) GetNonce(addr string) (int64, error) {
	params := make([]interface{}, 0)
	params = append(params, addr)
	request := request{
		ID:      10086,
		JSONRPC: "2.0",
		Method:  "Filecoin.MpoolGetNonce",
		Params:  params,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	//log.Printf("rpc.client.Do %v \n", req)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return 0, err
	}

	//if response.Status != "200 OK" {
	re, _ := json.Marshal(request)
	log.Infof("NewRequest %v ", string(re))
	log.Info("response code:", response.Status, string(body))
	//}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil { //&& len(data)
		log.Infof("NewRequest %v ", request)
		return 0, fmt.Errorf("ReadAll err: %v", err)
	}

	log.Info(string(data))
	resp := new(GetNonceResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return 0, fmt.Errorf("resp: %v , err: %v", string(data), err)
	}

	if resp.Error != nil {
		return 0, resp.Error
	}
	if resp.Message != "" {
		return 0, errors.New(resp.Message)
	}
	return resp.Result, nil
}

func (rpc *RpcClient) GetBalance(addr string) (decimal.Decimal, error) {
	params := make([]interface{}, 0)
	params = append(params, addr)
	request := request{
		ID:      10086,
		JSONRPC: "2.0",
		Method:  "Filecoin.WalletBalance",
		Params:  params,
	}
	re, _ := json.Marshal(request)
	log.Infof("NewRequest %v ", string(re))
	body, err := json.Marshal(request)
	if err != nil {
		return decimal.Decimal{}, err
	}
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return decimal.Decimal{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	//log.Printf("rpc.client.Do %v \n", req)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return decimal.Decimal{}, err
	}

	//if response.Status != "200 OK" {

	log.Info("response code:", response.Status, string(body))
	//}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil { //&& len(data)
		log.Infof("NewRequest %v ", request)
		return decimal.Decimal{}, fmt.Errorf("ReadAll err: %v", err)
	}

	log.Info(string(data))
	resp := new(GetBalanceResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return decimal.Decimal{}, fmt.Errorf("resp: %v , err: %v", string(data), err)
	}

	if resp.Error != nil {
		return decimal.Decimal{}, err
	}

	return resp.Result, nil
}
func (rpc *RpcClient) BaseFee() (fee int64, err error) {

	ret := new(SyncState)
	err = rpc.CallWithAuth("Filecoin.SyncState", rpc.Credentials, ret)
	if err != nil {
		return 0, err
	}
	for _, v := range ret.ActiveSyncs {
		if v.Target != nil {
			if len(v.Target.Blocks) > 0 {
				return strconv.ParseInt(v.Target.Blocks[0].ParentBaseFee, 10, 64)

			}
		}
	}
	return 0, errors.New("Filecoin.SyncState返回值出错")
}

type SyncState struct {
	ActiveSyncs []*ActiveSync `json:"ActiveSyncs"`
}

type ActiveSync struct {
	Height int64   `json:"Height"`
	Base   *Base   `json:"Base"`
	Target *Target `json:"Target"`
}
type Target struct {
	Cids   []map[string]string `json:"Cids"`
	Blocks []*BlockHeader      `json:"Blocks"`
}
type Base struct {
	Cids   []map[string]string `json:"Cids"`
	Blocks []*BlockHeader      `json:"Blocks"`
}

type BlockHeader struct {
	Miner         string      `json:"Miner"`
	Ticket        interface{} `json:"Ticket"`
	ElectionProof interface{} `json:"ElectionProof"`
	BeaconEntries interface{} `json:"BeaconEntries"`
	WinPoStProof  interface{} `json:"WinPoStProof"`
	//Parents               []string
	Parents               []map[string]string `json:"Parents"`
	ParentWeight          string              `json:"ParentWeight"`
	Height                int64               `json:"Height"`
	ParentStateRoot       map[string]string   `json:"ParentStateRoot"`
	ParentMessageReceipts map[string]string   `json:"ParentMessageReceipts"`
	Messages              map[string]string   `json:"Messages"`
	BLSAggregate          interface{}         `json:"BLSAggregate"`
	Timestamp             uint64              `json:"Timestamp"`
	BlockSig              interface{}         `json:"BlockSig"`
	ForkSignaling         uint64              `json:"ForkSignaling"`
	ParentBaseFee         string              `json:"ParentBaseFee"`
	validated             bool                `json:"validated"`
	Cid                   string              `json:"Cid"`
}
