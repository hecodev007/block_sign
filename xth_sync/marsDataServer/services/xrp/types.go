package xrp

import dao "marsDataServer/models/po/xrp"

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type TxInfo struct {
	Tx *dao.BlockTx
	//vouts      []*dao.BlockTxVout
	//vins       []*dao.BlockTxVin
	Contractxs []*dao.TokenTx
}

type ProcTask struct {
	Irreversible bool
	BestHeight   int64
	Block        *dao.BlockInfo

	TxInfos []*TxInfo
}

func (t *ProcTask) GetIrreversible() bool {
	return t.Irreversible
}

func (t *ProcTask) GetHeight() int64 {
	return t.Block.Height
}

func (t *ProcTask) GetBestHeight() int64 {
	return t.BestHeight
}

func (t *ProcTask) GetConfirms() int64 {
	return t.Block.Confirmations
}

func (t *ProcTask) GetBlockHash() string {
	return t.Block.Hash
}

func (t *ProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.TxInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *ProcTask) GetBlock() interface{} {
	return t.Block
}
