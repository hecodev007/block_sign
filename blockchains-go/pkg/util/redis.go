package util

import (
	"github.com/go-redis/redis/v7"
)

//go-redis断开会自动从重连，无需关心重连问题
//https://jackckr.github.io/golang/go-redis%E8%BF%9E%E6%8E%A5%E6%B1%A0%E5%AE%9E%E7%8E%B0/

func NewRedisCli(addr, password string, dbIndex int) (*redis.Client, error) {
	if dbIndex < 0 {
		dbIndex = 0
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       dbIndex,  // use default DB
	})
	_, err := client.Ping().Result()
	return client, err
}
