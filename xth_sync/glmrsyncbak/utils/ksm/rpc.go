package ksm

import (
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/onethefour/bifrost-go/client"
	"github.com/onethefour/bifrost-go/models"
	"github.com/shopspring/decimal"
	"time"
)

type RpcClient struct {
	*client.Client
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, node, password string) *RpcClient {
	cli ,err := client.New(url)
	if err != nil {
		panic(err.Error())
	}
	cli.SetPrefix(ss58.KsmPrefix)
	rpc := &RpcClient{
		Client:cli,
	}
	return rpc
}


func (rpc *RpcClient) GetBestHeight() (int64, error) {
	blockhash,err :=  rpc.Client.C.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return 0,err
	}
	bHeader, err := rpc.Client.C.RPC.Chain.GetHeader(blockhash)
	if err != nil {
		return 0,err
	}
	return int64(bHeader.Number),nil

}
func (rpc *RpcClient) GetBlockByNum(h int64) (ret *models.BlockResponse, err error) {
	return rpc.Client.GetBlockByNumber(h)
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
	err =  rpc.Client.C.Client.Call(result, "payment_queryInfo", rawtx, parentHash)
	if err != nil {
		time.Sleep(10 * time.Second)
		goto retry
	}
	return result.PartialFee.Shift(-10).String(), nil
}