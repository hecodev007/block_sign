package util

import (
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	mtx sync.RWMutex
)

func GetUUID() string {
	uid := uuid.NewV4()
	return strings.ReplaceAll(uid.String(), "-", "")
}

func GetRandomUpper(length int64) string {
	return GetRandomString("ABCDEFGHIJKLMNOPQRSTUVWXYZ", length)
}

func GetRandom(length int64) string {
	return GetRandomString("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", length)
}

//生成随机字符串
func GetRandomString(seed string, length int64) string {
	bytes := []byte(seed)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GetSeqNo() string {
	defer mtx.Unlock()
	mtx.Lock()
	ns := time.Now().UnixNano()
	time.Sleep(time.Microsecond)
	return fmt.Sprintf("%s%d", GetRandomUpper(10), ns)
}
