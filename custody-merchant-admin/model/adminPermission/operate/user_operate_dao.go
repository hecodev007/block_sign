package operate

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

// NewOperate
// 新增用户操作
func (e *Entity) NewOperate() error {
	tx := model.DB().Begin()
	if err := tx.Omit("deleted_at", "updated_at", "login_time").Create(e).Error; err != nil {
		log.Errorf("NewOperate error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (e *Entity) FindOperateByUserId(uId int64) ([]Entity, error) {
	lst := []Entity{}
	db := model.DB().Table("user_operate").Where("user_id=?", uId).Find(&lst)
	return lst, model.ModelError(db, global.MsgWarnModelNil)
}
