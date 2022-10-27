package wtc

import dao "hecosync/models/po/yotta"

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

func (t *ProcTask) ParentHash() string {
	return t.Block.Previousblockhash
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

func (t *ProcTask) SetBestHeight(bestheight int64, Irreversible bool) {
	t.BestHeight = bestheight
	t.Block.Confirmations = bestheight - t.Block.Height + 1
	t.Irreversible = Irreversible
}
