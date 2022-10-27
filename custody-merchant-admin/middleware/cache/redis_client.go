package cache

import (
	"context"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util"
	"fmt"
	goRds "github.com/go-redis/redis/v8"
	"time"
)

const (
	// go-redis在redis返回空的时候
	// 错误信息（error）为以下字符串
	redisNil = "redis: nil"
)

var (
	Client *rdb
	ctx    = context.Background()
)

type rdb struct {
	gr                goRds.Cmdable
	defaultExpiration time.Duration
}

func init() {
	GetRedisClientConn()
}

func InitRedis() {
	redisCfg := Conf.Redis
	prefix := "custody-merchant-admin"
	var cmdAble goRds.Cmdable
	if redisCfg.Model == "CLUSTER" {
		cmdAble = goRds.NewClusterClient(&goRds.ClusterOptions{
			Addrs:        redisCfg.ClusterAddress,
			Password:     redisCfg.ClusterPwd,
			DialTimeout:  30 * time.Second, // 设置连接超时
			ReadTimeout:  30 * time.Second, // 设置读取超时
			WriteTimeout: 30 * time.Second, // 设置写入超时
		})
		log.Infof("使用redis集群模式 %s", redisCfg.ClusterAddress)
	} else {
		cmdAble = goRds.NewClient(&goRds.Options{
			Addr:     redisCfg.AloneAddress,
			Password: redisCfg.AlonePwd,
		})
		log.Infof("使用redis单机模式 %s", redisCfg.AloneAddress)
	}
	log.Infof("redis key前缀:%s", prefix)
	Client = &rdb{
		gr:                cmdAble,
		defaultExpiration: time.Minute,
	}
	sta := Client.gr.Ping(ctx)
	if sta.Err() != nil {
		panic(fmt.Sprintf("尝试连接到redis服务失败: %s", sta.Err().Error()))
	}
	log.Infof("Redis服务初始化完成: %v ", sta)
	for i := 1; i < 5; i++ {
		GetRedisClientConn().Del(global.GetCacheKey(global.MenuTreeRole, i))
	}
}

func GetRedisClientConn() *rdb {
	if Client == nil || Client.gr == nil {
		InitRedis()
	}
	return Client
}

// Get
// redis命令：Get
// key对应的值不存在，返回空字符串
func (c *rdb) Get(key string, value interface{}) error {
	item := c.gr.Get(ctx, key)
	if item.Err() != nil {
		if item.Err().Error() == redisNil {
			//log.Warn("没有获取到任何数据")
			// 空值
			return nil
		}
		fmt.Printf(item.Err().Error())
		// 发生异常
		return item.Err()
	}
	bytes, err := item.Bytes()
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	err = util.Deserialize(bytes, value)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	return err
}

func (c *rdb) Set(key string, value interface{}, expiration time.Duration) error {
	b, err := util.Serialize(value)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	result := c.gr.Set(ctx, key, b, expiration)
	if result.Err() != nil {
		log.Errorf(err.Error())
		return result.Err()
	}
	return nil
}

// SetUseRawKey
//  SetUseRawKey使用原生的键
func (c *rdb) SetUseRawKey(key string, value interface{}, expiration time.Duration) error {
	result := c.gr.Set(ctx, key, value, expiration)
	if result.Err() != nil {
		log.Errorf(result.Err().Error())
		return result.Err()
	}
	return nil
}

func (c *rdb) ListPop(key string) ([]byte, error) {
	item := c.gr.LPop(ctx, key)
	if item.Err() != nil {
		if item.Err().Error() == redisNil {
			//log.Warn("没有获取到任何数据")
			return nil, nil
		} else {
			// 读取redis发生了错误
			return nil, item.Err()
		}
	}
	return []byte(item.Val()), nil
}

func (c *rdb) ListLen(key string) (int64, error) {
	value := c.gr.LLen(ctx, key)
	if value.Err() != nil {
		return 0, value.Err()
	}
	return value.Val(), nil
}

func (c rdb) ListRPush(key string, value interface{}) error {
	result := c.gr.RPush(ctx, key, value)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (c rdb) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	result := c.gr.SetNX(ctx, key, value, expiration)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val(), nil
}

func (c *rdb) Del(key string) error {
	result := c.gr.Del(ctx, key)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (c *rdb) ListRange(key string, start, stop int64) ([]string, error) {
	result := c.gr.LRange(ctx, key, start, stop)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result.Val(), nil
}
