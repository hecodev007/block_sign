package atom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"zcashDataServer/utils"
)

type AtomhttpClient struct {
	client *http.Client
	url    string
}

func NewAtomhttpClient(url string) *AtomhttpClient {
	c := &AtomhttpClient{
		client: http.DefaultClient,
		url:    url,
	}
	return c
}

func (c *AtomhttpClient) GetLatestBlockHeight() (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/blocks/latest", c.url), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, err
	}
	//log.Infof("call data: %s", string(data))
	result := new(ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("unmarshal %v", err)
	}

	return utils.ParseInt64(result.BlockMeta.Header.Height)
}

func (c *AtomhttpClient) GetBlockByHeight(height int64) (*Block, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/blocks/%d", c.url, height), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}
	result := new(ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	//log.Infof("block %v", result)
	return result.toBlock(), nil
}

func (c *AtomhttpClient) GetTransaction(txid string) (*Transaction, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/txs/%s", c.url, txid), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	//log.Infof("call data: %s", string(data))
	proxy := &ResponseTx{}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}

	return proxy.toCetTx()
}

func (c *AtomhttpClient) call(req *http.Request) (json.RawMessage, error) {
	//rpc.log.Infof("rpc.client.Do %v \n", req)
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
