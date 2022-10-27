package bchservice

import (
	"github.com/group-coldwallet/bchsign/model/bo"
	"github.com/group-coldwallet/bchsign/model/vo"
)

type BasicService interface {
	//BCH签名
	SignTx(tpl *bo.BchTxTpl) (hex string, err error)

	//BCH创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error)
}
