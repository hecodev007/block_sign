package cds

import dao "rsksync/models/po/cds"

type CdsProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      [][]*dao.BlockTX
}

func (t *CdsProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *CdsProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *CdsProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *CdsProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *CdsProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *CdsProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *CdsProcTask) GetBlock() interface{} {
	return t.block
}
