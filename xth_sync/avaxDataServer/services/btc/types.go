package btc

import dao "avaxDataServer/models/po/btc"

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

type BtcProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*TxInfo
}

func (t *BtcProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *BtcProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *BtcProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *BtcProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *BtcProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *BtcProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *BtcProcTask) GetBlock() interface{} {
	return t.block
}
