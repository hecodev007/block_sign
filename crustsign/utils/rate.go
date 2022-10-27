package utils

import (
	"sync"
	"time"
)

var limiter sync.Map

//地址不多可以这么干,多了要有回收机制,gocache
func Limit(addr string, sec int64) bool {
	value, ok := limiter.Load(addr)
	if !ok {
		limiter.Store(addr, time.Now().Unix())
		return true
	}
	if value.(int64) >= time.Now().Unix()-sec {
		return false
	} else {
		limiter.Store(addr, time.Now().Unix())
		return true
	}

}
func Free(addr string) {
	limiter.Delete(addr)
}
