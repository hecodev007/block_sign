package btc

import dao "btcsync/models/po/btc"

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type TxInfo struct {
	Tx         *dao.BlockTx
	Vouts      []*dao.BlockTxVout
	Vins       []*dao.BlockTxVin
	Contractxs []*dao.BtcUsdtTx
}

type ProcTask struct {
	Irreversible bool
	BestHeight   int64
	Block        *dao.BlockInfo
	TxInfos      []*TxInfo
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
func (t *ProcTask) SetBestHeight(bestheight int64, Irreversible bool) {
	t.BestHeight = bestheight
	t.Block.Confirmations = bestheight - t.Block.Height + 1
	t.Irreversible = Irreversible
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
