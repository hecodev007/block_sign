package rpcclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
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
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}

//RPC 请求参数数据结构
type request struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

//包装的RPC-HTTP 客户端
type RpcClient struct {
	client *http.Client
	url    string
	//log    *log.Logger
	Debug bool
	mutex *sync.Mutex
}

// New create new rpc RpcClient with given url
func NewRpcClient(url string, options ...func(rpc *RpcClient)) *RpcClient {
	rpc := &RpcClient{
		client: http.DefaultClient,
		url:    url,
		//log:    log.New(os.Stderr, "", log.LstdFlags),
		mutex: &sync.Mutex{},
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

//func (rpc *RpcClient) Urlfetch(ctx context.Context, seconds ...int) {
//
//	if len(seconds) > 0 {
//		ctx, _ = context.WithDeadline(
//			ctx,
//			time.Now().Add(time.Duration(1000000000*seconds[0])*time.Second),
//		)
//	}
//
//	rpc.client = urlfetch.Client(ctx)
//}
func (rpc *RpcClient) Call(method string, params, result interface{}) error {
	res, err := rpc.call(method, "", params)
	if err != nil {
		return err
	}
	//fmt.Println("result->", string(res))
	if result == nil {
		return nil
	}

	return json.Unmarshal(res, result)
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

		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
	//fmt.Println("call")
	body, err := json.Marshal(request)
	if err != nil {
		log.Println("marshal")
		return nil, err
	}
	//curl -s --data '{"jsonrpc":"2.0", "method":"condenser_api.get_config", "params":[], "id":1}' http://10.0.230.86:8090
	body1 := strings.Replace(string(body), "null", "", -1)
	fmt.Printf("%+v\n", body1)
	//log.Infof("NewRequest %v ", xutils.String(request))
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer([]byte(body1)))
	if err != nil {
		fmt.Println("new request")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if credentials != "" {
		req.Header.Add("Authorization", "Basic "+credentials)
	}
	//fmt.Println("do")
	//rpc.log.Printf("rpc.client.Do %v \n", req)
	response, err := rpc.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Println("do:", err.Error())
		return nil, err
	}
	//fmt.Println("read all")
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ReadAll:", err)
		return nil, err
	}

	//log.Infof("%v\nResponse: %s", method, data)

	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		fmt.Println("unmarshal:", err)
		return nil, err
	}
	//fmt.Printf("%+v\n", resp)
	if resp.Error != nil {
		fmt.Println("resp.Error:", resp.Error)
		return nil, resp.Error
	}

	return resp.Result, nil

}

// RawCall returns raw response of method call (Deprecated)
func (rpc *RpcClient) RawCall(method, credentials string, params ...interface{}) (json.RawMessage, error) {
	return rpc.call(method, credentials, params...)
}

// Close implements interfaces.CallCloser.
func (t *RpcClient) Close() error {
	return nil
}
