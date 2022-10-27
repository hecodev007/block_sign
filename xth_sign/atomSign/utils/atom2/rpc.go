package atom2

import (
	"atomSign/common/log"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/shopspring/decimal"
)

type RpcClient struct {
	client *http.Client
	url    string
}

func NewRpcClient(url, p, s string) *RpcClient {
	c := &RpcClient{
		client: http.DefaultClient,
		url:    url,
	}
	return c
}

type ResponseSendTx struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int64  `json:"id"`
	Result  struct {
		Code int64  `json:"code"`
		Data string `json:"data"`
		Log  string `json:"log"`
		Hash string `json:"hash"`
	} `json:"result"`
	Error struct {
		Code    int64  `json:"code"`
		Data    string `json:"data"`
		Message string `json:"message"`
	}
}

func (c *RpcClient) SendRawTransaction(rawTx string) (txid string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/broadcast_tx_sync?tx=%v", c.url, rawTx), nil)
	log.Info(fmt.Sprintf("%v/broadcast_tx_sync?tx=%v", c.url, rawTx))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return "", err
	}
	log.Infof("block %v", string(data))
	result := new(ResponseSendTx)
	err = json.Unmarshal(data, result)
	if err != nil {
		return "", fmt.Errorf("unmarshal %v", err)
	}
	if result.Error.Code != 0 {
		return "", errors.New(result.Error.Data)
	}
	if result.Result.Code != 0 {
		return "", errors.New(result.Result.Log)
	}

	return result.Result.Hash, nil
}

func (c *RpcClient) call(req *http.Request) (json.RawMessage, error) {
	//rpc.log.Printf("rpc.client.Do %v \n", req)
	response, err := c.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type NodeClient struct {
	client *http.Client
	url    string
}

func NewNodeClient(url string) *NodeClient {
	c := &NodeClient{
		client: http.DefaultClient,
		url:    url,
	}
	return c
}

type BalanceResponse struct {
	Balances   []*Coin `json:"balances"`
	Pagination struct {
		Total decimal.Decimal `json:"total"`
	}
	Message string `json:"message"`
}
type AuthAccountResponse struct {
	Height string `json:"height"`
	Result struct {
		Value struct {
			Coins         []*Coin `json:"coins"`
			Address       string  `json:"address"`
			AccountNumber string  `json:"account_number"`
			Sequence      string  `json:"sequence"`
		} `json:"value"`
	} `json:"result"`
	Error string `json:"error"`
}
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func (c *NodeClient) AuthBalance(addr string) (amount int64, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/cosmos/bank/v1beta1/balances/%v", c.url, addr), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, err
	}
	result := new(BalanceResponse)
	//println(string(data))
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("unmarshal %v", err)
	}
	if result.Message != "" {
		return 0, errors.New(result.Message)
	}
	if len(result.Balances) == 0 {
		return 0, errors.New("fromAddr not found")
	}

	return CoinToInt(result.Balances[0].Denom, result.Balances[0].Amount), nil
}
func (c *NodeClient) AuthAccount(addr string) (amount int64, account_number, sequence uint64, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/auth/accounts/%v", c.url, addr), nil)
	if err != nil {
		return 0, 0, 0, err
	}
	//println(fmt.Sprintf("%v/auth/accounts/%v", c.url, addr))

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, 0, 0, err
	}
	//println("block %v", string(data))
	result := new(AuthAccountResponse)
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("unmarshal %v", err)
	}
	if result.Error != "" {
		return 0, 0, 0, errors.New(result.Error)
	}
	accn, _ := strconv.ParseInt(result.Result.Value.AccountNumber, 10, 64)
	sq, _ := strconv.ParseInt(result.Result.Value.Sequence, 10, 64)
	//if len(result.Result.Value.Coins) == 0 {
	//	return 0, 0, 0, errors.New("fromAddr not found")
	//}

	balance, err := c.AuthBalance(addr)
	if err != nil {
		return 0, 0, 0, err
	}
	return balance, uint64(accn), uint64(sq), nil
}
func CoinToInt(denom, num string) int64 {
	amount, _ := strconv.ParseInt(num, 10, 64)
	if denom == "atom" {
		return amount * 1e6
	}
	return amount
}
func (c *NodeClient) call(req *http.Request) (json.RawMessage, error) {
	//rpc.log.Printf("rpc.client.Do %v \n", req)
	response, err := c.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
