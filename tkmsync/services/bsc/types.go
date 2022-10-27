package bsc

import dao "rsksync/models/po/bsc"

type BscProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      [][]*dao.BlockTX
}

func (t *BscProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *BscProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *BscProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *BscProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *BscProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *BscProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *BscProcTask) GetBlock() interface{} {
	return t.block
}
