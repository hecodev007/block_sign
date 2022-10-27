package service

import (
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/shopspring/decimal"
)

//零散金额回收，目前针对utxo出账的时候 小额utxo回收，或者大金额回收
type RecycleService interface {
	//暂时定model 0 为小金额，1为大金额
	RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error)
}
