package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	goRds "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"strings"
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
	gr goRds.Cmdable
}

func InitRedis(addr, pwd string, cluster bool) {
	if addr == "" && pwd == "" {
		log.Infof("地址和密码都为空，不启用redis服务")
		return
	}

	log.Infof("redis %s", addr)
	var cmdAble goRds.Cmdable

	if cluster {
		srvName := addr
		portIdx := strings.Index(srvName, ":")
		if portIdx > -1 {
			srvName = srvName[:portIdx]
		}
		cmdAble = goRds.NewClusterClient(&goRds.ClusterOptions{
			Addrs:     []string{addr},
			Password:  pwd,
			TLSConfig: &tls.Config{ServerName: srvName},
		})
		log.Info("使用redis集群模式")
	} else {
		cmdAble = goRds.NewClient(&goRds.Options{
			Addr:     addr,
			Password: pwd,
		})
		log.Info("使用redis单机模式")
	}

	Client = &rdb{
		gr: cmdAble,
	}
	if sta := Client.gr.Ping(ctx); sta.Err() != nil {
		panic(fmt.Sprintf("尝试连接到redis服务失败: %s", sta.Err().Error()))
	}
	log.Info("Redis服务初始化完成")
}

func (c *rdb) Get(key string) (string, error) {
	item := c.gr.Get(ctx, key)
	if item.Err() != nil {
		if item.Err().Error() == redisNil {
			// 空值
			return "", nil
		}
		// 发生异常
		return "", item.Err()
	}
	return item.Val(), nil
}

func (c *rdb) TxPipeline() goRds.Pipeliner {
	return c.gr.TxPipeline()
}

func (c *rdb) Set(key string, value interface{}, expiration time.Duration) error {
	result := c.gr.Set(ctx, key, value, expiration)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

//  SetUseRawKey使用原生的键
func (c *rdb) SetUseRawKey(key string, value interface{}, expiration time.Duration) error {
	result := c.gr.Set(ctx, key, value, expiration)
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (c *rdb) ListPop(key string) ([]byte, error) {
	item := c.gr.LPop(ctx, key)
	if item.Err() != nil {
		if item.Err().Error() == redisNil {
			// 没有获取到任何数据
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

func (c rdb) ListLRem(key string, count int64, value interface{}) error {
	result := c.gr.LRem(ctx, key, count, value)
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

func (c *rdb) ListLTrim(key string, start, stop int64) (string, error) {
	result := c.gr.LTrim(ctx, key, start, stop)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

func (c *rdb) ZRange(key string, start, stop int64) ([]string, error) {
	result := c.gr.ZRange(ctx, key, start, stop)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result.Val(), nil
}

func (c *rdb) ZCard(key string) (int64, error) {
	result := c.gr.ZCard(ctx, key)
	if result.Err() != nil {
		return 0, result.Err()
	}
	return result.Val(), nil
}

func (c *rdb) ZScore(key string, member string) (float64, error) {
	result := c.gr.ZScore(ctx, key, member)
	if result.Err() != nil {
		return 0, result.Err()
	}
	return result.Val(), nil
}

func (c *rdb) ZRem(key string, member string) (int64, error) {
	result := c.gr.ZRem(ctx, key, member)
	if result.Err() != nil {
		return 0, result.Err()
	}
	return result.Val(), nil
}
