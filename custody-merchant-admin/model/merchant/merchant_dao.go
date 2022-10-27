package merchant

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"errors"
	"gorm.io/gorm"
	"strings"
)

type ServiceLeveNums struct {
	Sid   int64 `json:"sid"  gorm:"column:sid"`
	Roles int64 `json:"roles" gorm:"column:roles"`
	Nums  int64 `json:"nums" gorm:"column:nums"`
}

func (e *Entity) InsertNewMerchant() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) UpdateMerchantItem() (err error) {
	err = e.Db.Table(e.TableName()).Updates(e).Error
	if err != nil {
		log.Errorf("UpdateMerchantInfo error: %v", err)
	}
	return
}

func (e *Entity) FindMerchantItemById(pId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? ", pId).Find(&e).Error
	if err != nil {
		log.Errorf("FindMerchantOneInfo error: %v", err)
	}
	if e.Phone == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindMerchantItemByPhone(phone string) (item Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("phone = ? ", phone).Find(&item).Error
	if err != nil {
		log.Errorf("FindMerchantOneInfo error: %v", err)
	}
	return
}

//FindMerchantItemByContactStr 联系方式查找用户
func (e *Entity) FindMerchantItemByContactStr(contactStr string) (item Entity, err error) {
	if strings.Contains(contactStr, "@") {
		err = e.Db.Table(e.TableName()).Where("email = ? ", contactStr).Find(&item).Error
	} else {
		err = e.Db.Table(e.TableName()).Where("phone = ? ", contactStr).Find(&item).Error
	}
	return
}

func (e *Entity) FindMerchantsByContactStr(contactStr string) (items []Entity, total int64, err error) {
	if strings.Contains(contactStr, "@") {
		err = e.Db.Table(e.TableName()).Where("email = ? ", contactStr).Find(&items).Error
	} else {
		err = e.Db.Table(e.TableName()).Where("phone = ? ", contactStr).Find(&items).Error
	}
	total = int64(len(items))
	return
}

func (e *Entity) FindMerchantListByReq(req domain.MerchantReqInfo) (list []MerchantApply, total int64, err error) {
	selectSql := e.Db.Table(e.TableName()).Select("user_info.id as id,user_info.is_push as is_push,user_info.is_test as is_test," +
		"user_info.name as name,user_info.phone as phone,user_info.email as email," +
		"ap.id_card_num as id_card_num,ap.passport_num as passport_num ,ap.created_at as created_at," +
		"ap.real_name_status as real_name_status,ap.real_name_at as real_name_at,ap.test_end as test_end," +
		"ap.contract_start_at as contract_start_at,ap.contract_end_at as contract_end_at," +
		"sf.verify_status as verify_status  ")
	selectSql = selectSql.Joins("left join service_finance sf on user_info.id = sf.account_id ")
	selectSql = selectSql.Joins("left join apply_pending ap on user_info.id = ap.account_id ")
	selectSql.Where("user_info.roles = 2 and user_info.apply_id is not null ")
	if req.AccountName != "" {
		selectSql = selectSql.Where("user_info.name = ?", req.AccountName)
	}
	if req.CardNum != "" {
		selectSql = selectSql.Where("user_info.id_card_num = ? or user_info.passport_num = ?", req.CardNum, req.CardNum)
	}
	if req.AccountId != "" {
		selectSql = selectSql.Where("user_info.id = ?", req.AccountId)
	}
	if req.RealNameStatus == "had_real" {
		selectSql = selectSql.Where("user_info.real_name_status = 1")
	} else if req.RealNameStatus == "no_real" {
		selectSql = selectSql.Where("user_info.real_name_status = 0")
	}
	//财务审核冻结，冻结正常异常筛选（normal/lock）
	if req.FvStatus != "" { //agree-通过，refuse-拒绝，wait-未处理
		if req.FvStatus == "agree" || req.FvStatus == "refuse" || req.FvStatus == "wait" {
			selectSql = selectSql.Where("sf.verify_status = ?", req.FvStatus)
		}
	}
	if req.LockStatus != "" { //agree-通过，refuse-拒绝，wait-未处理
		if req.LockStatus == "lock" {
			selectSql = selectSql.Where("sf.is_lock_finance = 1 ", req.FvStatus)
		} else if req.LockStatus == "unlock" {
			selectSql = selectSql.Where("sf.is_lock = 0", req.FvStatus)
		}
	}

	if req.ContactStr != "" {
		if strings.Contains(req.ContactStr, "@") {
			selectSql = selectSql.Where("user_info.email = ?", req.ContactStr)
		} else {
			selectSql = selectSql.Where("user_info.phone = ?", req.ContactStr)
		}
	}

	selectSql.Count(&total)
	err = selectSql.Limit(req.Limit).Offset(req.Offset).Order("id desc").Find(&list).Error
	if err != nil {
		log.Errorf("FindMerchantList error: %v", err)
	}
	return
}

//FindPushAbleApplies 获取所有可推送财务审核的申请
func (e *Entity) FindPushAbleApplies() (list []MerchantApply, err error) {
	selectSql := e.Db.Table(e.TableName()).Select("user_info.id as id,user_info.apply_id as apply_id,user_info.is_push,a.* ")
	selectSql = selectSql.Joins("left join apply_pending a on a.account_id = user_info.id ")
	selectSql = selectSql.Where("user_info.is_push = 0 and a.id_card_picture != '' and a.business_picture != '' and a.contract_picture != '' ")
	err = selectSql.Find(&list).Error
	if err != nil {
		log.Errorf("FindApplyList error: %v", err)
	}
	return
}

//UpdatePushAbleApplies 批量推送更新
func (e *Entity) UpdatePushAbleApplies(ids []int64) (err error) {
	selectSql := e.Db.Table(e.TableName())
	selectSql = selectSql.Where("id in (?) ", ids)
	err = selectSql.Update("is_push", 1).Error
	if err != nil {
		log.Errorf("FindApplyList error: %v", err)
	}
	return
}

// UpdateUser
// 更新用户
func (e *Entity) UpdateUser(u *Entity) error {
	tx := model.DB().Begin()
	if err := tx.Omit("create_time", "login_time").Save(u).Error; err != nil {
		log.Errorf("SaveUser error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// FindUserInfoList
// 查询用户信息列表
func (e *Entity) FindUserInfoList(uid int64, userSelect *domain.SelectUserInfo) ([]Entity, int64, error) {
	var (
		usi   = []Entity{}
		count int64
	)
	account := userSelect.Account
	if model.FilteredSQLInject(account) {
		return nil, 0, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(e).Where(" state != 2 ")
	if uid != 0 {
		db.Where(" (pid = ? or id =?) ", uid, uid)
	}
	if account != "" {
		if strings.Contains(account, "@") {
			db.Where(" email=? ", account)
		} else {
			db.Where(" phone=? ", account)
		}
	}
	db.Offset(userSelect.Offset).Order("user_info.id asc").Limit(userSelect.Limit).Find(&usi).Offset(-1).Limit(-1).Count(&count)
	return usi, count, model.ModelError(db, global.MsgWarnModelNil)
}

// FindSubUserInfoList
// 查询子用户信息列表
func (e *Entity) FindSubUserInfoList(uid int64, userSelect *domain.SelectUserInfo) ([]Entity, int64, error) {
	var (
		usi   = []Entity{}
		count int64
	)
	account := userSelect.Account
	if model.FilteredSQLInject(account) {
		return nil, 0, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(e).Where(" state != 2 ")
	if uid != 0 {
		db.Where(" (pid = ? or id =?) ", uid, uid)
	}
	db.Offset(userSelect.Offset).Order("user_info.id asc").Limit(userSelect.Limit).Find(&usi).Offset(-1).Limit(-1).Count(&count)
	return usi, count, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetUserMerchantPersonal(id int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Table("user_info").Where("id = ? ", id).First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}

func (e *Entity) GetMerchantPersonal(id int64, account string) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Table("user_info")
	if id != 0 {
		db = db.Debug().Where(" id =? ", id)
	}
	if account != "" {
		if strings.Contains(account, "@") {
			db = db.Where(" email=? ", account)
		} else {
			db = db.Where(" phone=? ", account)
		}
	}
	db.First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}

func (e *Entity) GetMerchantErr(id int64) ([]Entity, error) {
	u := []Entity{}
	db := model.DB().Table("user_info").Select("pwd_err,phone_code_err,email_code_err").Where(" id = ? or pid = ? ", id, id).Find(&u)
	return u, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetUserMerchantPersonalCode(id int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Table("user_info").Select("phone,phone_code,email").Where("id = ? ", id).First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}

// HaveUserByPIdAndUId
// 通过用户Id查询用户信息
func (e *Entity) HaveUserByPIdAndUId(id, pid int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Where("id = ? and pid = ? and state !=2", id, pid).First(u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetServiceUsers(rid, sid int) ([]Entity, error) {
	var u []Entity
	sql := "select user_info.* from user_info left join service_audit_role on service_audit_role.uid = user_info.id where user_info.roles=? and service_audit_role.sid=? and user_info.state = 0 "
	db := model.DB().Raw(sql, rid, sid).Scan(&u)
	if len(u) != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// UpdatePersonalUser
// 更新用户
func (e *Entity) UpdatePersonalUser(id int64, mp map[string]interface{}) error {
	if err := model.DB().Table("user_info").Where("id=?", id).Updates(mp).Error; err != nil {
		log.Errorf("UpdatePersonalUser error: %v", err)
		return err
	}
	return nil
}

// UpdatePersonalSubUser
// 更新用户
func (e *Entity) UpdatePersonalSubUser(id int64, mp map[string]interface{}) error {
	if err := model.DB().Table("user_info").Where("pid=?", id).Updates(mp).Error; err != nil {
		log.Errorf("UpdatePersonalUser error: %v", err)
		return err
	}
	return nil
}

// GetSubUserByEmail
// 通过用户邮箱查询用户信息
func (e *Entity) GetSubUserByEmail(email string) (*Entity, error) {
	u := Entity{}
	if model.FilteredSQLInject(email) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	l := model.DB().Where("email = ? and state != 2", email).First(&u)
	if u.Id != 0 {
		return &u, nil
	}
	return nil, model.ModelError(l, global.MsgWarnModelNil)
}

// GetSubUserByPhone
// 通过用户手机查询用户信息
func (e *Entity) GetSubUserByPhone(phone string) (*Entity, error) {
	u := Entity{}
	if model.FilteredSQLInject(phone) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	l := model.DB().Where("phone = ? and state != 2", phone).First(&u)
	return &u, model.ModelError(l, global.MsgWarnModelNil)
}

// UpdateSubUserById
// 通过用户Id更新用户
func (e *Entity) UpdateSubUserById(id int64, up map[string]interface{}) (int64, error) {

	db := model.DB().Model(&Entity{}).Where("id = ? ", id).Updates(up)
	return db.RowsAffected, model.ModelError(db, global.MsgWarnModelUpdate)
}

// ClearSubUserById
// 通过用户Id更新用户
func (e *Entity) ClearSubUserById(id int64, up map[string]interface{}) error {

	db := model.DB().Model(&Entity{}).Where("id = ? and state != 2", id).Updates(up)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

// ClearSubUserByPId
// 通过用户Id更新用户
func (e *Entity) ClearSubUserByPId(id int64, up map[string]interface{}) error {

	db := model.DB().Table("user_info").Where("(pid = ? or id =?) and state != 2", id, id).Updates(up)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

//LockMerchantItemById 冻结用户
func (e *Entity) LockMerchantItemById(userId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? and state != 2", userId).Update("state", 1).Error
	if err != nil {
		log.Errorf("FindMerchantOneInfo error: %v", err)
	}
	return
}

//UnlockMerchantItemById 解冻用户
func (e *Entity) UnlockMerchantItemById(userId int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("id = ? and state != 2", userId).Update("state", 0).Error
	if err != nil {
		log.Errorf("FindMerchantOneInfo error: %v", err)
	}
	return
}

func (e *Entity) CountLevelBySid(sId int) ([]ServiceLeveNums, error) {
	sar := []ServiceLeveNums{}
	db := model.DB().Table("user_info").
		Select(" service_audit_role.sid as sid, user_info.roles as roles, count(user_info.roles) as nums").
		Joins("left join service_audit_role on user_info.id = service_audit_role.uid").
		Where("service_audit_role.sid = ? ", sId).
		Group("service_audit_role.sid, user_info.roles").
		Order("user_info.roles asc").
		Find(&sar)
	return sar, model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) FindUserInfosBySid(sId, rId int) ([]Entity, error) {
	sar := []Entity{}
	db := model.DB().Table("user_info").
		Where("(select count(1) from service_audit_role where service_audit_role.sid = ? and service_audit_role.uid = user_info.id limit 1 ) > 0 and user_info.roles = ?", sId, rId).
		Find(&sar)
	return sar, model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) FindMerchantItemByRole(rid int64) (err error) {
	err = e.Db.Table(e.TableName()).Where("roles = ? ", rid).First(&e).Error
	if err != nil {
		log.Errorf("FindMerchantOneInfo error: %v", err)
	}
	if e.Phone == "" {
		err = gorm.ErrRecordNotFound
	}
	return
}

func (e *Entity) FindMerchantItemByPhoneAndEmail(phone, email string) (item Entity, err error) {
	err = e.Db.Table(e.TableName()).Where("phone = ? or email = ? and state != 2 ", phone, email).Find(&item).Error
	return
}
