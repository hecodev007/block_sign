package wit

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
)

type RpcClient struct {
	Id   int64
	Conn *Conn
}

// New create new rpc RpcClient with given url
func NewRpcClient(url, username, password string) *RpcClient {
	rpc := new(RpcClient)

	rpc.Conn = NewConn(url)
	return rpc
}
func (r *RpcClient) GetBlockByHeight(h int64) (*Block, error) {
	hash, err := r.BlockHash(h)
	if err != nil {
		//log.Println(err.Error())
		return nil, err
	}

	block, err := r.GetBlock(hash)
	if err != nil {
		return nil, err
	}
	block.Height = h
	block.Hash = hash
	return block, nil
}

func (r *RpcClient) GetBlock(blockhash string) (*Block, error) {
	r.Id++
	id := r.Id
	params := Params{
		Id:      id,
		Jsonrpc: "2.0",
		Method:  "getBlock",
		Params:  []interface{}{blockhash, true},
	}
	block := new(Block)
	err := r.RpcCall(params.Id, params, block)
	return block, err
}
func (r *RpcClient) RpcCall(id int64, params interface{}, result interface{}) error {

	resp, err := r.Conn.Call(id, params)
	if err != nil {
		return err
	}
	//println(string(resp))
	rpcret := new(Response)
	err = json.Unmarshal(resp, rpcret)
	if err != nil {
		//println(id, string(resp))
		return err
	}
	if rpcret.Error.Code != 0 {
		return errors.New(rpcret.Error.Message)
	}

	err = json.Unmarshal(rpcret.Result, result)
	if err != nil {
		//println("resuld", id, string(rpcret.Result))
	}
	return err
}
func (r *RpcClient) BlockChain(ephoch, limit int64) (string, error) {
	r.Id++
	id := r.Id
	params := Params{
		Id:      id,
		Jsonrpc: "2.0",
		Method:  "getBlockChain",
		Params:  []interface{}{ephoch, limit},
	}
	resp, err := r.Conn.Call(params.Id, params)
	return string(resp), err
}
func (r *RpcClient) BlockHash(h int64) (string, error) {
	resp, err := r.BlockChain(h, 1)
	if err != nil {
		return "", err
	}
	hash := gjson.Get(resp, "result.0.1").String()
	if hash == "" {
		return "", errors.New(fmt.Sprintf("高度%v的区块获取失败", h))
	}
	return hash, nil
}

func (r *RpcClient) NodeStats() (string, error) {
	r.Id++
	params := Params{
		Id:      r.Id,
		Jsonrpc: "2.0",
		Method:  "nodeStats",
	}
	resp, err := r.Conn.Call(params.Id, params)
	return string(resp), err
}
func (r *RpcClient) GetTransaction(txhash string) (*Transaction, error) {
	r.Id++
	params := Params{
		Id:      r.Id,
		Jsonrpc: "2.0",
		Method:  "getTransaction",
		Params:  []interface{}{txhash},
	}
	tx := new(Transaction)
	err := r.RpcCall(params.Id, params, tx)
	return tx, err
}
func (r *RpcClient) Method(method string) (string, error) {
	r.Id++
	params := Params{
		Id:      r.Id,
		Jsonrpc: "2.0",
		Method:  method,
	}
	resp, err := r.Conn.Call(params.Id, params)
	return string(resp), err
}
func (r *RpcClient) GetBlockCount() (int64, error) {
	resp, err := r.Method("syncStatus")
	if err != nil {
		return 0, err
	}
	h := gjson.Get(resp, "result.current_epoch").Int()

	if h == 0 {
		return 0, errors.New("未知错误")
	}
	return h, nil
}
