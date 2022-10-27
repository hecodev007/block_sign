package service

import (
	"github.com/group-coldwallet/dogeserver/model/bo"
	"github.com/group-coldwallet/dogeserver/model/vo"
	"github.com/group-coldwallet/dogeserver/pkg/dogeutil"
)

type ChainService interface {
	//签名
	SignTx(tpl *dogeutil.DogeTxTpl) (hex string, err error)

	//创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath string) (*vo.CreateAddrResult, error)
}
