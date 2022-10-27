package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

const (
	EnvProd = "prod"
	TxtJson = "json"
)

func init() {
	logger, _ = zap.NewDevelopment()
}

func InitLogger(isTerminal bool, level, format, outfile, errfile string) {
	var (
		encoder zapcore.Encoder
		options = make([]zap.Option, 0)
		cores   = make([]zapcore.Core, 0)
	)
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	config := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		CallerKey:      "caller",
		TimeKey:        "time",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, //将级别转换成大写
		EncodeTime:     SyslogTimeEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, //zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	// 设置日志格式
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(config)
	} else {
		encoder = zapcore.NewConsoleEncoder(config)
	}
	// 设置级别
	logLevel := zap.DebugLevel
	switch level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		logLevel = zap.InfoLevel
	}
	// 实现两个判断日志等级的interface  可以自定义级别展示
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel && lvl >= logLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel && lvl >= logLevel
	})

	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter(outfile)
	warnWriter := getWriter(errfile)
	// 将info及以下写入logPath,  warn及以上写入errPath
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel))
	cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(warnWriter), warnLevel))
	//
	if isTerminal {
		//日志都会在console中展示
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), infoLevel))
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), warnLevel))
		options = append(options, zap.Development())
	}
	options = append(options, zap.AddCaller())
	options = append(options, zap.AddCallerSkip(1))
	options = append(options, zap.AddStacktrace(zap.WarnLevel))
	// 最后创建具体的Logger
	core := zapcore.NewTee(cores...)
	//options = append(options, zap.AddCallerSkip(1))
	logger = zap.New(core, options...) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
}

//caller位置为相对项目路径地址，打印出来能够跳转，方便调试
func AbusulteCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fullPath := caller.FullPath()
	index := strings.Index(fullPath, "sync")
	if index < 0 {
		enc.AppendString(caller.TrimmedPath())
	} else {
		enc.AppendString(fullPath[index+5:])
	}
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func getWriter(filename string) io.Writer {
	//文件夹是否存在，不存则在创建

	dir := path.Dir(filename)
	os.MkdirAll(dir, os.ModePerm)
	// 生成rotatelogs的Logger 实际生成的文件名 demo.logger.YYmmddHH
	// demo.log是指向最新日志的链接
	fileabs, err := filepath.Abs(filename)
	if err != nil {
		panic(err.Error())
	}
	hook, err := rotatelogs.New(
		fileabs+".%Y%m%d", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*30),    // 保存30天
		rotatelogs.WithRotationTime(time.Hour*24), //切割频率 24小时
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func Debug(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Info(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Warn(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Error(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Error(strings.TrimRight(fmt.Sprintln(args...), "\n") + "\n" + string(debug.Stack()) + "\n")
}

func DPanic(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.DPanic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Panic(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Fatal(args ...interface{}) {
	if logger == nil || len(args) == 0 {
		return
	}
	logger.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Debugf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Debug(msg)
}

func Infof(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Info(msg)
}

func Warnf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Warn(msg)
}

func Errorf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Error(msg)
}

func DPanicf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.DPanic(msg)
}

func Panicf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Panic(msg)
}

func Fatalf(format string, args ...interface{}) {
	if logger == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	logger.Fatal(msg)
}

func Sync() {
	if logger == nil {
		return
	}
	logger.Sync()
}

func Debugw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Debug(msg, fs...)
		return
	}
	logger.Debug(msg)
}

func Infow(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Info(msg, fs...)
		return
	}
	logger.Info(msg)
}

func Warnw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Warn(msg, fs...)
		return
	}
	logger.Warn(msg)
}

func Errorw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Error(msg, fs...)
		return
	}
	logger.Error(msg)
}

func DPanicw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.DPanic(msg, fs...)
		return
	}
	logger.DPanic(msg)
}

func Panicw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Panic(msg, fs...)
		return
	}
	logger.Panic(msg)
}

func Fatalw(msg string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		logger.Fatal(msg, fs...)
		return
	}
	logger.Fatal(msg)
}
