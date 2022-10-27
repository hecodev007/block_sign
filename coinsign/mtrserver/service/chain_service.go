package service

import (
	"github.com/group-coldwallet/mtrserver/model/bo"
	"github.com/group-coldwallet/mtrserver/model/vo"
)

type ChainService interface {
	//签名
	SignTx(tpl *bo.TxTpl) (hex string, err error)

	//创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath string) (*vo.CreateAddrResult, error)
}
