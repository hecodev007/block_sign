package casbinRole

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/module/log"
	"fmt"
	"strings"
)

func GetAllSysRouter() (list []SysRouter, err error) {
	entity := &SysRouter{}
	err = model.DB().Table(entity.TableName()).Find(&list).Error
	return
}

func SaveNewMerchantSysRouter(db *orm.CacheDB, uid int64, routers []SysRouter) (err error) {

	err = DeleteRuleByV0(uid)
	if err != nil {
		return err
	}
	selectStr := "insert into casbin_rule (v0,v1,v2) values "
	values := make([]string, 0)
	for _, item := range routers {
		value := fmt.Sprintf("(%v,'%v','%v')", uid, item.Path, item.Method)
		values = append(values, value)
	}
	selectStr = selectStr + strings.Join(values, ",")
	err = db.Exec(selectStr).Error
	return
}

func DeleteRuleByV0(uid int64) error {
	if err := model.DB().Where("v0 =?", fmt.Sprintf("%d", uid)).Delete(Entity{}).Error; err != nil {
		log.Errorf("DeleteRuleByV0 error: %v", err)
		return err
	}
	return nil
}
