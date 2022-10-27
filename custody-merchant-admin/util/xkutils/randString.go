package xkutils

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

var (
	r   *rand.Rand
	num int64
)

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

func NewUUId(namespace string) string {
	u1 := uuid.NewV5(uuid.NewV4(), namespace)
	uuid := strings.Replace(u1.String(), "-", "", 4)
	builder := StringBuilder{}
	build := builder.StringBuild("%s", uuid)
	return build.ToString()
}

// Generate
//生成20位订单号
//前面15位代表时间精确到毫秒，最后4位代表序号
func Generate(domains string, t time.Time) string {
	m := t.UnixMilli()
	ms := sup(m, 15)
	i := atomic.AddInt64(&num, 1)
	r := i % 10000
	rs := sup(r, 3)
	n := fmt.Sprintf("%s%s%s", domains, ms, rs)
	return n
}

// 对长度不足n的数字前面补0
func sup(i int64, n int) string {
	m := fmt.Sprintf("%d", i)
	for len(m) < n {
		m = fmt.Sprintf("0%s", m)
	}
	return m
}
