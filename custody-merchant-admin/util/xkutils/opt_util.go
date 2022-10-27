package xkutils

import (
	"time"
)

type OptsToDo struct {
	Block func()
}

func ThreeDo(b bool, v1, v2 interface{}) interface{} {
	if b {
		return v1
	} else {
		return v2
	}
}

func (opt OptsToDo) NilHave(b bool, block func()) OptsToDo {
	if b {
		block()
	}
	return opt
}

//判断时间是当年的第几周

func WeekByDate(t time.Time) int {

	yearDay := t.YearDay()

	yearFirstDay := t.AddDate(0, 0, -yearDay+1)

	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天

	firstWeekDays := 1

	if firstDayInWeek != 0 {

		firstWeekDays = 7 - firstDayInWeek + 1

	}

	var week int

	if yearDay <= firstWeekDays {

		week = 1

	} else {

		week = (yearDay-firstWeekDays)/7 + 2

	}

	return week

}

type WeekDate struct {
	WeekTh string

	StartTime time.Time

	EndTime time.Time
}
