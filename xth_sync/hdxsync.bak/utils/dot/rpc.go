package dot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	gsClient "github.com/stafiprotocol/go-substrate-rpc-client/client"

	"github.com/shopspring/decimal"
)

type RpcClient struct {
	client *http.Client
	cl     gsClient.Client
	url    string
	node   string
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, node, password string) *RpcClient {
	cl, err := gsClient.Connect(node)
	if err != nil {
		panic(err.Error())
	}
	rpc := &RpcClient{
		client: http.DefaultClient,
		url:    url,
		node:   node,
		cl:     cl,
	}
	return rpc
}

//获取RPC服务URL
func (rpc *RpcClient) URL() string {
	return rpc.url
}

type HeadResponse struct {
	Number decimal.Decimal `json:"number"`
	Hash   string          `json:"hash"`
}

func (rpc *RpcClient) GetBestHeight() (int64, error) {
	return rpc.GetBlockCount()
}
func (rpc *RpcClient) GetBlockCount() (bestBlockCount int64, err error) {
	url := rpc.url + "/blocks/head"
	data, err := rpc.get(url)
	if err != nil {
		return
	}
	response := new(HeadResponse)
	err = json.Unmarshal(data, response)
	if err != nil {
		return
	}
	bestBlockCount = response.Number.IntPart()
	if bestBlockCount == 0 {
		err = errors.New(url + ",内容返回错误:" + string(data))
	}
	return
}

func (rpc *RpcClient) get(url string) (data []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type BlockResponse struct {
	Number         decimal.Decimal `json:"number"`
	Hash           string          `json:"hash"`
	ParentHash     string          `json:"parentHash"`
	StateRoot      string          `json:"stateRoot"`
	ExtrinsicsRoot string          `json:"extrinsicsRoot"`
	AuthorId       string          `json:"authorId"`
	Logs           []*Log          `json:"logs"`
	//OnInitialize OnInitialize `json:"onInitialize"`
	Extrinsics []*Extrinsic `json:"extrinsics"`
	//OnFinalize OnFinalize `json:"onFinalize"`
	Finalized bool `json:"finalized"`
}
type Log struct {
	Tpye  string          `json:"tpye"`
	Index decimal.Decimal `json:"index"`
	Value []string        `json:"value"`
}
type Extrinsic struct {
	Method struct {
		Pallet string `json:"pallet"`
		Method string `json:"method"`
	} `json:"method"`
	Signature struct {
		Signer struct {
			Id string `json:"id"`
		} `json:"signer"`
	} `json:"signature"`
	Nonce string `json:"nonce"`
	Args  struct {
		Dest interface{} `json:"dest"`
		//Dest struct{
		//	Id string `json:"id"`
		//}  `json:"dest"`
		Value decimal.Decimal `json:"value"`
		Calls []struct {
			Method struct {
				Pallet string `json:"pallet"`
				Method string `json:"method"`
			} `json:"method"`
			Args struct {
				Dest struct {
					Id string `json:"id"`
				} `json:"dest"`
				Value decimal.Decimal `json:"value"`
			}
		} `json:"calls"`
	}
	Tip  decimal.Decimal `json:"tip"`
	Hash string          `json:"hash"`
	Info struct {
		Weight     decimal.Decimal `json:"weight"`
		Class      string          `json:"class"`
		PartialFee decimal.Decimal `json:"partialFee"`
	}
	Events []struct {
		Method struct {
			Pallet string `json:"pallet"`
			Method string `json:"method"`
		}
		Data interface{} `json:"data"`
	}
	Success bool `json:"success"`
	PaysFee bool `json:"paysFee"`
}

func (rpc *RpcClient) GetBlock(h int64) (ret *BlockResponse, err error) {
	url := fmt.Sprintf("%v/blocks/%v", rpc.url, h)
	data, err := rpc.get(url)
	if err != nil {
		return
	}
	ret = new(BlockResponse)
	err = json.Unmarshal(data, ret)
	return
}
func (rpc *RpcClient) GetBlockByNum(h int64) (ret *BlockResponse, err error) {
	return rpc.GetBlock(h)
}

type NodeBlock struct {
	Block struct {
		Extrinsics []string `json:"extrinsics"`
	} `json:"block"`
}

func (rpc *RpcClient) GetExtrinsicsByNum(height int64) (Extrinsics []string, err error) {

	var hash string
	err = rpc.cl.Call(&hash, "chain_getBlockHash", height)
	if err != nil {
		return nil, err
	}
	//println(hash)
	block := new(NodeBlock)
	//var block map[string]interface{}
	err = rpc.cl.Call(block, "chain_getBlock", hash)
	if err != nil {
		return
	}
	//for _, v := range block.Block.Extrinsics {
	//	txbytes, _ := hex.DecodeString(strings.Replace(v, "0x", "", 1))
	//	Extrinsics = append(Extrinsics, txbytes)
	//}
	return block.Block.Extrinsics, nil
}

type QueryInfo struct {
	Class      string          `json:"class"`
	PartialFee decimal.Decimal `json:"partialFee"`
	Weight     int64           `json:"weight"`
}

func (rpc *RpcClient) PartialFee(rawtx, parentHash string) (fee string, err error) {
	//println(rawtx)
	//println(parentHash)
	result := new(QueryInfo)
retry:
	err = rpc.cl.Call(result, "payment_queryInfo", rawtx, parentHash)
	if err != nil {
		time.Sleep(10 * time.Second)
		goto retry
	}
	return result.PartialFee.Shift(-12).String(), nil

}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
