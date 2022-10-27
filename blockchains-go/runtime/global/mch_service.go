package global

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
)

var MchServices map[int][]*entity.FcMchService

func InitMchService() {
	MchServices = map[int][]*entity.FcMchService{}
	mchs, err := dao.FcMchServiceFinds()
	if err != nil {
		panic(err.Error())
	}
	for _, srv := range mchs {
		MchServices[srv.MchId] = append(MchServices[srv.MchId], srv)
	}
}
