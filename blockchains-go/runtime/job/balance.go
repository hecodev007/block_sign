package job

import (
	"encoding/json"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"strconv"
)

const (
	mchAllBalanceCacheKey  = "bchsc_mchAllBalance_"
	mchAllBalanceExpireSec = 300
)

// Job Specific Functions
type BalanceJob struct {
	Second int64
}

func (e BalanceJob) Run() {
	log.Infof("%s金额设置缓存", util.GetChinaTimeNowFormat())
	//读取商户信心
	for _, v := range global.MchBaseInfo {
		//暂时先限制商户进来
		if v.AppId != 1 && v.AppId != 102 {
			continue
		}
		redisHelper, err := util.AllocRedisClient()
		if err != nil {
			log.Errorf("redis 异常:%s", err.Error())
			continue
		}
		if e.Second == 0 {
			e.Second = mchAllBalanceExpireSec
		}
		err = getMchAllBalance(v.AppId, redisHelper, e.Second)
		if err != nil {
			log.Errorf("商户[%d],获取金额异常：%s", v.AppId, err.Error())
			continue
		}
	}

}

func getMchAllBalance(appid int, redisHelper *util.RedisClient, second int64) error {
	defer redisHelper.Close()
	result, err := api.BalanceService.GetMchAllBalance(appid)
	if err != nil {
		return err
	}
	log.Infof("GetMchAllBalance AppId=%d 设置数据到缓存", appid)
	key := getCacheKey(strconv.Itoa(appid))
	setToCache(redisHelper, key, &result, second)
	return nil
}

func setToCache(redisHelper *util.RedisClient, key string, data *[]*model.CoinBalance, second int64) {
	ms, err := json.Marshal(data)
	if err != nil {
		log.Errorf("setToCache json.Marshal error:%v", err)
		return
	}

	err = redisHelper.Set(key, ms)
	if err != nil {
		log.Errorf("setToCache redis Set error %v", err)
		return
	}
	err = redisHelper.Expire(key, second)
	if err != nil {
		log.Errorf("setToCache redis Set error %v", err)
		return
	}
}

func getCacheKey(appID string) string {
	return mchAllBalanceCacheKey + appID
}
