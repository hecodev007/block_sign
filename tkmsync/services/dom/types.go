package dom

import dao "rsksync/models/po/dom"

type DomProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      [][]*dao.BlockTX
}

func (t *DomProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *DomProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *DomProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *DomProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *DomProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *DomProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *DomProcTask) GetBlock() interface{} {
	return t.block
}
