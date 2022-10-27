package limit

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (l *LimitTransfer) FindLimitTransfer(serviceId int) (LimitTransfer, error) {
	var limits LimitTransfer
	db := model.DB().Where("service_id =?", serviceId).First(&limits)
	return limits, model.ModelError(db, global.MsgWarnModelNil)
}

func (l *LimitTransfer) CreateLimitTransfer(mp LimitTransfer) (int, error) {
	db := model.DB().Begin()
	db.Omit("update_time").Create(&mp)
	if err := db.Error; err != nil {
		db.Rollback()
		log.Errorf("CreateLimitTransfer error: %v", err)
		return 0, err
	}
	db.Commit()
	return 1, nil
}

func (l *LimitTransfer) UpDateLimitTransfer(id int64, mp map[string]interface{}) (int, error) {
	db := model.DB().Model(&LimitTransfer{}).Where("id=?", id).Updates(mp)
	if err := db.Error; err != nil {
		log.Errorf("UpDateLimitTransfer error: %v", err)
		return 0, err
	}
	return 1, nil
}
