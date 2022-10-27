package audit

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (o *OrderAudit) BatchCreateAuditInfo(orders []OrderAudit) error {

	for i := 0; i < len(orders); i++ {
		db := model.DB().Begin()
		order := orders[i]
		if err := db.Omit("result_name", "update_time").Create(&order).Error; err != nil {
			log.Errorf("BatchInsertUserService error: %v", err)
			db.Rollback()
			return err
		}
		db.Commit()
	}
	return nil
}

func (o *OrderAudit) CountOrderAuditPassByOIdUId(oId int64) (int64, error) {
	var count int64
	db := model.DB().Model(&OrderAudit{}).Where("order_id=? and audit_result = 1", oId).Count(&count)
	return count, model.ModelError(db, global.MsgWarnModelNil)
}
