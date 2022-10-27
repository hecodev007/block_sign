package services

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

func (u *ServiceAuditConfig) UpdateServiceAuditConfig(id int64, mp map[string]interface{}) (int, error) {

	db := model.DB().Table("service_audit_config").Where("id = ?", id).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (u *ServiceAuditConfig) GetServiceConfigLevel(id int64) (ServiceAuditConfig, error) {
	var sl ServiceAuditConfig
	db := model.DB().Where("id = ? ", id).First(&sl)
	return sl, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *ServiceAuditConfig) GetServiceConfigBySLid(sid, lid int) (*ServiceAuditConfig, error) {
	var sl = new(ServiceAuditConfig)
	db := model.DB().Where("service_id = ? and audit_level = ? ", sid, lid).First(sl)
	if sl != nil && sl.Id > 0 {
		return sl, nil
	}
	return sl, model.ModelError(db, global.MsgWarnModelNil)
}

func (u *ServiceAuditConfig) CreateServiceConfigLevel(cl *ServiceAuditConfig) error {
	db := model.DB()
	db.Table("service_audit_config").Omit("service_name", "audit_type").Create(cl)
	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		return err
	}
	return nil
}

func (u *ServiceAuditConfig) UpdateServiceConfigLevel(id int64, mp map[string]interface{}) error {
	db := model.DB()
	db.Table("service_audit_config").Where("id = ?", id).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return err
	}
	return nil
}
