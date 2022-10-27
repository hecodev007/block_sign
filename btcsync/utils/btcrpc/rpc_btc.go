package btcrpc

import (
	"btcsync/common/log"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/onethefour/common/xutils"
)

//获取节点信息
func (c *RpcClient) BtcGetChainInfo() (*BtcChaininfo, error) {
	var (
		datas []byte
		info  *BtcChaininfo
		err   error
	)
	datas, err = c.Call("getblockchaininfo")
	if err != nil {
		return nil, err
	}
	info, err = decodeBtcChaininfo(datas)
	if err != nil {
		return nil, err
	}
	return info, err
}

func (c *RpcClient) BtcGetBlockCount() (*BtcBlockCountInfo, error) {
	var (
		datas []byte
		info  *BtcBlockCountInfo
		err   error
	)
	datas, err = c.Call("getblockcount")
	if err != nil {
		return nil, err
	}
	info, err = decodeBtcBlockCountInfo(datas)
	if err != nil {
		log.Info(string(datas))
		return nil, err
	}
	return info, err
}
func (c *RpcClient) GetBlockByHeight(height int64) (*BtcBlockInfo, error) {
	hashInfo, err := c.BtcGetBlockHash(height)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	if hashInfo.Result == "" {
		log.Info(fmt.Sprintf("%v hash为空", height))
		return nil, errors.New(fmt.Sprintf("%v hash为空", height))
	}
	ret, err := c.BtcGetblock(hashInfo.Result)
	if err != nil {
		log.Info(hashInfo.Result + err.Error())
		return nil, err
	}
	if ret.Error != nil {
		log.Info(xutils.String(ret.Error))
		return nil, errors.New(xutils.String(ret.Error))
	}
	return ret.Result, nil
}

//根据高度获取区块hash
func (c *RpcClient) BtcGetBlockHash(index int64) (*BtcGetBlockHash, error) {
	var (
		datas []byte
		tx    *BtcGetBlockHash
		err   error
	)
	datas, err = c.Call("getblockhash", index)
	if err != nil {
		return nil, err
	}
	tx, err = decodeBtcGetBlockHash(datas)
	if err != nil {
		return nil, err
	}
	return tx, err
}

//根据块hash获取区块信息
func (c *RpcClient) BtcGetblock(hash string) (*BtcBlock, error) {
	var (
		datas []byte
		block *BtcBlock
		err   error
	)
	datas, err = c.Call("getblock", hash, 2)
	if err != nil {
		return nil, err
	}
	block, err = decodeBtcBlock(datas)
	if err != nil {
		return nil, err
	}
	return block, err
}

//根据块hash获取区块信息
func (c *RpcClient) BtcGetblock1(hash string) (*BtcBlock1, error) {
	var (
		datas []byte
		//block *BtcBlock1
		err error
	)
	datas, err = c.Call("getblock", hash, 1)
	if err != nil {
		return nil, err
	}
	ri := &BtcBlock1{}
	err = json.Unmarshal(datas, &ri)
	if err != nil {
		return nil, err
	}

	return ri, err
}

//根据txid获取详细的的usdt交易信息
func (c *RpcClient) BtcGetrawtransaction(txid string) (*BtcTxInfo, error) {
	var (
		datas []byte
		tx    *BtcTx
		err   error
	)
	datas, err = c.Call("getrawtransaction", txid, 2)
	if err != nil {
		return nil, err
	}
	tx, err = decodeBtcGetrawtransaction(datas)
	if err != nil {
		return nil, err
	}
	if tx.Error != nil {
		return nil, errors.New(xutils.String(tx.Error))
	}
	return tx.Result, nil
}

//根据块hash获取区块信息
func (c *RpcClient) BtcGetblockheader(hash string) (*BtcBlockHeader, error) {
	var (
		datas       []byte
		blockHeader *BtcBlockHeader
		err         error
	)
	datas, err = c.Call("getblockheader", hash)
	if err != nil {
		return nil, err
	}
	blockHeader, err = decodeBtcBlockHeader(datas)
	if err != nil {
		return nil, err
	}
	return blockHeader, err
}
