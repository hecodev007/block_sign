package fil

import dao "starDataServer/models/po/fil"

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type TxInfo struct {
	BlockTxs []*dao.BlockTx
}

type ProcTask struct {
	Irreversible bool
	BestHeight   int64
	BlockChain   *dao.BlockChain
	TxInfos      []*dao.BlockTx
}

func (t *ProcTask) GetIrreversible() bool {
	return t.Irreversible
}

func (t *ProcTask) GetHeight() int64 {
	return t.BlockChain.Height
}

func (t *ProcTask) GetBestHeight() int64 {
	return t.BestHeight
}

func (t *ProcTask) GetConfirms() int64 {
	return t.BlockChain.Confirmations
}

func (t *ProcTask) GetBlockHash(index int) string {
	if len(t.BlockChain.Cids) == 0 {
		return ""
	}
	return t.BlockChain.Cids[0]["/"]
}

func (t *ProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.TxInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *ProcTask) GetBlock(index int) interface{} {
	return t.BlockChain
}
