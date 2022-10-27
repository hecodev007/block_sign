package cfx

import dao "cfxDataServer/models/po/cfx"

type ScanTask struct {
	Txids chan string
	Done  chan int
}
type TxInfo = dao.BlockTx
type ProcTask struct {
	Irreversible bool
	BestHeight   int64
	Block        *dao.BlockInfo
	TxInfos      []*dao.BlockTx
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
