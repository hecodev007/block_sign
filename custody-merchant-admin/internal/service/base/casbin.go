package base

import (
	"custody-merchant-admin/db"
	"custody-merchant-admin/model/adminPermission/api"
)

type CasbinService struct {
}

// CasbinCreate
// 新增权限
func (d *CasbinService) CasbinCreate(userId, path, method string) error {
	cm := db.AdminCasbinRule{
		Ptype: "p",
		V0:    userId,
		V1:    path,
		V2:    method,
	}
	return cm.Create()
}

// CasbinCreateBatch
// 批量新增权限
func (d *CasbinService) CasbinCreateBatch(uId string, info []api.Entity) error {
	cm := db.AdminCasbinRule{
		V0: uId,
	}
	return cm.CreateBatch(info)
}

// CasbinList
// 获取权限
func (d *CasbinService) CasbinList(userId string) [][]string {
	cm := db.AdminCasbinRule{V0: userId}
	return cm.List()
}

// CasbinRemove
// 移除权限
func (d *CasbinService) CasbinRemove(userId, path, method string) error {
	cm := db.AdminCasbinRule{
		Ptype: "p",
		V0:    userId,
		V1:    path,
		V2:    method,
	}
	return cm.Remove()
}

// CasbinRemoveBatch
// 批量移除权限
func (d *CasbinService) CasbinRemoveBatch(uId string, info []api.Entity) error {
	cm := db.AdminCasbinRule{
		Ptype: "p",
		V0:    uId,
	}
	return cm.RemoveBatch(info)
}
