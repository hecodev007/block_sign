package api

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db           *orm.CacheDB `json:"-" gorm:"-"`
	Sex          int          `json:"sex" gorm:"column:sex"`
	State        int          `json:"state" gorm:"column:state"`
	Id           int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Pid          int64        `json:"pid" gorm:"column:pid"`
	Uid          int64        `json:"uid" gorm:"column:uid"`
	Name         string       `json:"name" gorm:"column:name"`
	Email        string       `json:"email" gorm:"column:email"`
	Phone        string       `json:"phone" gorm:"column:phone"`
	PhoneCode    string       `json:"phone_code" gorm:"column:phone_code"`
	Password     string       `json:"password" gorm:"column:password"`
	Salt         string       `json:"salt" gorm:"column:salt"`
	Role         int          `json:"roles" gorm:"column:roles"`
	Remark       string       `json:"remark" gorm:"column:remark"`
	Reason       string       `json:"reason" gorm:"column:reason"`
	Passport     string       `json:"passport" gorm:"column:passport"`
	Identity     string       `json:"identity" gorm:"column:identity"`
	PwdErr       int          `json:"pwd_err" gorm:"column:pwd_err"`
	PhoneCodeErr int          `json:"phone_code_err" gorm:"column:phone_code_err"`
	EmailCodeErr int          `json:"email_code_err" gorm:"column:email_code_err"`
	LoginTime    time.Time    `json:"login_time" gorm:"column:login_time"`
	CreatedAt    time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_user"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
