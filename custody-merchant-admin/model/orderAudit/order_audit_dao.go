package orderAudit

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
)

type NumsAudit struct {
	Nums       int `json:"nums" gorm:"column:nums"`
	AuditLevel int `json:"audit_level" gorm:"column:audit_level"`
}

func (e *Entity) FindOrderAuditByUId(uId int64) (Entity, error) {

	var auth Entity
	db := model.DB().Where("user_id =?", uId).First(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindOrderDetail(oId, uid int64) (*Entity, error) {
	var (
		auth  = new(Entity)
		build = new(xkutils.StringBuilder)
	)
	build.AddString(" select order_audit.* from order_audit ").
		AddString(" where order_audit.audit_result != 0 and order_audit.order_id = ? and order_audit.user_id =? ")

	db := model.DB().Raw(build.ToString(), oId, uid).Order("audit_level").Scan(auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindOrderDetailByNull(oId int64) ([]Entity, error) {
	var (
		auth []Entity
	)
	db := model.DB().Table("order_audit").
		Where("state = 0 and order_id = ? and audit_level != 4  and update_time is null ", oId).
		Order("audit_level").
		Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindOrderDetailBySuper(oId int64) ([]Entity, error) {
	var (
		auth []Entity
	)
	db := model.DB().Table("order_audit").Where(" state = 0 and order_id = ? and audit_level=4", oId).Order("update_time desc").Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindOrderDetailByASC(oId int64) ([]Entity, error) {
	var (
		auth  []Entity
		build = new(xkutils.StringBuilder)
	)
	build.AddString(" select order_audit.* from order_audit ").
		AddString(" where order_audit.state = 0 and order_audit.order_id = ? and order_audit.audit_level != 4 and order_audit.update_time is not null order by update_time ASC")
	db := model.DB().Raw(build.ToString(), oId).Scan(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) CountOrderByOIdLevel(oId int64, level int) (int, error) {
	count := int64(0)
	db := model.DB().Model(&Entity{}).Where("order_id = ?  and audit_result = 1 and audit_level = ? ", oId, level).Count(&count)
	return int(count), model.ModelError(db, global.MsgWarnModelNil)

}

func (e *Entity) FindAndDelAuditBySIdUId(sId, level int, uIds []int64) error {
	var (
		auth []Entity
		aIds []int64
	)
	findSql := "select a.* from order_audit a left join orders o on a.order_id = o.id where o.service_id = ? and a.audit_result = 0  and a.audit_level = ? and a.user_id in (?) "
	db := model.DB().Raw(findSql, sId, level, uIds).Scan(&auth)
	err := model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		return err
	}
	for i := 0; i < len(auth); i++ {
		aIds = append(aIds, auth[i].Id)
	}
	if err = model.DB().Exec("delete from order_audit where id in (?)", aIds).Error; err != nil {
		log.Errorf("FindAndDelAuditBySIdUId error: %v", err)
		return err
	}
	return nil
}

func (e *Entity) CreateAuditInfo(orders *Entity) error {
	db := model.DB().Begin()
	db.Omit("update_time").Create(&orders)
	if err := db.Error; err != nil {
		db.Rollback()
		log.Errorf("CreateAuditInfo error: %v", err)
		return err
	}
	db.Commit()
	return nil
}

func (e *Entity) BatchCreateAuditInfo(orders []Entity) error {

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

func (e *Entity) UpdateAuditInfo(oId, uId int64, mp map[string]interface{}) (int, error) {
	up := model.DB().Model(&Entity{}).Where("order_id=? and user_id=? and (audit_result = 0 or audit_result = 3)", oId, uId).Updates(mp)
	if err := up.Error; err != nil {
		log.Errorf("UpDateAuditInfo error: %v", err)
		return 0, err
	}
	return 1, nil
}

func (e *Entity) UpDateAuditInfoByLevel(oid int64, level int, mp map[string]interface{}) (int, error) {
	up := model.DB().Model(&Entity{}).Where("order_id=? and audit_level = ?", oid, level).Updates(mp)
	if err := up.Error; err != nil {
		log.Errorf("UpDateAuditInfoByLevel error: %v", err)
		return 0, err
	}
	return 1, nil
}

func (e *Entity) UpDateAuditInfoById(oid int64, level int, mp map[string]interface{}) error {

	up := model.DB().Model(&Entity{}).Where("order_id= ? and audit_level = ? and audit_result = 0 ", oid, level).Updates(mp)
	if err := up.Error; err != nil {
		log.Errorf("UpDateAuditInfoById error: %v", err)
		return err
	}
	return nil
}

func (e *Entity) FindAuditInfoByLevel(oid, uid int64) (Entity, error) {
	var a = Entity{}
	db := model.DB().Where("order_id=? and user_id = ?", oid, uid).First(&a)
	return a, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindNoResultAudit(oId int64) ([]NumsAudit, error) {
	var a []NumsAudit
	db := model.DB().Raw("select count(id) as nums, audit_level from order_audit where audit_result = 1 and order_id = ? group by audit_level order by audit_level", oId).Scan(&a)
	return a, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) DelOrderAudit(uid int64, res int) error {
	oa := new(Entity)
	db := model.DB().Where("user_id = ? and audit_result = ? ", uid, res).Delete(oa)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) GetOrderAuditByLevel(level, sid int) ([]Entity, error) {
	sql := "select order_audit.order_id from order_audit left join orders o on o.id = order_audit.order_id where order_audit.audit_level = ? and o.service_id = ? and order_audit.audit_result = 0 group by order_audit.order_id"
	var oa []Entity
	db := model.DB().Raw(sql, level, sid).Scan(&oa)
	if len(oa) > 0 {
		return oa, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetOrderNoAuditBySid(sid int) ([]Entity, error) {
	sql := "select order_audit.order_id from order_audit left join orders o on o.id = order_audit.order_id where o.service_id = ? and o.order_result = 0 group by order_audit.order_id"
	var oa []Entity
	db := model.DB().Raw(sql, sid).Scan(&oa)
	if len(oa) > 0 {
		return oa, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) DelOrderAuditByLevel(uids []int64, sid, level int) error {
	oa := new(Entity)
	db := model.DB().Where("user_id in (?) audit_result = 0 and (select count(1) from orders o where o.id = order_audit.order_id and o.service_id = ? limit 1) > 0 and audit_level = ? ", uids, sid, level).Delete(oa)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) CountOrderAuditPassByOIdUId(oId int64) (int64, error) {
	var count int64
	db := model.DB().Model(&Entity{}).Where("order_id=? and audit_result = 1", oId).Count(&count)
	return count, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) UpdateByState(id int64, state int, mp map[string]interface{}) error {
	up := model.DB().Model(&Entity{}).Where("user_id = ? and audit_result = 0 and state = ?", id, state).Updates(mp)
	if err := up.Error; err != nil {
		log.Errorf("UpDateAuditInfoById error: %v", err)
		return err
	}
	return nil
}

func (e *Entity) DelOrderAuditBySId(uids []int64, sid int) error {
	oa := new(Entity)
	db := model.DB().Where("user_id in (?) and audit_result = 0 and (select count(1) from orders o where o.id = order_audit.order_id and o.service_id = ? limit 1) > 0  ", uids, sid).Delete(oa)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) DelOrderAuditByUSId(uid int64, sid int) error {
	oa := new(Entity)
	db := model.DB().Where("user_id = ? and audit_result = 0 and (select count(1) from orders o where o.id = order_audit.order_id and o.service_id = ? limit 1) > 0  ", uid, sid).Delete(oa)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) FindOrderAuditByOId(oId int64) ([]Entity, error) {
	var list = []Entity{}
	db := model.DB().Model(&Entity{}).Where("order_id=?", oId).Order("update_time desc").Find(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}
