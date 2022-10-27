package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type RpcConnConfig struct {
	Host string //<host|ip:[port]>
	User string //用户名
	Pass string //密码
}

//rpc配置接口
type RpcClient struct {
	connConfig *RpcConnConfig
}

//创建一个新实例
func NewRpcClient(connConfig *RpcConnConfig) *RpcClient {
	return &RpcClient{connConfig}
}

// rpc调用
// param args 参数列表
func (c *RpcClient) Call(method string, args ...interface{}) ([]byte, error) {
	url := fmt.Sprintf("http://%s/", c.connConfig.Host)
	params := make([]interface{}, 0)
	params = append(params, args...)
	message, err := encodeClientRequest(method, params)
	//fmt.Println(string(message))
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	//log.Println("URL:", url)
	//log.Println("message:", string(message))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.connConfig.User, c.connConfig.Pass)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error in sending request to %s. %s", url, err)
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Couldn't decode response. %s", err)
		return nil, err
	}
	return bytes, err
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
