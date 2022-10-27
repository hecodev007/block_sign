package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/model/serviceChains"
)

func SaveMerchantChains(chains *domain.UpdateChains) error {
	return deals.SaveMerchantChains(chains)
}

func GetMerchantChainList(userSelect *domain.SearchChains) ([]serviceChains.SCUInfo, int64, error) {
	return deals.GetMerchantChainList(userSelect)
}

func GetMerchantChainsByAddr(chains *domain.UpdateChains) error {
	return deals.GetMerchantChainsByAddr(chains)
}

func GetMerchantChainsBySecureKey(secureKey string) (*serviceChains.Entity, error) {
	dao := serviceChains.NewEntity()
	err := dao.GetMerchantChainsBySecureKey(secureKey)
	if err != nil {
		return &serviceChains.Entity{}, err
	}
	return dao, nil
}

func GetMerchantChainsInfo(id int64) (*serviceChains.SCUInfo, error) {
	return deals.GetServiceChainsInfo(id)
}

func UpdateMerchantChainsInfo(id int64, mp map[string]interface{}) error {
	return deals.UpdateMerchantChainsInfo(id, mp)
}

func DeleteMerchantChainsInfo(id int64) error {
	return deals.DeleteMerchantChainsInfo(id)
}

func FindServiceChainsByMid(sel *domain.SelectUserInfo) ([]domain.MerchantServiceChains, int64, error) {
	return deals.FindServiceChainsByMainlist(sel)
}

func GetServiceChainsInfo(id int) (domain.ServiceAndCoin, error) {
	return deals.GetServiceChains(id)
}

func GetServiceChainsRolesInfo(id int) (domain.ServiceRolesInfo, error) {
	return deals.GetServiceChainsRolesInfo(id)
}
