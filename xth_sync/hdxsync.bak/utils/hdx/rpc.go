package hdx

import (
	subrpc "github.com/itering/substrate-api-rpc/rpc"
	gsClient "github.com/stafiprotocol/go-substrate-rpc-client/client"
	"strconv"
	"strings"
)
type RpcClient struct {
	gsClient.Client
}
func NewRpcClient(url, node, password string) *RpcClient {
	client,err := gsClient.Connect(url)
	if err != nil {
		panic(err.Error())
	}
	return &RpcClient{
		client,
	}
}
func (rpc *RpcClient)GetBestHeight()(h int64,err error){
	headhash,err := rpc.chain_getHead()
	if err != nil {
		return
	}
	blockrsp,err := rpc.chain_getBlock(headhash)
	if err != nil {
		return
	}
	height,err := strconv.ParseUint(strings.Replace(blockrsp.Block.Header.Number,"0x","",1),16,64)
	return int64(height), err
}
func (rpc *RpcClient)GetBlockByHeight(h int64)(block *subrpc.BlockResult,err error){
	blockhash,err :=rpc.chain_getBlockHash(h)
	if err!= nil {
		return
	}
	block,err = rpc.chain_getBlock(blockhash)
	return
}
func (rpc *RpcClient)GetEvents(blockhash string)(rawEvents string,err error){
	err = rpc.Call(&rawEvents,"state_getStorageAt","0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7",blockhash)
	return
}
func (rpc *RpcClient)chain_getRuntimeVersion() (int,error){
	var result subrpc.RuntimeVersion
	err := rpc.Call(&result,"chain_getRuntimeVersion")
	return result.SpecVersion,err
}
func (rpc *RpcClient)state_getMetadata(blockhash string)(meta string,err error){
	err = rpc.Call(&blockhash,"state_getMetadata",blockhash)
	return
}
func (rpc *RpcClient)chain_getHead()( blockhash string,err error){
	err = rpc.Call(&blockhash,"chain_getHead")
	return
}

func (rpc *RpcClient)chain_getBlock(hash string)(block *subrpc.BlockResult,err error){
	block  = new(subrpc.BlockResult)
	err = rpc.Call(&block,"chain_getBlock",hash)
	return
}

func (rpc *RpcClient)chain_getBlockHash(h int64)(blockhash string,err error){
	err = rpc.Call(&blockhash,"chain_getBlockHash",h)
	return
}
