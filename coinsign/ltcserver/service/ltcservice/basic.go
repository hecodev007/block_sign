package ltcservice

import (
	"github.com/group-coldwallet/ltcserver/model/bo"
	"github.com/group-coldwallet/ltcserver/model/vo"
)

type BasicService interface {
	//LTC签名
	SignTx(tpl *bo.LtcTxTpl) (hex string, err error)

	//LTC创建地址
	CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error)
}
