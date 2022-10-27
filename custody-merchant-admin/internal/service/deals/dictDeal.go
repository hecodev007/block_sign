package deals

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/adminPermission/role"
	"custody-merchant-admin/model/base"
)

type DictDeal struct {
}

func FindDictByType(t string) ([]domain.DictListInfo, error) {
	var (
		dao = base.DictList{}
		dl  []domain.DictListInfo
	)
	// 根据字典类型名称获取字典
	byType, err := dao.FindDictByType(t)
	if err != nil {
		return nil, err
	}
	// 遍历字典列表
	for i := 0; i < len(byType); i++ {
		// 返回字典列表给前端
		dl = append(dl, domain.DictListInfo{
			DictName:  byType[i].DictName,
			DictValue: byType[i].DictValue,
		})
	}
	return dl, nil
}

func GetSysRoleAllByRid(rid int) ([]role.Entity, error) {
	dao := role.NewEntity()
	return dao.GetAdminRoleAllByRid(rid)
}
