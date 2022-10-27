package sync_sys

import (
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/util/xkutils"
	"time"
)

func InitSyncData() {
	go func() {
		// 清理日志文件
		ClearLog()
		// 初始化redis
		cache.InitRedis()

	}()
	go func() {
		dict.InitAllData()
		// 初始化字典
		xkutils.NewJobsSchedule(func() {
			dict.SyncHooPrice()
			go func() {
				dict.SyncHooGeekPrice()
			}()
		}, time.Hour*1)
	}()
}
