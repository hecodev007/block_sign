package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/casbinRole"
	"custody-merchant-admin/model/orm"
	"encoding/json"
	"time"
)

//SetNewMerchantCasbin 新商户设置权限
func SetNewMerchantCasbin(db *orm.CacheDB, uid int64) (err error) {
	//获取所有权限目录
	key := global.GetCacheKey(global.AllSysRouter)
	allRouter := make([]casbinRole.SysRouter, 0)
	var saveStr string
	cache.GetRedisClientConn().Get(key, &saveStr)
	if saveStr != "" {
		json.Unmarshal([]byte(saveStr), &allRouter)
	}
	if len(allRouter) <= 0 {
		allRouter, err = casbinRole.GetAllSysRouter()
		if err != nil {
			return err
		}
	}
	if len(allRouter) > 0 {
		saveByte, _ := json.Marshal(&allRouter)
		cache.GetRedisClientConn().Set(key, string(saveByte), time.Hour)
	}
	//insert
	err = casbinRole.SaveNewMerchantSysRouter(db, uid, allRouter)
	return
}
