package service

import "github.com/group-coldwallet/blockchains-go/entity"

type MchService interface {
	GetAppId(mchName string) (*entity.FcMch, error)
	GetMchName(appId int) (*entity.FcMch, error)
	GetAppIdByApiKey(apiKey string) (*entity.FcMch, error)
}
