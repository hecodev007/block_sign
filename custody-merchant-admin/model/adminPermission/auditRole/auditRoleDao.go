package auditRole

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

// GetAuditLevelAll
// 获取审核角色等级
func GetAuditLevelAll() ([]AuditRole, error) {
	var service []AuditRole
	db := model.DB().Order("id asc").Find(&service)
	if len(service) != 0 {
		return service, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAuditLevelById
// 获取审核角色等级
func GetAuditLevelById(id int) (*AuditRole, error) {
	var service = new(AuditRole)
	db := model.DB().Where("id =? and state = 0", id).First(&service)
	if service != nil && service.Id > 0 {
		return service, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
