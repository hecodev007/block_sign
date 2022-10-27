package zecservice

import (
	"github.com/group-coldwallet/zecserver/model/bo"
	"github.com/group-coldwallet/zecserver/model/vo"
)

type BasicService interface {
	//ZEC签名
	SignTx(tpl *bo.ZecTxTpl) (hex string, err error)

	//ZEC创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error)
}
