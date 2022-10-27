package service

import (
	"github.com/group-coldwallet/ucaserver/model/bo"
	"github.com/group-coldwallet/ucaserver/model/vo"
)

type ChainService interface {
	//签名
	SignTx(tpl *bo.UcaTxTpl) (hex string, err error)

	//创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath string) (*vo.CreateAddrResult, error)
}
