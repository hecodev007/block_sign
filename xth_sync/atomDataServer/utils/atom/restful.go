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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/blocks/latest", c.url), nil)
	//println(fmt.Sprintf("%s/blocks/latest", c.url))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, err
	}
	//log.Printf("%s/blocks/latest", c.url)
	//log.Printf("call data: %s", string(data))
	result := new(ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("unmarshal %v", err)
	}
	if result.Block.Header.ChainID != "cosmoshub-4" {
		return 0, errors.New("链已升级到cosmoshub-4=>" + result.Block.Header.ChainID)
	}
	return result.Block.Header.Height.IntPart(),nil
}

func (c *RpcClient) GetBlockByHeight(height int64) (*Block, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/blocks/%d", c.url, height), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}
	result := new(Block)
	err = json.Unmarshal(data, result)
	if err != nil {
		log.Printf("block %v", result)
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	if result.Error != ""{
		return nil,errors.New(result.Error)
	}
	return result, nil
}
func (c *RpcClient) GetRawTransaction(txid string) (*Transaction, error) {
	rtx, err := c.GetTransaction(txid)
	if err != nil {
		return nil, err
	}
	return rtx.ToTx()
}
func (c *RpcClient) GetTransaction(txid string) (*ResponseTx, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/txs/%s", c.url, txid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	//log.Printf("call data: %s", string(data))
	proxy := &ResponseTx{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	if proxy.Error != "" {
		//return nil, errors.New(proxy.Error)
	}
	return proxy, nil
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
