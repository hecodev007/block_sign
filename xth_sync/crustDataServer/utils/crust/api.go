package crust

import (
	"crustDataServer/common/log"
	"encoding/hex"
	"encoding/json"
	"fmt"

	scale "github.com/itering/scale.go"
)

func (rpc *RpcClient) BlockHeight() (int64, error) {
	header, err := rpc.GetHeader()
	if err != nil {
		return 0, err
	}

	return HexToInt(header.Number), nil
}
func (rpc *RpcClient) GetMetadata() (m *scale.MetadataDecoder, err error) {
	res, err := rpc.Call("state_getMetadata")
	if err != nil {
		return nil, err
	}
	var rawMeta string
	err = json.Unmarshal(res, &rawMeta)
	if err != nil {
		log.Info(err.Error())
	}

	meta, err := hex.DecodeString(rawMeta[2:])
	if err != nil {
		return nil, err
	}
	m = &scale.MetadataDecoder{}
	m.Init(meta)
	if err := m.Process(); err != nil {
		return m, err
	}
	if m.Version != "MetadataV11Decoder" {
		return nil, fmt.Errorf("MetadataV11 version should equal 11:%v", m.Version)
	}
	return m, nil
}
func (rpc *RpcClient) BlockHash(height int64) (blockHash string, err error) {
	body, err := rpc.Call("chain_getBlockHash", height)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &blockHash)
	return
}
func (rpc *RpcClient) GetBlockByHeight(height int64) (block *Block, err error) {
	blockHash, err := rpc.BlockHash(height)
	if err != nil {
		return nil, err
	}
	return rpc.GetBlock(blockHash)
}
func (rpc *RpcClient) GetBlock(hash string) (block *Block, err error) {
	body, err := rpc.Call("chain_getBlock", hash)
	if err != nil {
		return nil, err
	}
	block = new(Block)
	err = json.Unmarshal(body, &block)
	if err != nil {
		return nil, err
	}
	block.Hash = hash
	block.Block.Header.Height = HexToInt(block.Block.Header.Number)
	return
}

//获取最新区块header
func (rpc *RpcClient) GetHeader() (ret Header, err error) {
	body, err := rpc.Call("chain_getHeader")
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(body, &ret)
	ret.Height = HexToInt(ret.Number)
	//log.Info(string(body), ret.Height)
	return
}
