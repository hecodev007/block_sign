package db

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	CDB *gorm.DB
	Acf *AdapterConfig
)

type AdapterConfig struct {
	Adapter  *gormadapter.Adapter
	Enforcer *casbin.Enforcer
}

type AdminCasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"column:ptype;size:128"`
	V0    string `gorm:"size:128"`
	V1    string `gorm:"size:128"`
	V2    string `gorm:"size:128"`
	V3    string `gorm:"size:128"`
	V4    string `gorm:"size:128"`
	V5    string `gorm:"size:128"`
}

func (r *AdminCasbinRule) TableName() string {
	return "admin_casbin_rule"
}

// CasbinDB
// @function: CasbinDB
// @description: 持久化到数据库  引入自定义规则
// @return: *casbin.Enforcer
func CasbinDB() *casbin.Enforcer {
	conf := Conf.DB["web"]
	if CDB == nil {
		sqlConnection := conf.UserName + ":" + conf.Pwd + "@tcp(" + conf.Host + ":" + conf.Port + ")/" + conf.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(sqlConnection), &gorm.Config{})
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		CDB = db
		adapter, err := gormadapter.NewAdapterByDBWithCustomTable(CDB, &AdminCasbinRule{}, "admin_casbin_rule")
		if err != nil {
			log.Errorf(err.Error())
		}
		cas := Conf.Casbin
		enforcer, err := casbin.NewEnforcer(cas.ModelPath, adapter)
		if err != nil {
			log.Errorf(err.Error())
		}
		Acf = &AdapterConfig{
			Adapter:  adapter,
			Enforcer: enforcer,
		}
		Acf.Enforcer.AddFunction("ParamsMatch", ParamsMatchFunc)
	}
	_ = Acf.Enforcer.LoadPolicy()
	return Acf.Enforcer
}
