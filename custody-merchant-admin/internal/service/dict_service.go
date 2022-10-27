package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/model/adminPermission/role"
	"custody-merchant-admin/model/unitUsdt"
	"custody-merchant-admin/module/dict"
)

func FindDictByTypeService(t string) ([]domain.DictListInfo, error) {
	return deals.FindDictByType(t)
}

func FindPhoneCodeAllService() ([]domain.PhoneInfo, error) {

	return deals.FindPhoneCodeAll()
}

func GetSysRoleAll(rid int) ([]role.Entity, error) {
	return deals.GetSysRoleAllByRid(rid)
}

func GetMerchantRoleAll() ([]role.Entity, error) {
	var rl = make([]role.Entity, 0)
	for i, s := range dict.SysRoleNameList {
		if i <= 1 {
			continue
		}
		rl = append(rl, role.Entity{
			Id:   i + 1,
			Name: s,
		})
	}

	return rl, nil
}

func FindAllUnit() ([]unitUsdt.UnitUsdt, error) {
	dao := new(unitUsdt.UnitUsdt)
	return dao.GetUnitUsdtList()
}
