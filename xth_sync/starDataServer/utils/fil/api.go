package fil

//区块高度
func (rpc *RpcClient) BlockHeight() (int64, error) {
	ret := new(SyncState)
	err := rpc.CallWithAuth("Filecoin.SyncState", rpc.Credentials, ret)
	if err != nil {
		return 0, err
	}
	if len(ret.ActiveSyncs) > 0 && len(ret.ActiveSyncs[0].Target.Blocks)>0 {
		return ret.ActiveSyncs[0].Target.Blocks[0].Height, nil
	}

	return 0, nil
}
func (rpc *RpcClient) BlockHead() (ret *ChainHead, err error) {
	ret = new(ChainHead)
	data := new(SyncState)
	err = rpc.CallWithAuth("Filecoin.SyncState", rpc.Credentials, data)
	if err != nil {
		return nil, err
	}
	if len(data.ActiveSyncs) < 1 {
		return ret, nil
	}
	ret.Cids = data.ActiveSyncs[0].Target.Cids
	ret.Blocks = data.ActiveSyncs[0].Target.Blocks
	if len(data.ActiveSyncs[0].Target.Blocks) > 0 {
		ret.Height = data.ActiveSyncs[0].Target.Blocks[0].Height
	}
	return ret, err
}

//当前头高度  HeadHeight<BlockHeight  HeadHeight能通过高度直接获取到区块信息
func (rpc *RpcClient) HeadHeight() (int64, error) {
	ret := new(BlockHeader)
	err := rpc.CallWithAuth("Filecoin.ChainHead", rpc.Credentials, ret)
	if err != nil {
		return 0, err
	}
	return ret.Height, nil
}

func (rpc *RpcClient) GetBlockChain(h int64, cids []map[string]string) (ret *ChainHead, err error) {

	ret = new(ChainHead)
	err = rpc.CallWithAuth("Filecoin.ChainGetTipSetByHeight", rpc.Credentials, ret, h, cids)

	return ret, err
}
func (rpc *RpcClient) GetBlockByHash(cid int64) (*BlockHeader, error) {
	return nil, nil
}
func (rpc *RpcClient) GetRawTransaction(h string) (*Transaction, error) {
	//ChainGetMessage
	return nil, nil
}
func (rpc *RpcClient) GetBlockTransactions(h string) (ret BlockMessages, err error) {

	param := make(map[string]string)
	param["/"] = h
	err = rpc.CallWithAuth("Filecoin.ChainGetBlockMessages", rpc.Credentials, &ret, param)
	if err == nil {
		for i, _ := range ret.BlsMessages {
			ret.BlsMessages[i].Cid = ret.Cids[i]["/"]
		}
		for i, _ := range ret.SecpkMessages {
			ret.SecpkMessages[i].Message.Cid = ret.Cids[len(ret.BlsMessages)+i]["/"]
		}
	}
	return
}

func (rpc *RpcClient) GetTransactionReceipt(txid string, tipset interface{}) (ret Receipt, err error) {
	//ret = new(Receipt)

	p := make(map[string]string)
	p["/"] = txid

	//var param2 []interface{}
	//for _, v := range blockcids {
	//	p2 := make(map[string]string)
	//	p2["/"] = v
	//	param2 = append(param2, p2)
	//}
	err = rpc.CallWithAuth("Filecoin.StateGetReceipt", rpc.Credentials, &ret, p, tipset)
	return
	//StateGetReceipt
}
func (rpc *RpcClient) GetParentReceipts(cid string) (ret []*Receipt, err error) {
	//ret = new(Receipt)

	p := make(map[string]string)
	p["/"] = cid

	//var param2 []interface{}
	//for _, v := range blockcids {
	//	p2 := make(map[string]string)
	//	p2["/"] = v
	//	param2 = append(param2, p2)
	//}
	err = rpc.CallWithAuth("Filecoin.ChainGetParentReceipts", rpc.Credentials, &ret, p)
	return
}
func (rpc *RpcClient) GetParentMessages(cid string) (ret []*Message, err error) {
	p := make(map[string]string)
	p["/"] = cid

	//var param2 []interface{}
	//for _, v := range blockcids {
	//	p2 := make(map[string]string)
	//	p2["/"] = v
	//	param2 = append(param2, p2)
	//}
	err = rpc.CallWithAuth("Filecoin.ChainGetParentMessages", rpc.Credentials, &ret, p)
	if err == nil {
		for i, _ := range ret {
			ret[i].Message.Cid = ret[i].Cid["/"]
		}
	}
	return
}
func (rpc *RpcClient) GetBlockByCid(cid string) (ret *BlockHeader, err error) {
	p := make(map[string]string)
	p["/"] = cid
	ret = new(BlockHeader)
	err = rpc.CallWithAuth("Filecoin.ChainGetBlock", rpc.Credentials, &ret, p)
	return ret, err
}
