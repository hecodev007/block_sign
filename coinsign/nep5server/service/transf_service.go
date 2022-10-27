package service

import (
	"github.com/group-coldwallet/nep5server/model/bo"
	"github.com/group-coldwallet/nep5server/model/vo"
)

type TransfService interface {
	//token签名
	TokenSign(from, to, scriptAddr string, amount int64) (raw, txid string, err error)

	//创建地址
	CreateAddr(param *bo.CreateAddrBO) (*vo.CreateAddrVO, error)
}
