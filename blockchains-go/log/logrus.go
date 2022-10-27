package log

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	lg *logger
)

type logger struct {
	*logrus.Logger
}

type hook struct {
	caller  string
	traceID string
	app     string
	skip    int
	levels  []logrus.Level
}

func init() {
	writers := []io.Writer{os.Stdout}
	lg = &logger{}
	lg.Logger = newJSONLogger("info")
	lg.Logger.Out = io.MultiWriter(writers...)
}

// New 会初始化一个输出到指定文件的日志对象
// 同时也会输出到控制台
func New(level, path, fileName string, maxAgeHour int64) {
	if path == "" {
		path = "./logs" // 默认目录
	}
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		panic("Failed to create logger folder:" + path + ". err:" + err.Error())
	}

	if fileName == "" {
		fileName = "unknown"
	}

	if maxAgeHour <= 0 {
		maxAgeHour = int64(24 * 30)
	}
	filePath := path + "/" + fileName + "-%Y%m%d.log"
	linkPath := path + "/" + fileName + ".log"

	fileWriter, err := rotatelogs.New(
		filePath,
		rotatelogs.WithLinkName(linkPath),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour), // 日志按照每一天来切割
		rotatelogs.WithMaxAge(time.Duration(maxAgeHour)*time.Hour),
	)
	if err != nil {
		panic("Failed to create rotate logs. err:" + err.Error())
	}

	// 在指定的目录输出日志
	// 同时也在控制台输出
	writers := []io.Writer{
		fileWriter,
		os.Stdout}

	lg = &logger{}
	lg.Logger = newJSONLogger(level)
	lg.Logger.Out = io.MultiWriter(writers...)
}

// 使用自定义的hook来塞入caller、app、和tracerId
func (h *hook) Fire(entry *logrus.Entry) error {
	caller := findCaller(h.skip)
	entry.Data[h.caller] = caller
	return nil
}

// 可以使用hook的日志级别
func (h *hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewHook(levels ...logrus.Level) logrus.Hook {
	hook := hook{
		caller: "caller",
		skip:   5,
		levels: levels,
	}
	if len(hook.levels) == 0 {
		hook.levels = logrus.AllLevels
	}
	return &hook
}

func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	n := 0
	// 获取执行代码的文件名
	for i := len(file) - 1; i > 0; i-- {
		if string(file[i]) == "/" {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line
}

func findCaller(skip int) string {
	file := ""
	line := 0
	for i := 0; i < 10; i++ {
		file, line = getCaller(skip + i)
		// 过滤掉logrus的包和当前文件程序
		if !strings.HasPrefix(file, "logrus") && !strings.HasPrefix(file, "log/logrus") {
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func NewLogger(level logrus.Level, format logrus.Formatter, hook logrus.Hook) *logrus.Logger {
	log := logrus.New()
	log.Level = level
	log.SetFormatter(format)
	log.Hooks.Add(hook)
	return log
}

func newJSONLogger(level string) *logrus.Logger {
	return NewLogger(getLevel(level), &logrus.JSONFormatter{
		TimestampFormat:   "2006-01-02T15:04:05.000+0800", // 指定的时间格式
		DisableHTMLEscape: true,
	}, NewHook())
}

func formatPrint(format string, args ...interface{}) string {
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return msg
}

func Tracef(format string, args ...interface{}) {
	Trace(formatPrint(format, args...))
}

func Debugf(format string, args ...interface{}) {
	Debug(formatPrint(format, args...))
}

func Infof(format string, args ...interface{}) {
	Info(formatPrint(format, args...))
}

func Errorf(format string, args ...interface{}) {
	Error(formatPrint(format, args...))
}

func Printf(format string, args ...interface{}) {
	Print(formatPrint(format, args...))
}

func Warnf(format string, args ...interface{}) {
	Warn(formatPrint(format, args...))
}

func Warningf(format string, args ...interface{}) {
	Warning(formatPrint(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	Fatal(formatPrint(format, args...))
}

func Panicf(format string, args ...interface{}) {
	Panic(formatPrint(format, args...))
}

func Info(args ...interface{}) {
	lg.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Trace(args ...interface{}) {
	lg.Trace(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Debug(args ...interface{}) {
	lg.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Print(args ...interface{}) {
	lg.Print(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Warn(args ...interface{}) {
	lg.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Warning(args ...interface{}) {
	lg.Warning(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Error(args ...interface{}) {
	lg.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Fatal(args ...interface{}) {
	lg.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Panic(args ...interface{}) {
	lg.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func getLevel(level string) logrus.Level {
	if level != "" {
		level = strings.ToLower(level)
	}
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}
