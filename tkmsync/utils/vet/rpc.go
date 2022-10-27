package vet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type VetHttpClient struct {
	client *http.Client
	url    string
}

func NewVetHttpClient(url string) *VetHttpClient {
	return &VetHttpClient{
		client: http.DefaultClient,
		url:    url,
	}
}

func (c *VetHttpClient) GetBestHeight() (int64, error) {
	url := fmt.Sprintf("%s/blocks/best", c.url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}
	data, err := c.call(req)
	if err != nil {
		return -1, err
	}

	block := &Block{}
	err = json.Unmarshal(data, block)
	if err != nil {
		return -1, err
	}

	return block.Height, nil
}

func (c *VetHttpClient) GetBlockByHeight(height int64) (*Block, error) {
	url := fmt.Sprintf("%s/blocks/%d", c.url, height)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	block := &Block{}
	err = json.Unmarshal(data, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (c *VetHttpClient) GetBlockByHash(hash string) (*Block, error) {
	url := fmt.Sprintf("%s/blocks/%s", c.url, hash)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	block := &Block{}
	err = json.Unmarshal(data, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (c *VetHttpClient) GetTransaction(txid string) (*Transaction, error) {
	url := fmt.Sprintf("%s/transactions/%s", c.url, txid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{}
	err = json.Unmarshal(data, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (c *VetHttpClient) GetTransactionReceipt(txid string) (*TransactionReceipt, error) {
	url := fmt.Sprintf("%s/transactions/%s/receipt", c.url, txid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}

	txreceipt := &TransactionReceipt{}
	err = json.Unmarshal(data, txreceipt)
	if err != nil {
		return nil, err
	}

	return txreceipt, nil
}

func (c *VetHttpClient) call(req *http.Request) (json.RawMessage, error) {

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
