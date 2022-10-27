package luna

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/terra-money/core/app"
	"io/ioutil"
	"lunasign/common/log"
	"net/http"
	"strconv"
	"strings"

	//"github.com/kava-labs/rosetta-kava/kava"
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

func (c *NodeClient) AuthBalance(addr string,demon string) (amount int64, err error) {
	if len(demon) >12 {
		return c.TokenBalance(addr,demon)
	}
	if demon == "" {
		demon = "uluna"
	}
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
	println(string(data))
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("unmarshal %v", err)
	}
	if result.Message != "" {
		return 0, errors.New(result.Message)
	}
	//if len(result.Balances) == 0 {
	//	return 0, errors.New("fromAddr not found")
	//}
	//log.Info(xutils.String(result))
	for _,v := range result.Balances{
		//log.Info(v.Denom,demon)
		if v.Denom == demon {
			amount,_ = strconv.ParseInt(v.Amount, 10, 64)
			break
		}
	}
	return amount, nil
}
type TokenBalanceResponse struct {
	Height string `json:"height"`
	Result struct{
		Balance decimal.Decimal `json:"balance"`
	} `json:"result"`
	Error string `json:"error"`
}
func (c *NodeClient) TokenBalance(addr string,token string) (amount int64, err error) {
	url := c.url+"/wasm/contracts/"+token+"/store?query_msg={%22balance%22:{%22address%22:%20%22"+addr+"%22}}"
	req, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer req.Body.Close()
	if req.StatusCode != 200 {
		return 0,errors.New(req.Status)
	}
	//log.Info(  fmt.Sprintf("%v/wasm/contracts/%v/store?query_msg={\"balance\":{\"address\": \"%v\"}}", c.url, token,addr))
	//req.Header.Set("Content-Type", "application/json")
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return 0, err
	}
	//log.Info()
	result := new(TokenBalanceResponse)
	//log.Info(string(data))
	err = json.Unmarshal(data, result)
	if err != nil {
		log.Info(string(data))
		return 0, fmt.Errorf("unmarshal %v", err)
	}
	if result.Error != "" {
		return 0, errors.New(result.Error)
	}
	amount = result.Result.Balance.IntPart()
	return amount, nil
}

func (c *NodeClient) AuthAccount(addr string,demon string) (amount int64, account_number, sequence uint64, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/auth/accounts/%v", c.url, addr), nil)
	if err != nil {
		return 0, 0, 0, err
	}
	//log.Info(fmt.Sprintf("%v/auth/accounts/%v", c.url, addr))

	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return 0, 0, 0, err
	}
	log.Info(string(data))
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

	balance := int64(0)
	//log.Info(xutils.String(result))
	//if len(result.Result.Value.Coins) > 0 {
	//	balance = CoinToInt(result.Result.Value.Coins[0].Denom, result.Result.Value.Coins[0].Amount)
	//}
	balance,err = c.AuthBalance(addr,demon)
	return balance, uint64(accn), uint64(sq), nil
}
func (c *NodeClient) AuthAccountbak(addr string) (amount int64, account_number, sequence uint64, err error) {
	return
	//client, err := kava.NewHTTPClient(c.url)
	//if err != nil {
	//	return 0, 0, 0, err
	//}
	//addrAcc, err := sdk.AccAddressFromBech32(addr)
	//if err != nil {
	//	return 0, 0, 0, err
	//}
	//acc, err := client.Account(context.Background(),addrAcc, 0)
	//if err != nil {
	//	return 0, 0, 0, err
	//}
	//balance := int64(0)
	//
	////if len(acc.GetCoins()) > 0 {
	////	balance = CoinToInt(acc.GetCoins()[0].Denom, acc.GetCoins()[0].Amount.String())
	////}
	//return balance, acc.GetAccountNumber(), acc.GetSequence(), nil

}
func CoinToInt(denom, num string) int64 {
	amount, _ := strconv.ParseInt(num, 10, 64)
	if denom == "kava" {
		return amount * 1e6
	}
	if denom != "kava" && denom != "ukava" {
		return 0
	}
	return amount
}
func (c *NodeClient) call(req *http.Request) (json.RawMessage, error) {
	//rpc.log.Printf("rpc.client.Do %v \n", req)
	num:=0
	redo:
		num++
	response, err := c.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Info(err)
		return nil, err
	}
	if response.StatusCode !=200 && num<3 {
		log.Info(num)
		goto  redo
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func (c *NodeClient) SendRawTransaction(rawtx string)  (txid string, err error) {
	txBytes,err := hex.DecodeString(strings.TrimPrefix(rawtx,"0x"))
	if err != nil {
		return "", err
	}
	broadcastReq := tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	reqBytes, err := json.Marshal(broadcastReq)

	if err != nil {
		return "", errors.New( "failed to marshal")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/cosmos/tx/v1beta1/txs", c.url), bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := c.call(req)
	if err != nil {
		return "", err
	}

	var broadcastTxResponse tx.BroadcastTxResponse
	err = app.MakeEncodingConfig().Marshaler.UnmarshalJSON(data, &broadcastTxResponse)
	if err != nil {
		return "", errors.New( "failed to unmarshal response")
	}

	txResponse := broadcastTxResponse.TxResponse
	if txResponse.Code != 0 {
		return "", fmt.Errorf("tx failed with code %d: %s", txResponse.Code, txResponse.RawLog)
	}

	return txResponse.TxHash, nil
}