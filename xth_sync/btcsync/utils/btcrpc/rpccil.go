package btcrpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type RpcClient struct {
	rpchost string //<host|ip:[port]>
	rpcuser string //用户名
	rpcpass string //密码
	client  *http.Client
}

func NewRpcClient(host, user, pass string) *RpcClient {
	return &RpcClient{rpchost: host, rpcuser: user, rpcpass: pass, client: &http.Client{
		Timeout: 60 * time.Second,
	}}
}

type clientRequest struct {
	Method string        `json:"method"` //方法名称
	Params []interface{} `json:"params"` //参数对象
	Id     uint64        `json:"id"`     //id
}

//对请求参数进行编码
func encodeClientRequest(method string, args []interface{}) ([]byte, error) {
	c := &clientRequest{
		Method: method,
		Params: args,
		Id:     uint64(rand.Int63()),
	}
	return json.Marshal(c)
}

//返回响应的byte
func (c *RpcClient) Call(method string, args ...interface{}) ([]byte, error) {
	params := make([]interface{}, 0)
	params = append(params, args...)
	message, err := encodeClientRequest(method, params)
	if err != nil {
		return nil, err
	}
	//log.Info(c.rpchost)
	//log.Info(string(message))
	req, err := http.NewRequest("POST", c.rpchost, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.rpcuser != "" {
		req.SetBasicAuth(c.rpcuser, c.rpcpass)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//if resp.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf("That’s an error.HTTP code:%d,error:%s", resp.StatusCode, string(resp.Body))
	//}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//log.Println("string(bytes):", string(bytes))
	return bytes, err
}

//返回响应的byte
func (c *RpcClient) Calls(method string, params []interface{}) ([]byte, error) {
	message, err := encodeClientRequest(method, params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.rpchost, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.rpcuser, c.rpcpass)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//if resp.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf("That’s an error.HTTP code:%d,error:%s", resp.StatusCode, string(resp.Body))
	//}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//log.Println("string(bytes):", string(bytes))
	return bytes, err
}
