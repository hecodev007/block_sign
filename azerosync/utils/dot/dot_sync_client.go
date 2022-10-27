package ksm

import (
	"github.com/shopspring/decimal"
	"time"
)

type RpcClient struct {
	*DotNodeRpc
	*DotScanApi
}

func NewRpcClient(node_url string, scan_api_url string, scan_key string) (*RpcClient, error) {
	//panic(node_url)
	node_rpc, err := NewDotNodeRpc(node_url)
	if err != nil {
		return nil, err
	}
	scan_api := NewDotScanApi(scan_api_url, scan_key)

	return &RpcClient{
		DotNodeRpc: node_rpc,
		DotScanApi: scan_api,
	}, nil
}

func (rpc *RpcClient) GetBestHeight() (int64, error) {
	return rpc.DotNodeRpc.LatestBlock()
}

func (rpc *RpcClient) GetBlockByNum(h int64) (ret *Block, err error) {
	hash, err := rpc.DotNodeRpc.GetBlockHash(h)
	if err != nil {
		return nil, err
	}
	block, err := rpc.DotScanApi.Block(hash)
	if err != nil {
		return nil, err
	}

	return block, err
}

type QueryInfo struct {
	Class      string          `json:"class"`
	PartialFee decimal.Decimal `json:"partialFee"`
	Weight     int64           `json:"weight"`
}

func (rpc *RpcClient) PartialFee(rawtx, parentHash string) (fee string, err error) {
	result := new(QueryInfo)
retry:
	err = rpc.Client.Call(result, "payment_queryInfo", rawtx, parentHash)
	if err != nil {
		time.Sleep(10 * time.Second)
		goto retry
	}
	return result.PartialFee.Shift(-10).String(), nil
}
