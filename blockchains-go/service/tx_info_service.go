package service

import "github.com/group-coldwallet/blockchains-go/model"

type TransactionInfoService interface {
	FindTxList(params *model.SearchTx) (results []*model.TxInfo, count int64, err error)
}
