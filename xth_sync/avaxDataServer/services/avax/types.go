package avax

import (
	"avaxDataServer/utils/avax"
)

const ASSETId = "nznftJBicce1PfWQeNEVBmDyweZZ6zcM3p78z9Hy9Hhdhfaxm"

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type AvaxProcTask struct {
	Irreversible bool
	Height       int64
	BestHeight   int64
	Confirms     int64
	TxInfos      []*avax.Transaction
}

func (t *AvaxProcTask) GetIrreversible() bool {
	return t.Irreversible
}

func (t *AvaxProcTask) GetHeight() int64 {
	return t.Height
}

func (t *AvaxProcTask) GetBestHeight() int64 {
	return t.BestHeight
}

func (t *AvaxProcTask) GetConfirms() int64 {
	return t.Confirms
}

func (t *AvaxProcTask) GetBlockHash() string {
	if len(t.TxInfos) > 0 {
		return t.TxInfos[0].ID
	}
	return ""
}

func (t *AvaxProcTask) GetTxs() []interface{} {
	ret := make([]interface{}, 0, len(t.TxInfos))
	for k, _ := range t.TxInfos {
		ret = append(ret, t.TxInfos[k])
	}
	return ret
}

func (t *AvaxProcTask) GetBlock() interface{} {
	return nil
}
