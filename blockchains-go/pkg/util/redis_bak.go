package util

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"os"
	"time"
)

var redisConnPool *redis.Pool

type RedisClient struct {
	redis.Conn
}

// redis connect pool
func CreateRedisPool(url, user, passwd string) {
	redisConnPool = &redis.Pool{
		MaxIdle:     16, // Free play
		IdleTimeout: 8,  // Free play
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", url)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil, err
			}
			if user != "" && passwd != "" {
				if _, err := c.Do("AUTH", passwd); err != nil {
					c.Close()
					fmt.Fprintln(os.Stderr, err)
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func CloseRedisPool() {
	if redisConnPool != nil {
		redisConnPool.Close()
	}
}

// get redis pool
func GetRedisPool() *redis.Pool {
	return redisConnPool
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	return err
}

// alloc from redis pool
func AllocRedisClient() (*RedisClient, error) {
	if redisConnPool == nil {
		return nil, errors.New("Redis connect pool not create, please use CreateRedisPool function")
	}

	return &RedisClient{redisConnPool.Get()}, nil
}

// free to redis pool
func FeeeRedisClient(c *RedisClient) {
	if c != nil {
		c.Conn.Close()
	}
}

func CreateRedisClient(url string) (*RedisClient, error) {
	conn, err := redis.Dial("tcp", url)
	if err != nil {
		return nil, err
	}

	return &RedisClient{conn}, nil
}

func (c *RedisClient) Get(key string) (string, error) {
	value, err := redis.String(c.Do("GET", key))
	return value, err
}

func (c *RedisClient) SetNx(args ...interface{}) (reply interface{}, err error) {
	reply, err = c.Do("SETNX", args...)
	return
}

func (c *RedisClient) Set(args ...interface{}) error {
	_, err := c.Do("SET", args...)
	return err
}

func (c *RedisClient) Expire(k string, second int64) error {
	_, err := c.Do("EXPIRE", k, second)
	return err
}

func (c *RedisClient) Exists(key string) (bool, error) {
	isKeyExit, err := redis.Bool(c.Do("EXISTS", key))
	return isKeyExit, err
}

func (c *RedisClient) Del(key string) error {
	_, err := c.Do("DEL", key)
	return err
}

func (c *RedisClient) SetJson(args ...interface{}) (interface{}, error) {
	n, err := c.Do("SETNX", args...)
	return n, err
}

func (c *RedisClient) GetJson(key string) ([]byte, error) {
	valueGet, err := redis.Bytes(c.Do("GET", key))
	return valueGet, err
}

func (c *RedisClient) LeftPush(args ...interface{}) error {
	_, err := c.Do("lpush", args...)
	return err
}

func (c *RedisClient) LRange(args ...interface{}) ([]interface{}, error) {
	values, err := redis.Values(c.Do("lrange", args...))
	return values, err
}

func (c *RedisClient) Rpop(args ...interface{}) ([]byte, error) {
	value, err := redis.Bytes(c.Do("rpop", args...))
	return value, err
}

func (c *RedisClient) LLEN(args ...interface{}) (int64, error) {
	value, err := redis.Int64(c.Do("llen", args...))
	return value, err
}

func (c *RedisClient) Ping() bool {
	if value, _ := redis.String(c.Conn.Do("ping")); value != "PONG" {
		return false
	}
	return true
}

func (c *RedisClient) Close() {
	c.Conn.Close()
}
