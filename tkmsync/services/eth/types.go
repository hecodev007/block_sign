package eth

import dao "rsksync/models/po/eth"

type EthProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      [][]*dao.BlockTX
}

func (t *EthProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *EthProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *EthProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *EthProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *EthProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *EthProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *EthProcTask) GetBlock() interface{} {
	return t.block
}
