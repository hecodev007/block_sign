package limit

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (l *LimitWithdrawal) FindLimitWithdrawal(serviceId int) (LimitWithdrawal, error) {
	var limits LimitWithdrawal
	db := model.DB().Where(" service_id =?", serviceId).First(&limits)
	return limits, model.ModelError(db, global.MsgWarnModelNil)
}

func (l *LimitWithdrawal) CreateLimitWithdrawal(mp LimitWithdrawal) (int, error) {
	db := model.DB().Begin()
	db.Omit("update_time").Create(&mp)
	if err := db.Error; err != nil {
		db.Rollback()
		log.Errorf("CreateLimitWithdrawal error: %v", err)
		return 0, err
	}
	db.Commit()
	return 1, nil
}

func (l *LimitWithdrawal) UpDateLimitWithdrawal(id int64, mp map[string]interface{}) (int, error) {
	db := model.DB().Model(&LimitWithdrawal{}).Where("id=?", id).Updates(mp)
	if err := db.Error; err != nil {
		log.Errorf("UpDateLimitWithdrawal error: %v", err)
		return 0, err
	}
	return 1, nil
}
