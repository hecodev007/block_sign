package dot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

type RpcClient struct {
	client      *http.Client
	url         string

}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string) *RpcClient {
	rpc := &RpcClient{
		client:      http.DefaultClient,
		url:         url,
	}
	return rpc
}

//获取RPC服务URL
func (rpc *RpcClient) URL() string {
	return rpc.url
}
type HeadResponse struct {
	Number decimal.Decimal `json:"number"`
	Hash string `json:"hash"`
}
func (rpc *RpcClient) GetBestHeight()(int64,error){
	return rpc.GetBlockCount()
}
func (rpc *RpcClient) GetBlockCount() (bestBlockCount int64, err error) {
	url := rpc.url+"/blocks/head"
	data ,err := rpc.get(url)
	if err != nil {
		return
	}
	response := new(HeadResponse)
	err = json.Unmarshal(data,response)
	if err != nil {
		return
	}
	bestBlockCount = response.Number.IntPart()
	if bestBlockCount == 0 {
		err = errors.New(url+",内容返回错误:"+string(data))
	}
	return
}

func (rpc *RpcClient) get(url string)(data []byte,err error){
	resp ,err:=http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type BlockResponse struct{
	Number decimal.Decimal `json:"number"`
	Hash string `json:"hash"`
	ParentHash string `json:"parentHash"`
	StateRoot string `json:"stateRoot"`
	ExtrinsicsRoot string `json:"extrinsicsRoot"`
	AuthorId string `json:"authorId"`
	Logs []*Log `json:"logs"`
	//OnInitialize OnInitialize `json:"onInitialize"`
	Extrinsics []*Extrinsic `json:"extrinsics"`
	//OnFinalize OnFinalize `json:"onFinalize"`
	Finalized bool `json:"finalized"`
}
type Log struct{
	Type string `json:"type"`
	Index decimal.Decimal `json:"index"`
	Value []string `json:"value"`
}
type Extrinsic struct{
	Method struct{
		Pallet string `json:"pallet"`
		Method string `json:"method"`
	} `json:"method"`
	Signature struct{
		Signer string `json:"signer"`
	} `json:"signature"`
	Nonce string `json:"nonce"`
	Args struct{
		Dest interface{} `json:"dest"`
		//Dest struct{
		//	Id string `json:"id"`
		//}  `json:"dest"`
		Value decimal.Decimal `json:"value"`
		Calls []struct{
			Method struct{
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
	Tip decimal.Decimal `json:"tip"`
	Hash string `json:"hash"`
	Info struct{
		Weight decimal.Decimal `json:"weight"`
		Class string `json:"class"`
		PartialFee decimal.Decimal `json:"partialFee"`
	}
	Events []struct{
		Method 	struct{
			Pallet string `json:"pallet"`
			Method string `json:"method"`
		}
		Data interface{} `json:"data"`
	}
	Success bool `json:"success"`
	PaysFee bool `json:"paysFee"`
}
func (rpc *RpcClient)GetBlock(h int64)(ret *BlockResponse,err error){
	url := fmt.Sprintf("%v/blocks/%v",rpc.url,h)
	data,err :=rpc.get(url)
	if err != nil {
		return
	}
	ret = new(BlockResponse)
	err = json.Unmarshal(data,ret)
	return
}
func (rpc *RpcClient)GetBlockByNum(h int64)(ret *BlockResponse,err error){
	return rpc.GetBlock(h)
}