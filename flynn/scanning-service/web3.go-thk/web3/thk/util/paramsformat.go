package util

func (param *GetAccountJson) FormatParams(address string, chainid string) error {
	param.ChainId = chainid
	param.Address = address
	return nil
}

func (param *GetBlockTxsJson) FormatParams(chainId string, height string, page string, size string) error {
	param.ChainId = chainId
	param.Height = height
	param.Page = page
	param.Size = size
	return nil
}

func (param *Transaction) FormatParams(transcation *Transaction) error {
	return nil
}

func (param *GetTxByHash) FormatParams(chainId string, hash string) error {
	param.ChainId = chainId
	param.Hash = hash
	return nil
}

func (param *GetBlockHeader) FormatParams(chainId string, height string) error {
	param.ChainId = chainId
	param.Height = height
	return nil
}

func (param *GetCommitteeJson) FormatParams(chainId string, epoch int) error {
	param.ChainId = chainId
	param.Epoch = epoch
	return nil
}
func (param *CompileContractJson) FormatParams(chainId string, contract string) error {
	param.ChainId = chainId
	param.Contract = contract
	return nil
}

func (param *PingJson) FormatParams(chainid string) error {
	param.ChainId = chainid
	return nil
}

func (param *GetChainInfoJson) FormatParams(chainIds []int) error {
	param.ChainId = chainIds
	return nil
}

func (param *GetStatsJson) FormatParams(chainId string) error {
	param.ChainId = chainId
	return nil
}
func (param *GetMultiStatsJson) FormatParams(chainId string) error {
	param.ChainId = chainId
	return nil
}

//chainId ,address ,startHeight ,endHeight string
func (param *GetTransactionsJson) FormatParams(chainId, address, startHeight, endHeight string) error {
	param.ChainId = chainId
	param.Address = address
	param.StartHeight = startHeight
	param.EndHeight = endHeight
	return nil
}
