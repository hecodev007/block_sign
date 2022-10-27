package common

import (
	"fmt"
	"strconv"
	"time"
)

// 获取毫秒UTC时间
func GetMillTime() int64 {
	timestamp := time.Now().UnixNano() / 1000000
	return timestamp
}

func Int64ToTime(timestamp int64) time.Time {
	ts := fmt.Sprintf("%d", timestamp)
	if len(ts) > 10 {
		ts = ts[:10]
	}
	i, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(i, 0)
}
