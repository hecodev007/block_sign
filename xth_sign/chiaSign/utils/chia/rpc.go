package chia

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
func NewRpcClient(url, caCertPath, cakeyPath string) *RpcClient {
	pool := x509.NewCertPool()
	//caCertPath := "./private_full_node.crt"

	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return nil
	}
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair(caCertPath, cakeyPath)
	if err != nil {
		panic(err.Error())
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	rpc := &RpcClient{
		client:      client,
		url:         url,
		log:         log.New(os.Stderr, "", log.LstdFlags),
		mutex:       &sync.Mutex{},
		Credentials: "",
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
func (rpc *RpcClient) CallNoAuth(method string, target interface{}, params []byte) error {
	result, err := rpc.call(method, "", params)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}
	return json.Unmarshal(result, target)
}

//需要权限认证的RPC请求。
func (rpc *RpcClient) CallWithAuth(method, credentials string, target interface{}, params []byte) error {
	result, err := rpc.call(method, credentials, params)

	if err != nil {
		fmt.Println("CallWithAuth", method, err.Error())
		return err
	}

	if target == nil {
		return nil
	}
	fmt.Println("CallWithAuth", method)
	fmt.Println("resp", string(result))

	return json.Unmarshal(result, target)
}

// Call returns raw response of method call
func (rpc *RpcClient) call(method, credentials string, body []byte) (json.RawMessage, error) {
	//rpc.log.Infof("rpc.client.Do %v \n", req)
	response, err := rpc.client.Post(rpc.url+method, "application/json", bytes.NewBuffer(body))
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
	log.Println(string(data))
	return data, nil

}

// RawCall returns raw response of method call (Deprecated)
func (rpc *RpcClient) RawCall(method, credentials string, params []byte) (json.RawMessage, error) {
	return rpc.call(method, credentials, params)
}

type StateResponse struct {
	BlockchainState struct {
		Peak struct {
			Height int64 `json:"height"`
		} `json:"peak"`
	} `json:"blockchain_state"`
}

func (rpc *RpcClient) State() (*StateResponse, error) {
	params := "{\"\":\"\"}"
	ret := new(StateResponse)
	err := rpc.CallNoAuth("/get_blockchain_state", ret, []byte(params))
	return ret, err
}

type PublicKeysResponse struct {
	PublicKeyFingerprints []int64 `json:"public_key_fingerprints"`
	Success               bool    `json:"success"`
}

func (rpc *RpcClient) GetPublicKeys() (*PublicKeysResponse, error) {
	params := "{\"\":\"\"}"
	ret := new(PublicKeysResponse)
	err := rpc.CallNoAuth("/get_public_keys", ret, []byte(params))
	return ret, err
}

type GenMonicResponse struct {
	Mnemonic []string `json:"mnemonic"`
	Success  bool     `json:"success"`
}

func (rpc *RpcClient) GenerateMnemonic() (*GenMonicResponse, error) {
	params := "{\"\":\"\"}"
	ret := new(GenMonicResponse)
	err := rpc.CallNoAuth("/generate_mnemonic", ret, []byte(params))
	return ret, err
}

type AddMonicResponse struct {
	Fingerprint int64 `json:"fingerprint"`
	Success     bool  `json:"success"`
}

func (rpc *RpcClient) AddMonic(monic string) (Fingerprint int64, err error) {
	monics := strings.Split(monic, " ")
	params := make(map[string]interface{})
	params["mnemonic"] = monics
	params["type"] = "new_wallet"
	param_str, _ := json.Marshal(params)
	ret := new(AddMonicResponse)
	err = rpc.CallNoAuth("/add_key", ret, param_str)
	if err != nil {
		return 0, err
	}
	if !ret.Success {
		return 0, errors.New("add_key失败")
	}
	return ret.Fingerprint, err
}

func (rpc *RpcClient) Get_private_key(fingerprint int64) {

}

type NextAddressReponse struct {
	Address  string `json:"address"`
	Success  bool   `json:"success"`
	WalletId int64  `json:"wallet_id"`
}

func (rpc *RpcClient) Get_next_address(wallet_id int64) (addr string, err error) {
	params := fmt.Sprintf("{\"wallet_id\": %v, \"new_address\":true}", wallet_id)
	ret := new(NextAddressReponse)
	err = rpc.CallNoAuth("/get_next_address", ret, []byte(params))
	if !ret.Success {
		return "", errors.New("get_next_address失败")
	}
	return ret.Address, nil
}

type LoginResponse struct {
	Fingerprint int64 `json:"fingerprint"`
	Success     bool  `json:"success"`
}

func (rpc *RpcClient) Login(fingerprint int64) (ok bool) {
	params := fmt.Sprintf("{\"fingerprint\": %v, \"type\":\"start\"}", fingerprint)
	ret := new(LoginResponse)
	rpc.CallNoAuth("/log_in", ret, []byte(params))
	return ret.Success
}

type WalletInfoResponse struct {
	Success bool `json:"success"`
	Wallets []struct {
		Data string `json:"data"`
		Id   int64  `json:"id"`
		Name string `json:"name"`
		Type int64  `json:"type"`
	} `json:"wallets"`
}

func (rpc *RpcClient) WalletInfo() (wallets *WalletInfoResponse, err error) {
	params := "{\"\":\"\"}"

	var ret WalletInfoResponse
	err = rpc.CallNoAuth("/get_wallets", &ret, []byte(params))
	if ret.Success == false {
		return nil, errors.New("get_wallets失败")

	}
	if len(ret.Wallets) == 0 {
		return nil, errors.New("get_wallets失败")
	}
	return &ret, err
}

type WalletBalanceResponse struct {
	Success       bool `json:"success"`
	WalletBalance struct {
		Confirmed_wallet_balance   int64 `json:"confirmed_wallet_balance"`
		Max_send_amount            int64 `json:"max_send_amount"`
		Pending_change             int64 `json:"pending_change"`
		Spendable_balance          int64 `json:"spendable_balance"`
		Unconfirmed_wallet_balance int64 `json:"unconfirmed_wallet_balance"`
		Wallet_id                  int64 `json:"wallet_id"`
	} `json:"wallet_balance"`
}

func (rpc *RpcClient) WalletBalance(walletid int64) (ret *WalletBalanceResponse, err error) {
	params := fmt.Sprintf("{\"wallet_id\":\"%v\"}", walletid)
	ret = new(WalletBalanceResponse)
	err = rpc.CallNoAuth("/get_wallet_balance", ret, []byte(params))
	if ret.Success == false {
		return nil, errors.New("get_wallet_balance失败")

	}
	return ret, err
}

type SendTransactionResponse struct {
	Success        bool        `json:"success"`
	Transaction_id string      `json:"transaction_id"`
	Tranasction    Transaction `json:"tranasction"`
	Error          string      `json:"error"`
}
type Transaction struct {
	confirmed_at_index int64  `json:"confirmed_at_index"`
	created_at_time    int64  `json:"created_at_time"`
	to_address         string `json:"to_address"`
	amount             int64  `json:"amount"`
	fee_amount         int64  `json:"fee_amount"`
	incoming           bool   `json:"incoming"`
	confirmed          bool   `json:"confirmed"`
	sent               int64  `json:"sent"`
	//spend_bundle
	additions []Coin `json:"additions"`
	removals  []Coin `json:"removals"`
	wallet_id int64  `json:"wallet_id"`
}
type Coin struct {
	spent_block_index int64 `json:"spent_block_index"`
	spent             bool  `json:"spent"`
	coinbase          bool  `json:"coinbase"`
	//STANDARD_WALLET = 0,
	//RATE_LIMITED = 1,
	//ATOMIC_SWAP = 2,
	//AUTHORIZED_PAYEE = 3,
	//MULTI_SIG = 4,
	//CUSTODY = 5,
	//COLOURED_COIN = 6,
	//RECOVERABLE = 7,
	wallet_type int64 `json:"wallet_type"`
	wallet_id   int64 `json:"wallet_id"`
}

func (rpc *RpcClient) SendTransaction(walletid int64, amount int64, address string, fee int64) (ret *SendTransactionResponse, err error) {
	params := fmt.Sprintf("{\"wallet_id\":\"%v\",\"amount\":%v,\"address\":\"%v\",\"fee\":%v}", walletid, amount, address, fee)
	ret = new(SendTransactionResponse)
	err = rpc.CallNoAuth("/send_transaction", ret, []byte(params))
	if ret.Success == false {
		return nil, errors.New("send_transaction:" + ret.Error)

	}
	return ret, err
}
func (rpc *RpcClient) DelAllKey() (bool, error) {
	params := "{\"\":\"\"}"
	ret := new(GenMonicResponse)
	err := rpc.CallNoAuth("/delete_all_keys", ret, []byte(params))
	return ret.Success, err
}

//func (rpc *RpcClient) GetTransactions(walletid int64) (ret *WalletBalanceResponse, err error) {
//	params := fmt.Sprintf("{\"wallet_id\":\"%v\"}", walletid)
//	ret = new(WalletBalanceResponse)
//	err = rpc.CallNoAuth("/get_transactions", ret, []byte(params))
//	if ret.Success == false {
//		return nil, errors.New("get_transactions失败")
//
//	}
//	return ret, err
//}
