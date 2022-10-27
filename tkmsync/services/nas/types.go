package nas

import dao "rsksync/models/po/nas"

type NasProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      []*dao.BlockTX
}

func (t *NasProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *NasProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *NasProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *NasProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *NasProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *NasProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *NasProcTask) GetBlock() interface{} {
	return t.block
}
