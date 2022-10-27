package btc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"marsDataServer/common/log"
	"io/ioutil"
	"net/http"
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
	client *http.Client
	url    string
	//log         *log.Logger
	Debug       bool
	mutex       *sync.Mutex
	Credentials string //访问权限认证的 base58编码
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string, options ...func(rpc *RpcClient)) *RpcClient {
	credentials := ""
	if username != "" || password != "" {
		credentials = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}
	rpc := &RpcClient{
		client: http.DefaultClient,
		url:    url,
		//log:         log.New(os.Stderr, "", log.LstdFlags),
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
	//	log.Printf("NewRequest %v ", request)
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if credentials != "" {
		req.Header.Add("Authorization", "Basic "+credentials)
	}

	//log.Infof("rpc.client.credentials %v \n", credentials)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err, string(body))
		return nil, err
	}
	if response.Status != "200 OK" {
		log.Infof("NewRequest %v", request)
		log.Info("response code:", response.Status)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil { //&& len(data)
		//log.Info("NewRequest %v ", request)
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}

	//log.Infof(fmt.Sprintf("%s\nResponse: %s\n", method, data))
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
