package kardia

import dao "rsksync/models/po/kardia"

type KardiaProcTask struct {
	irreversible bool
	bestHeight   int64
	block        *dao.BlockInfo
	txInfos      [][]*dao.BlockTX
}

func (t *KardiaProcTask) GetIrreversible() bool {
	return t.irreversible
}

func (t *KardiaProcTask) GetHeight() int64 {
	return t.block.Height
}

func (t *KardiaProcTask) GetBestHeight() int64 {
	return t.bestHeight
}

func (t *KardiaProcTask) GetConfirms() int64 {
	return t.block.Confirmations
}

func (t *KardiaProcTask) GetBlockHash() string {
	return t.block.Hash
}

func (t *KardiaProcTask) GetTxs() []interface{} {
	var res []interface{}
	for _, txInfo := range t.txInfos {
		res = append(res, txInfo)
	}
	return res
}

func (t *KardiaProcTask) GetBlock() interface{} {
	return t.block
}
