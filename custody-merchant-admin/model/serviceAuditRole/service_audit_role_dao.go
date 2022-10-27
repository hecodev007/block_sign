package serviceAuditRole

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
)

type UserAuditRoleName struct {
	Uid         int64  `gorm:"column:uid" json:"uid,omitempty"`
	Aid         int    `gorm:"column:aid" json:"aid,omitempty"`
	Sid         int    `gorm:"column:sid" json:"sid,omitempty"`
	ServiceName string `gorm:"column:service_name" json:"service_name,omitempty"`
	AuditName   string `gorm:"column:audit_name" json:"audit_name,omitempty"`
}

type UserServiceName struct {
	Id        int    `json:"id" gorm:"column:id"`
	Name      string `json:"name" gorm:"column:name"`
	AuditType int    `json:"audit_type" gorm:"column:audit_type"`
	TypeName  string `json:"type_name" gorm:"column:type_name"`
}

func (e *Entity) InsertNewPackage() (err error) {
	err = e.Db.Table(e.TableName()).Create(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

// SaveUserService
// 新增用户
func (e *Entity) SaveUserService(u *Entity) error {
	tx := model.DB().Begin()
	if err := tx.Save(u).Error; err != nil {
		log.Errorf("SaveUser error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (e *Entity) FindSARByUIdAndSId(uid int64, sid int) (*Entity, error) {
	sr := new(Entity)
	db := model.DB().Where("uid = ? and sid = ? and state != 2", uid, sid).First(sr)
	if sr != nil && sr.Id > 0 {
		return sr, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetUserServiceByUidSidNotLevel(uid int64, sid, level int) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Where("uid = ? and sid = ? and aid != ? and state = 0 ", uid, sid, level).First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindSARBySIdAndAId(sid, aid int) ([]Entity, error) {
	var sr []Entity
	db := model.DB().Where("sid = ? and aid = ? and state != 2", sid, aid).Find(&sr)
	if len(sr) > 0 {
		return sr, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) SaveSARInfo(sr *Entity) error {
	tx := model.DB().Begin()
	if err := tx.Save(sr).Error; err != nil {
		log.Errorf("SaveSARInfo error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (e *Entity) CreateSARInfoList(sr []Entity) error {
	for i := 0; i < len(sr); i++ {
		r := sr[i]
		tx := model.DB().Begin()
		if err := tx.Create(&r).Error; err != nil {
			log.Errorf("CreateSARInfoList error: %v", err)
			tx.Rollback()
			return err
		}
		tx.Commit()
	}
	return nil
}

func (e *Entity) FindUserAuditRoleName(uid int64) ([]UserAuditRoleName, error) {

	var (
		finds []UserAuditRoleName
		build xkutils.StringBuilder
	)
	build.AddString("select service_audit_role.sid as sid,service_audit_role.aid as aid,s.name as service_name, a.name as audit_name from service_audit_role ").
		AddString(" left join service s on service_audit_role.sid = s.id").
		AddString(" left join audit_role a on service_audit_role.aid = a.id")
	build.StringBuild(" where service_audit_role.uid = %d and service_audit_role.state != 2 ", uid)
	db := model.DB().Raw(build.ToString()).Scan(&finds)
	return finds, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindServiceHaveUserBySid(sid int) ([]Entity, error) {
	var sar []Entity
	db := model.DB().Table("service_audit_role").Where("aid != 0 and sid =? and (state = 0 or state = 1)", sid).Find(&sar)
	if len(sar) != 0 {
		return sar, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
func (e *Entity) GetUserServiceByUid(uid int64) ([]Entity, error) {
	var u []Entity
	db := model.DB().Where(" uid = ? ", uid).Find(&u)
	return u, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetUserServiceBySid(sid int) ([]Entity, error) {
	var u []Entity
	db := model.DB().Where("sid = ?", sid).Find(&u)
	if len(u) > 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) DelUserAuditRole(uid int64) error {
	db := model.DB().Table("service_audit_role").Where("uid =? and state = 0 ", uid).Delete(e)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) DelUserHaveService(uid int64, sids []int) error {
	sar := new(Entity)
	db := model.DB().Where("uid = ? and sid not in (?) ", uid, sids).Delete(sar)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) GetUserServiceByUidAndSid(uid int64, sid int) (*Entity, error) {
	u := Entity{}
	db := model.DB().Where("uid = ? and sid = ? ", uid, sid).First(&u)
	if &u != nil && u.Id > 0 {
		return &u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) UpdateUserServices(uid int64, sids []int) error {

	err := e.DelUserHaveService(uid, sids)
	if err != nil {
		return err
	}
	for i, _ := range sids {
		us := &Entity{}
		sid, err := e.GetUserServiceByUidAndSid(uid, sids[i])
		if err != nil {
			return err
		}
		if sid != nil {
			us.Id = sid.Id
		}
		us.Uid = uid
		us.Sid = sids[i]
		us.Aid = 0
		err = e.SaveUserService(us)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) UpdateUserAuditRole(uid int64, mp map[string]interface{}) error {
	db := model.DB().Table("service_audit_role").Where("uid =? and state = 0 ", uid).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) FindLevelUid(sId int, uId int64) (*Entity, error) {
	sar := new(Entity)
	db := model.DB().Table("service_audit_role").Where("sid =? and uid =? and state = 0 ", sId, uId).First(sar)
	if sar != nil && sar.Id > 0 {
		return sar, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelDelete)
}

func (e *Entity) FindUserServiceByUid(uid int64) ([]UserServiceName, error) {
	var (
		build xkutils.StringBuilder
		u     []UserServiceName
	)
	build.AddString("select distinct service.id as id, service.name as name,service.audit_type as audit_type from service_audit_role ").
		AddString(" left join service on service_audit_role.sid = service.id where service.state != 2 ")
	if uid != 0 {
		build.StringBuild(" and service_audit_role.uid = %d", uid)
	}
	db := model.DB().Raw(build.ToString()).Scan(&u)
	return u, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindLevelUidAllService(sId int, uId int64) (*Entity, error) {
	sar := new(Entity)
	db := model.DB().Table("service_audit_role").Where("sid =? and uid =? and state != 2 ", sId, uId).First(sar)
	if sar != nil && sar.Id > 0 {
		return sar, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelDelete)
}
