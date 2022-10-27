package service

import (
	"custody-merchant-admin/global"
	"fmt"
	"strconv"
	"time"
)

func GetIntFromInterface(in interface{}) (out int64) {
	switch in.(type) {
	case float64:
		f := in.(float64)
		out = int64(f)
	case string:
		f := in.(string)
		out, _ = strconv.ParseInt(f, 10, 64)
	case int64:
		out = in.(int64)
	case int:
		f := in.(int)
		out = int64(f)
	}
	return
}

func GetFloat64FromInterface(in interface{}) (out float64) {
	switch in.(type) {
	case float64:
		out = in.(float64)
	case string:
		f := in.(string)
		out, _ = strconv.ParseFloat(f, 64)
	case int64:
		f := in.(int64)
		out = float64(f)
	case int:
		f := in.(int)
		out = float64(f)
	}
	return
}

func TimeFromString(str string) (t time.Time) {
	if len(str) == 10 {
		t, _ = time.ParseInLocation(global.YyyyMmDd, str, time.Local)
	} else if len(str) == 19 {
		t, _ = time.ParseInLocation(global.YyyyMmDdHhMmSs, str, time.Local)
	}
	return t
}

//SubMonth 计算月份差
func SubMonth(t1, t2 time.Time) (month int) {
	y1 := t1.Year()
	y2 := t2.Year()
	m1 := int(t1.Month())
	m2 := int(t2.Month())
	d1 := t1.Day()
	d2 := t2.Day()

	yearInterval := y1 - y2
	// 如果 d1的 月-日 小于 d2的 月-日 那么 yearInterval-- 这样就得到了相差的年数
	if m1 < m2 || m1 == m2 && d1 < d2 {
		yearInterval--
	}
	// 获取月数差值
	monthInterval := (m1 + 12) - m2
	if d1 < d2 {
		monthInterval--
	}
	monthInterval %= 12
	month = yearInterval*12 + monthInterval
	return
}

func OperateName(s string) (name string) {
	switch s {
	case "create":
		name = "创建"
	case "delete":
		name = "删除"
	case "update":
		name = "更新"
	case "lock":
		name = "冻结"
	case "unlock":
		name = "解冻"
	default:
		name = s
	}
	return
}
func GetTimeString(t time.Time) string {
	if !TimeIsNull(t) {
		return t.Format(global.YyyyMmDdHhMmSs)
	}
	return ""
}

func VerifyStatus(s string) (name string) {
	switch s {
	case "agree":
		name = "已通过"
	case "refuse":
		name = "已否决"
	case "lock":
		name = "冻结"
	case "unlock":
		name = "解冻"
	default:
		name = "待审核"
	}
	return
}

func TimeIsNull(t time.Time) bool {
	nullTime := time.Time{}
	return nullTime.Equal(t)
}

func GetStringFromInterfaceArr(arr []interface{}) string {
	var str string
	for _, s := range arr {
		str = fmt.Sprintf("%v,%v", str, s)
	}
	if len(str) > 1 {
		str = str[1:]
	}
	return str
}
