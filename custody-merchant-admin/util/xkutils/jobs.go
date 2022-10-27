package xkutils

import "time"

func NewJobsSchedule(backCall func(), duration time.Duration) {
	go func() {
		for {
			backCall()
			nowTime := time.Now()
			// 计算下一个零点
			next := nowTime.Add(duration)
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location())
			t := time.NewTimer(next.Sub(nowTime))
			<-t.C
		}
	}()
}
