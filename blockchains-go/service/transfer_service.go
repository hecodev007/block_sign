package service

import (
	"github.com/group-coldwallet/blockchains-go/entity"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type TransferService interface {
	//各个币种的交易接口
	//热钱包会即时返回txid，
	//但是热钱包特别注意，如果没写入order_hot表会报错，有的币种交易会自己写入order_hot表，有的不会则需要在方法里面实现
	TransferHot(ta *entity.FcTransfersApply) (txid string, err error)

	TransferCold(ta *entity.FcTransfersApply) error

	VaildAddr(address string) error
}
