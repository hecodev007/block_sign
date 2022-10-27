package util

import (
	"fmt"
	"time"
)

const Layout = "2006-01-02 15:04:05" //时间常量

func GetChinaTimeNow() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timeStr := time.Now().Format(Layout)
	timeChina, _ := time.ParseInLocation(Layout, timeStr, loc)
	return timeChina
}

func GetChinaTimeNowFormat() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timeStr := time.Now().Format(Layout)
	timeChina, _ := time.ParseInLocation(Layout, timeStr, loc)
	return timeChina.Format(Layout)
}

func UtcFormatChainTime(utcStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		fmt.Println(11111)
		return "", err
	}
	utcTime := t.In(time.Local)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	timeStr := utcTime.Format(Layout)
	timeChina, err := time.ParseInLocation(Layout, timeStr, loc)
	if err != nil {
		return "", err
	}
	return timeChina.Format(Layout), nil
}

func TimeStrFormatToUnix(timeStr string) (int64, error) {
	t, err := time.Parse(Layout, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil

}

func UnixFormatToStr(unix int64) (string, error) {
	t := time.Unix(unix, 0)
	timeStr := t.Format(Layout)
	return timeStr, nil
}
