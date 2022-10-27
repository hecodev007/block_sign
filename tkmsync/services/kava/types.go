package kava

import dao "rsksync/models/po/kava"

type TxInfo struct {
	tx     *dao.BlockTX
	txmsgs []*dao.BlockTXMsg
}

type ProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*TxInfo
}

func (t *ProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *ProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *ProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *ProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *ProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *ProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *ProcTask) GetBlock() interface{} {
	return t.block
}
