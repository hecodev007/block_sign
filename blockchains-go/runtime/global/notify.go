package global

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"time"
)

func NotifyPrepare() error {
	if redis.Client == nil {
		return errors.New("redis.Client is NULL")
	}
	err := redis.Client.Set(redis.GetPrepareNotifyCacheKey(), "1", 10*24*time.Hour)
	if err != nil {
		log.Infof("重新加载配置，通知prepare失败:%v", err)
	} else {
		log.Infof("重新加载配置，通知prepare完成")
	}
	return err
}
