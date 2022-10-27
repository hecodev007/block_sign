package kava

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io/ioutil"
	"net/http"
	"time"
	"zcashDataServer/common/log"
	"zcashDataServer/utils"
)

func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("kava", "kavapub")
	config.SetBech32PrefixForValidator("kavavaloper", "kavavaloperpub")
	config.Seal()
}

type HttpClient struct {
	client *http.Client
	url    string
	cdc    *codec.Codec
}

func NewHttpClient(url string) *HttpClient {
	c := &HttpClient{
		client: http.DefaultClient,
		url:    url,
		cdc:    makeCodec(),
	}
	return c
}
func (c *HttpClient) GetLastBlockHeight() (int64, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/abci_info", c.url), nil)
	if err != nil {
		return 0, err
	}
	data, err := c.call(req)
	if err != nil {
		return 0, err
	}
	//log.Infof("abci_info: %s", string(data))
	result := new(ResponseInfo)
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("unmarshal %v", err)
	}
	return utils.ParseInt64(result.Response.LastBlockHeight)
}

func (c *HttpClient) GetBlockByHeight(height int64) (*Block, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/block?height=%d", c.url, height), nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		log.Errorf("block %v, err :%v", data, err)
		return nil, err
	}
	//log.Infof("get block : %s", string(data))
	result := new(ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	return result.toBlock(), nil
}
func (c *HttpClient) GetTransactionByHash(txid string, blockTime time.Time) (*Transaction, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tx?hash=0x%s", c.url, txid), nil)
	if err != nil {
		return nil, err
	}
	data, err := c.call(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("call data: %s", string(data))
	proxy := &ResponseTx{
		Timestamp: blockTime,
	}
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	return proxy.toTransaction(c.cdc)
}

type Error struct {
	// A Number that indicates the error type that occurred.
	Code int `json:"code"` /* required */
	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */
	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *Error) Error() string {
	return e.Message
}

//RPC 响应返回数据结构
type Response struct {
	ID      string          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}

func (c *HttpClient) call(req *http.Request) (json.RawMessage, error) {
	response, err := c.client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	resp := new(Response)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}
