package atom

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RpcClient struct {
	client *http.Client
	url    string
}

func NewRpcClient(url, a, s string) *RpcClient {
	c := &RpcClient{
		client: http.DefaultClient,
		url:    url,
	}
	return c
}

func (c *RpcClient) GetBlockCount() (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/block", c.url), nil)
	//log.Println(fmt.Sprintf("%s/blocks", c.url))
	if err != nil {
		return 0, err
	}

	//req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, err
	}
	//log.Printf("%s/blocks/latest", c.url)
	//log.Printf("call data: %s", string(data))
	result := new(BlockResponse)
	//log.Println(xutils.String(data))
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("json 解析错误: %v", err)
	}
	//log.Printf(xutils.String(result))
	if result.Result.Block.Header.ChainID != "columbus-5" {
		return 0, errors.New("链已升级到columbus-5=>" + result.Result.Block.Header.ChainID)
	}
	return result.Result.Block.Header.Height.IntPart(), nil
}

func (c *RpcClient) GetBlockByHeight(height int64) (*Block, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/block?height=%d", c.url, height), nil)
	if err != nil {
		return nil, err
	}
	//log.Printf(fmt.Sprintf("%s/block?height=%d", c.url, height))
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}
	result := new(BlockReponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		log.Printf("block %v", string(data))
		return nil, fmt.Errorf("json解码失败: %v", err)
	}
	if result.Error.Message != "" {
		return nil, errors.New(result.Error.Message)
	}
	return &result.Result, nil
}
func (c *RpcClient) GetRawTransaction(txid string) ([]*Transaction, error) {
	rtx, err := c.GetTransaction(txid)
	if err != nil {
		return nil, err
	}
	return rtx.ToTx()
}
func (c *RpcClient) GetTransaction(txid string) (*Result, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tx?hash=0x%v&prove=false", c.url, txid), nil)
	if err != nil {
		return nil, err
	}
	//log.Printf(fmt.Sprintf("%s/tx?hash=0x%v&prove=false", c.url, txid))
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	//log.Printf("call data: %s", string(data))
	result := &TxReponse{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	if result.Error.Message != "" {
		return nil, errors.New(result.Error.Message)
	}
	return &result.Result, nil
}

func (c *RpcClient) call(req *http.Request) (json.RawMessage, error) {
	//rpc.log.Printf("rpc.client.Do %v \n", req)
	response, err := c.client.Do(req)
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

	return data, nil
}
