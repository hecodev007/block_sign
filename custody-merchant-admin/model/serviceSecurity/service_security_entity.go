package serviceSecurity

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

//Entity 业务线安全信息表
type Entity struct {
	Db          *orm.CacheDB `json:"-" gorm:"-"`
	Id          int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	BusinessId  int64        `json:"business_id" gorm:"column:business_id"` //业务线id
	ClientId    string       `json:"client_id" gorm:"column:client_id"`
	Secret      string       `json:"secret" gorm:"column:secret"`
	IpAddr      string       `json:"ip_addr" gorm:"column:ip_addr"`
	CallbackUrl string       `json:"callback_url" gorm:"column:callback_url"`
	//IsResetSecret int          `json:"is_reset_secret" gorm:"column:is_reset_secret"`
	IsSms           int       `json:"is_sms" gorm:"column:is_sms"`
	IsEmail         int       `json:"is_email" gorm:"column:is_email"`
	IsWithdrawal    int       `json:"is_withdrawal" gorm:"column:is_withdrawal"`
	IsWhitelist     int       `json:"is_whitelist" gorm:"is_whitelist"`
	IsGetAddr       int       `json:"is_get_addr" gorm:"column:is_get_addr"`
	IsIp            int       `json:"is_ip" gorm:"column:is_ip"`
	IsPlatformCheck int       `json:"is_platform_check" gorm:"is_platform_check"`
	IsAccountCheck  int       `json:"is_account_check" gorm:"is_account_check"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (e *Entity) TableName() string {
	return "service_security"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type BusinessSecurityDB struct {
	Id           int    `json:"id" gorm:"id"`
	ClientId     string `json:"client_id" gorm:"client_id"`
	Secret       string `json:"secret" gorm:"secret"`
	IpAddr       string `json:"ip_addr" gorm:"ip_addr"`
	CallbackUrl  string `json:"callback_url" gorm:"callback_url"`
	IsWithdrawal int    `json:"is_withdrawal" gorm:"is_withdrawal"`
	IsIp         int    `json:"is_ip" gorm:"is_ip"`
	Phone        string `json:"phone" gorm:"phone"`
	Email        string `json:"email" gorm:"email"`
}
