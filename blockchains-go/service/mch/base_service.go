package mch

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/service"
)

type MchBaseService struct {
}

func (m *MchBaseService) GetMchName(appId int) (*entity.FcMch, error) {
	return dao.FcMchFindById(appId)
}

func (m *MchBaseService) GetAppIdByApiKey(apiKey string) (*entity.FcMch, error) {
	return dao.FcMchFindByApikey(apiKey)
}

func (m *MchBaseService) GetAppId(mchName string) (*entity.FcMch, error) {
	return dao.FcMchFindByPlatform(mchName)
}

func NewMchBaseService() service.MchService {
	return &MchBaseService{}
}
