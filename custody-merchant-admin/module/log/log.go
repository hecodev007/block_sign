package log

/**
 *
 * 日志文件
 *
 */

import (
	. "custody-merchant-admin/config"
	"fmt"
	l "github.com/labstack/gommon/log"
	"os"
	"time"
)

// Logger 结构体
type Logger struct {
	*l.Logger // 输出Logger的结构体的值
}

var (
	// 自定义全局log
	// 使用gommon/log的global由于限制3层调用栈获取不到log的准确file路径
	global = l.New("log")
	// 用于日志标头的设置，默认值为
	defaultHeader = `{"time":"${time_rfc3339}","level":"${level}","prefix":"${prefix}",` +
		`"files":"${long_file}","line":"${line}"}`
	logFile = new(os.File)
)

// init 初始化
func init() {
	// 全局log属性设置
	l.SetHeader(defaultHeader)
	l.SetLevel(l.DEBUG)
	// 自定义全局logs属性设置
	global.SetHeader(defaultHeader)
	global.SetLevel(l.DEBUG)
}

func newLog() {
	var err error
	st := time.Now().Local().Format("2006-01-02")
	err = os.Mkdir(Conf.LogFile, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	fls := fmt.Sprintf("%s/log-%s.log", Conf.LogFile, st)
	logFile, err = os.OpenFile(fls, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log files failed, err:", err)
		return
	}
	global.SetOutput(logFile)
}

func CreateLogFile() {
	logNewJobsSchedule(func() { newLog() }, time.Hour*24)
}

func logNewJobsSchedule(backCall func(), duration time.Duration) {
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

func SetLevel(v l.Lvl) {
	l.SetLevel(v)
	global.SetLevel(v)
	CreateLogFile()
}

func Debug(i ...interface{}) {
	global.Debug(i)
}

func Debugf(format string, values ...interface{}) {
	global.Debugf(format, values...)
}

func Info(i ...interface{}) {
	global.Info(i)
}

func Infof(format string, values ...interface{}) {
	global.Infof(format, values...)
}

func Warn(i ...interface{}) {
	global.Warn(i)
}

func Warnf(format string, values ...interface{}) {
	global.Warnf(format, values...)
}

func Error(i ...interface{}) {

	global.Error(i)
}

func Errorf(format string, values ...interface{}) {
	global.Errorf(format, values...)
}

func Fatal(i ...interface{}) {
	global.Fatal(i)
}

func Fatalf(format string, values ...interface{}) {
	global.Fatalf(format, values...)
}

func Panic(i ...interface{}) {
	global.Panic(i)
}

func Panicf(format string, args ...interface{}) {
	global.Panicf(format, args)
}
