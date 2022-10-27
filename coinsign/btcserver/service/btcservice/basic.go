package btcservice

import (
	"github.com/group-coldwallet/btcserver/model/bo"
	"github.com/group-coldwallet/btcserver/model/vo"
)

type BasicService interface {
	//BTC签名
	SignTx(tpl *bo.BtcTxTpl) (hex string, err error)

	//BTC创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error)
}
