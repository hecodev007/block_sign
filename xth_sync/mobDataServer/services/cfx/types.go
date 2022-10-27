package cfx

import dao "cfxDataServer/models/po/cfx"
var monitorids =[]string{"5b42ba51a0ef6eff63dd8013dee661bfc8525017f55b7090a50d02e1640f5513"}

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
