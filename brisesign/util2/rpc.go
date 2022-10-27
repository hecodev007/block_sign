package util2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type RpcClient struct {
	rpcUrl      string
	rpcUser     string
	rpcPassword string
}

type RequestBody struct {
	ReqNotHaveParams
	Params []interface{} `json:"params"`
}
type ReqNotHaveParams struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Id      int    `json:"id"`
}
type RespBody struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
}
type RespErrorBody struct {
	JsonRpc string                 `json:"jsonrpc"`
	Error   map[string]interface{} `json:"error"`
	Id      int                    `json:"id"`
}

//初始化一个rpc客户端
func New(url, user, password string) *RpcClient {
	return &RpcClient{
		rpcUrl:      url,
		rpcUser:     user,
		rpcPassword: password,
	}
}

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

//RPC 响应返回数据结构
type Response struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}

//RPC 请求参数数据结构
type Request struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

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

// Call returns raw response of method call
func (rpc *RpcClient) call(method, credentials string, params ...interface{}) (json.RawMessage, error) {
	request := Request{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	log.Printf("NewRequest %v ", request)
	req, err := http.NewRequest("POST", rpc.rpcUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if credentials != "" {
		req.Header.Add("Authorization", "Basic "+credentials)
	}

	log.Printf("rpc.client.Do %v \n", req)
	client := &http.Client{}
	response, err := client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("%s\nResponse: %s\n", method, data))

	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}

	return resp.Result, nil

}

func (rpc *RpcClient) SendRequest(method string, params []interface{}) ([]byte, error) {
	id := rand.Intn(10000)
	var (
		reqBytes []byte
		err      error
	)
	if params != nil {
		var reqBody RequestBody
		reqBody.JsonRpc = "2.0"
		reqBody.Id = id
		reqBody.Method = method
		reqBody.Params = params
		reqBytes, err = json.Marshal(reqBody)
	} else {
		var reqBody ReqNotHaveParams
		reqBody.JsonRpc = "2.0"
		reqBody.Id = id
		reqBody.Method = method
		reqBytes, err = json.Marshal(reqBody)
	}
	if err != nil {
		return nil, err
	}
	reqBuf := bytes.NewBuffer(reqBytes)
	var (
		req *http.Request
	)

	if req, err = http.NewRequest(http.MethodPost, rpc.rpcUrl, reqBuf); err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	//设置rpc的用户和密码
	//如果为空就不设置
	if rpc.rpcUser != "" && rpc.rpcPassword != "" {
		req.SetBasicAuth(rpc.rpcUser, rpc.rpcPassword)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	//解析resp
	var response RespBody
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, errors.New(fmt.Sprintf("Parse resp error,Err=【%v】", err))
	}
	//如果返回的结果直接是一个string，就不在做json处理了，直接返回
	switch response.Result.(type) {
	case string:
		return []byte(response.Result.(string)), nil
	default:
		data, err := json.Marshal(response.Result)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Rpc marshal result error,Err=【%v】,Data=[%s]", err, string(resp)))
		}
		//处理rpc调用失败的情况
		if string(data) == "null" {
			var respError RespErrorBody
			if err := json.Unmarshal(resp, &respError); err != nil {
				return nil, errors.New(fmt.Sprintf("Parse resp error code error,Err=【%v】,Data=[%s]", err, string(resp)))
			}
			rpcErr := respError.Error
			if rpcErr != nil {

				return nil, errors.New(fmt.Sprintf("Rpc get error,Code=【%0.f】,Message=【%s】", rpcErr["code"].(float64), rpcErr["message"].(string)))
			}
			return nil, fmt.Errorf("Rpc response error,can parse data,Data=[%s]", string(resp))
		}
		return data, nil
	}
}
