package bch

//获取节点信息
func (c *RpcClient) GetChainInfo() (*BchChaininfo, error) {
	var (
		datas []byte
		info  *BchChaininfo
		err   error
	)
	datas, err = c.Call("getblockchaininfo")
	if err != nil {
		return nil, err
	}
	info, err = decodeBchChaininfo(datas)
	if err != nil {
		return nil, err
	}
	return info, err
}

func (c *RpcClient) GetBlockCount() (int64, error) {
	var (
		datas []byte
		info  *BchBlockCountInfo
		err   error
	)
	datas, err = c.Call("getblockcount")
	//log.Println("getblockcounta", string(datas))
	if err != nil {
		return 0, err
	}
	info, err = decodeBchBlockCountInfo(datas)
	if err != nil {
		return 0, err
	}
	return info.Result, err
}

//根据高度获取区块hash
func (c *RpcClient) GetBlockHash(index int64) (*BchGetBlockHash, error) {
	var (
		datas []byte
		tx    *BchGetBlockHash
		err   error
	)
	datas, err = c.Call("getblockhash", index)
	if err != nil {
		return nil, err
	}
	tx, err = decodeBchGetBlockHash(datas)
	if err != nil {
		return nil, err
	}
	return tx, err
}

//根据块hash获取区块信息
func (c *RpcClient) Getblock(hash string) (*BchBlock, error) {
	var (
		datas []byte
		block *BchBlock
		err   error
	)
	datas, err = c.Call("getblock", hash, 2)
	if err != nil {
		return nil, err
	}
	block, err = decodeBchBlock(datas)
	if err != nil {
		return nil, err
	}
	return block, err
}

//根据txid获取详细的的usdt交易信息
func (c *RpcClient) Getrawtransaction(txid string) (*BchTx, error) {
	var (
		datas []byte
		tx    *BchTx
		err   error
	)
	datas, err = c.Call("getrawtransaction", txid, 2)
	if err != nil {
		return nil, err
	}
	tx, err = decodeBchGetrawtransaction(datas)
	if err != nil {
		return nil, err
	}
	return tx, err
}

//根据块hash获取区块信息
func (c *RpcClient) Getblockheader(hash string) (*BchBlockHeader, error) {
	var (
		datas       []byte
		blockHeader *BchBlockHeader
		err         error
	)
	datas, err = c.Call("getblockheader", hash)
	if err != nil {
		return nil, err
	}
	blockHeader, err = decodeBchBlockHeader(datas)
	if err != nil {
		return nil, err
	}
	return blockHeader, err
}
