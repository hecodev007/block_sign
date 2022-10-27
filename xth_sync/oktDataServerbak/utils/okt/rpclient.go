package okt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdktypes "github.com/okex/exchain-go-sdk/types"

	gosdk "github.com/okex/exchain-go-sdk"
	"github.com/okex/exchain-go-sdk/module/tendermint/types"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

type RpcClient struct {
	client *http.Client
	Scan   string
	Url    string
	gosdk.Client
}

func NewRpcClient(url, user, pwd string) *RpcClient {
	config, err := gosdk.NewClientConfig(url, "exchain-66", gosdk.BroadcastBlock, decimal.NewFromInt(100).Shift(-4).String()+"okt", 200000,
		0, "")
	if err != nil {
		panic(err.Error())
	}

	client := &RpcClient{
		client: http.DefaultClient,
		Url:    url,
		Client: gosdk.NewClient(config)}
	return client
}

//func (r *RpcClient) GetBlockCount()(int64,error){
//	block,err :=r.Tendermint().QueryBlock(-1)
//	if err != nil {
//		return 0,err
//	}
//	return block.Height,nil
//}
func (r *RpcClient) GetBlockByHeight(h int64) (*types.Block, error) {
	return r.Tendermint().QueryBlock(h)
}
func (r *RpcClient) GetBlockResult(h int64) (*types.ResultBlockResults, error) {
	return r.Tendermint().QueryBlockResults(h)
}
func (r *RpcClient) TransactionByHash(txhash string) (*types.ResultTx, *authtypes.StdTx, error) {
	resultx, err := r.Tendermint().QueryTxResult(txhash, false)
	if err != nil {
		return nil, nil, err
	}
	stdtx := new(authtypes.StdTx)

	err = r.Client.Token().(sdktypes.BaseClient).GetCodec().UnmarshalBinaryLengthPrefixed([]byte(resultx.Tx), stdtx)
	return resultx, stdtx, err
}

func (r *RpcClient) GetTxResult(txhash string) (*types.ResultTx, error) {
	return r.Tendermint().QueryTxResult(txhash, false)
}

func (c *RpcClient) GetBlockCount() (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/abci_info?", c.Url), nil)
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
	if result.Result.Response.Data != "OKExChain" {
		return 0, errors.New("链已升级到cosmoshub-3=>" + result.Result.Response.Data)
	}
	return result.Result.Response.LastBlockHeight.IntPart(), nil
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

func (c *RpcClient) GetTxFromScan(txhash string) (string, error) {
	resp, err := http.Get("https://www.oklink.com/api/explorer/v1/okexchain_test/transactions/" + txhash)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", nil
	}
	status := gjson.Get(string(bytes), "data.status").String()
	return status, nil
}
