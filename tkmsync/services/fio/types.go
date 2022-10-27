package fio

import (
	"encoding/json"
	dao "rsksync/models/po/fio"
)

type ProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*dao.BlockTX
}
type procTaskJson struct {
	Irreversible bool           `json:"irreversible"`
	BestHeight   int64          `json:"bestHeight"`
	Block        *dao.BlockInfo `json:"block"`
	TxInfos      []*dao.BlockTX `json:"txInfos"`
}

func (t *ProcTask) MarshalJSON() ([]byte, error) {
	p := procTaskJson{
		Irreversible: t.irreversible,
		BestHeight:   t.bestHeight,
		Block:        t.block,
		TxInfos:      t.txInfos,
	}
	return json.Marshal(p)
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
