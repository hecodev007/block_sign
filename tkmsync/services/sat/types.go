package sat

import dao "rsksync/models/po/sat"

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type TxInfo struct {
	tx    *dao.BlockTX
	vouts []*dao.BlockTXVout
	vins  []*dao.BlockTXVout
}

type ProcTask struct {
	TxInfos chan *TxInfo
	Done    chan int
}

type SatProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*TxInfo
}

func (t *SatProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *SatProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *SatProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *SatProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *SatProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *SatProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *SatProcTask) GetBlock() interface{} {
	return t.block
}
