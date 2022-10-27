package atom

import dao "rsksync/models/po/atom"

type TxInfo struct {
	tx     *dao.BlockTX
	txmsgs []*dao.BlockTXMsg
}

type AtomProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*TxInfo
}

func (t *AtomProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *AtomProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *AtomProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *AtomProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *AtomProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *AtomProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *AtomProcTask) GetBlock() interface{} {
	return t.block
}
