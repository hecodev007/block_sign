package vet

import dao "rsksync/models/po/vet"

type TxInfo struct {
	tx      *dao.BlockTX
	details []*dao.BlockTxDetail
}

type VetProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*TxInfo
}

func (t *VetProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *VetProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *VetProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *VetProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *VetProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *VetProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *VetProcTask) GetBlock() interface{} {
	return t.block
}
