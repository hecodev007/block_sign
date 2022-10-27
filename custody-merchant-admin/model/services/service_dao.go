package services

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

// GetServiceById
// 根据业务Id获取业务线信息
func (s *ServiceEntity) GetServiceById(id int) (*ServiceEntity, error) {
	service := new(ServiceEntity)
	db := model.DB().Where("id = ? and state != 2", id).First(service)
	if service.Id != 0 {
		return service, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (s *ServiceEntity) CreateService(mp ServiceEntity) (int, error) {
	db := model.DB().Begin()
	db.Omit("update_time").Create(&mp)
	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		db.Rollback()
		return 0, err
	}
	db.Commit()
	return 1, nil
}

func (s *ServiceEntity) UpdateService(id int, mp map[string]interface{}) (int, error) {
	db := model.DB().Table("service").Where("id = ? and state = 0", id).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

// GetServiceAll
// 获取全部的业务线
func (s *ServiceEntity) GetServiceAll() ([]ServiceEntity, error) {
	var service []ServiceEntity
	db := model.DB().Where("state = 0").Find(service)
	if len(service) != 0 {
		return service, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
