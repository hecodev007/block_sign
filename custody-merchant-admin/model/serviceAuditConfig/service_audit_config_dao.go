package serviceAuditConfig

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

func (e *Entity) CreateServiceConfigLevel() error {
	db := model.DB()
	db.Table("service_audit_config").Omit("service_name", "audit_type").Create(e)
	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		return err
	}
	return nil
}
