package log

import (
	"fmt"
	"github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	log *zap.Logger
)

const (
	EnvDevelopment = "dev"
	EnvProd        = "prod"
	LvlDebug       = "debug"
	LvlInfo        = "info"
	LvlWarn        = "warn"
	LvlError       = "error"
	LvlDPanic      = "dpanic"
	LvlPanic       = "panic"
	LvlFatal       = "fatal"
	TxtJson        = "json"
	TxtConsole     = "console"
)

type Logcfg struct {
	Level           string
	Env             string
	TxtType         string
	TimeKey         string
	LevelKey        string
	NameKey         string
	CallerKey       string
	MessageKey      string
	StacktraceKey   string
	LogSplitPath    string
	LogSplitMaxSize int // 每个分割的文件多少M
	// LogSplitMaxBackups int   //只保留多少个log日志文件先隐藏
	LogSplitMaxAge   int  // 保留几天
	LogSplitCompress bool // 是否进行文件压缩
	OutputPaths      []string
	ErrorOutputPaths []string
}

func InitLogger(isDebug bool, level, format, outfile, errfile string) {
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
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 将级别转换成大写
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	// 设置日志格式
	if format == TxtJson {
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
	// 日志都会在console中展示
	cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(config),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), logLevel))
	options = append(options, zap.Development())

	options = append(options, zap.AddCaller())
	options = append(options, zap.AddCallerSkip(1))
	options = append(options, zap.AddStacktrace(zap.WarnLevel))
	// 最后创建具体的Logger
	core := zapcore.NewTee(cores...)
	// options = append(options, zap.AddCallerSkip(1))
	log = zap.New(core, options...) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
}

func getWriter(filename string) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	file := filename
	paths := strings.Split(filename, "/")
	if len(paths) > 0 {
		file = paths[len(paths)-1]
		filePath := strings.Replace(filename, file, "", 1)
		if !filepath.IsAbs(filePath) {
			filePath, _ = filepath.Abs(filePath)
		}
		if err := os.MkdirAll(filePath, 0700); err != nil {
			panic("Failed to create logger folder:" + filePath + ". err:" + err.Error())
		}
	}

	hook, err := rotatelogs.New(
		filename+".%Y%m%d", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(file),
		rotatelogs.WithMaxAge(time.Hour*24*30),    // 保存30天
		rotatelogs.WithRotationTime(time.Hour*24), // 切割频率 24小时
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func Debug(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Info(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Warn(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Error(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func DPanic(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.DPanic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Panic(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Fatal(args ...interface{}) {
	if log == nil || len(args) == 0 {
		return
	}
	log.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Debugf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Debug(msg)
}

func Infof(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Info(msg)
}

func Warnf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Warn(msg)
}

func Errorf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Error(msg)
}

func DPanicf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.DPanic(msg)
}

func Panicf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Panic(msg)
}

func Fatalf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	msg := format
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Fatal(msg)
}

func Sync() {
	if log == nil {
		return
	}
	log.Sync()
}

func Debugw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Debug(msg, fs...)
		return
	}
	log.Debug(msg)
}

func Infow(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Info(msg, fs...)
		return
	}
	log.Info(msg)
}

func Warnw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Warn(msg, fs...)
		return
	}
	log.Warn(msg)
}

func Errorw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Error(msg, fs...)
		return
	}
	log.Error(msg)
}

func DPanicw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.DPanic(msg, fs...)
		return
	}
	log.DPanic(msg)
}

func Panicw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Panic(msg, fs...)
		return
	}
	log.Panic(msg)
}

func Fatalw(msg string, fields map[string]interface{}) {
	if log == nil {
		return
	}
	if len(fields) > 0 {
		fs := make([]zap.Field, 0, len(fields))
		for k, v := range fields {
			fs = append(fs, zap.Any(k, v))
		}
		log.Fatal(msg, fs...)
		return
	}
	log.Fatal(msg)
}
