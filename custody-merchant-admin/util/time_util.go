package util

import (
	"fmt"
	"math"
	"time"
)

// WeekByDate
// 判断时间是当年的第几周
func WeekByDate(t time.Time) string {

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
	return fmt.Sprintf("%d-%d", t.Year(), week)
}

type WeekDate struct {
	WeekTh    string
	StartTime time.Time
	EndTime   time.Time
}

// GroupByWeekDate
// 将开始时间和结束时间分割为周为单位
func GroupByWeekDate(startTime, endTime time.Time) []WeekDate {

	weekDate := make([]WeekDate, 0)
	diffDuration := endTime.Sub(startTime)
	days := int(math.Ceil(float64(diffDuration/(time.Hour*24)))) + 1
	currentWeekDate := WeekDate{}
	currentWeekDate.WeekTh = WeekByDate(endTime)
	currentWeekDate.EndTime = endTime
	currentWeekDay := int(endTime.Weekday())

	if currentWeekDay == 0 {
		currentWeekDay = 7
	}

	currentWeekDate.StartTime = endTime.AddDate(0, 0, -currentWeekDay+1)
	nextWeekEndTime := currentWeekDate.StartTime
	weekDate = append(weekDate, currentWeekDate)

	for i := 0; i < (days-currentWeekDay)/7; i++ {
		weekData := WeekDate{}
		weekData.EndTime = nextWeekEndTime
		weekData.StartTime = nextWeekEndTime.AddDate(0, 0, -7)
		weekData.WeekTh = WeekByDate(weekData.StartTime)
		nextWeekEndTime = weekData.StartTime
		weekDate = append(weekDate, weekData)
	}

	if lastDays := (days - currentWeekDay) % 7; lastDays > 0 {
		lastData := WeekDate{}
		lastData.EndTime = nextWeekEndTime
		lastData.StartTime = nextWeekEndTime.AddDate(0, 0, -lastDays)
		lastData.WeekTh = WeekByDate(lastData.StartTime)
		weekDate = append(weekDate, lastData)
	}
	return weekDate
}
