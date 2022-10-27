package util

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	mtx sync.RWMutex
)

//生成随机字符串
func GetRandomString(length int64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GetTimeNowStr() string {
	defer mtx.Unlock()
	mtx.Lock()
	ns := time.Now().UnixNano()
	time.Sleep(time.Microsecond)
	return fmt.Sprintf("%d", ns)
}
