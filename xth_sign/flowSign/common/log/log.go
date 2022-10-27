package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	LogSplitMaxSize int //每个分割的文件多少M
	//LogSplitMaxBackups int   //只保留多少个log日志文件先隐藏
	LogSplitMaxAge   int  //保留几天
	LogSplitCompress bool //是否进行文件压缩
	OutputPaths      []string
	ErrorOutputPaths []string
}

//func sample() {
//	//_log,_ := zap.NewProduction(zap.AddCaller())
//	//defer _log.Sync()
//	InitLog()
//	Debug("debug,最灵繁的人也看不见自己的背脊")
//	Info("info,最困难的事情就是认识自己。")
//	Warn("warn,有勇气承担命运这才是英雄好汉")
//	Error("error,与肝胆人共事，无字句处读书。")
//	DPanic("dpanic,阅读使人充实，会谈使人敏捷，写作使人精确。")
//	//Panic("panic,最大的骄傲于最大的自卑都表示心灵的最软弱无力。")
//	//Fatal("fatal,自知之明是最难得的知识。")
//	Debugf("debugf,勇气通往天堂，怯懦通往地狱。")
//	Infof("infof,有时候读书是%s一种巧%s妙地避开思考%s的方法。","test","demo","done")
//	Warnf("warnf,阅读%s一切好书%s如同和过去%s最杰出的人谈话。","test","demo","done")
//	Errorf("errorf,越是%s没有本领%s的就越加%s自命不凡。","test","demo","done")
//	DPanicf("dpanicf,越是%s无能的人，%s越喜欢挑剔%s别人的错儿。","test","demo","done")
//	//Panicf("panicf,知人者智%s，自知者明%s。胜人者有力%s，自胜者强。","test","demo","done")
//	//Fatalf("fatalf,意志坚强%s的人能把%s世界放在手中像泥块%s一样任意揉捏。","test","demo","done")
//}

func InitLogger(mode, level, format, outfile, errfile string) {
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
		EncodeCaller:   AbusulteCallerEncoder,
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
	if mode != EnvProd {
		//日志都会在console中展示
		cores = append(cores, zapcore.NewCore(zapcore.NewConsoleEncoder(config),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), logLevel))
		options = append(options, zap.Development())
	}
	options = append(options, zap.AddCaller())
	options = append(options, zap.AddCallerSkip(1))
	options = append(options, zap.AddStacktrace(zap.WarnLevel))
	// 最后创建具体的Logger
	core := zapcore.NewTee(cores...)
	//options = append(options, zap.AddCallerSkip(1))
	log = zap.New(core, options...) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
}
func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
func AbusulteCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fullPath := caller.FullPath()
	index := strings.Index(fullPath, "Sign")
	if index < 0 {
		enc.AppendString(caller.TrimmedPath())
	} else {
		enc.AppendString(fullPath[index+5:])
	}
}

//目录不存在则创建
func checkDir(filepath string) {
	dir := path.Dir(filepath)
	os.MkdirAll(dir, os.ModePerm)
}

func getWriter(filename string) io.Writer {
	checkDir(filename)
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
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

/*func InitLogger(dir, name, format, level string) {
	baseLogPath := path.Join(dir, name)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(name),             // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(30*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Errorf("conf local file system logger error. %+v", err)
	}
	var f logrus.Formatter
	if format == "json" {
		f = &logrus.JSONFormatter{}
	} else {
		f = &logrus.TextFormatter{
			ForceColors: true,
		}
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer, // 为不同级别设置不同的输出目的
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, f)
	logrus.AddHook(lfHook)
}
*/
/*func InitLog(cfg *Logcfg) {
	var (
		level   zap.AtomicLevel
		wss     = make([]zapcore.WriteSyncer, 0)
		options = make([]zap.Option, 0)
		core    zapcore.Core
	)
	switch cfg.Level {
	case LvlDebug:
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case LvlInfo:
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case LvlWarn:
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case LvlError:
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case LvlDPanic:
		level = zap.NewAtomicLevelAt(zap.DPanicLevel)
	case LvlPanic:
		level = zap.NewAtomicLevelAt(zap.PanicLevel)
	case LvlFatal:
		level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	zapCfg := zapcore.EncoderConfig{
		TimeKey:        cfg.TimeKey,
		LevelKey:       cfg.LevelKey,
		NameKey:        cfg.NameKey,
		CallerKey:      cfg.CallerKey,
		MessageKey:     cfg.MessageKey,
		StacktraceKey:  cfg.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder, //zapcore.SecondsDurationEncoder,//zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if len(cfg.OutputPaths) > 0 {
		for _, op := range cfg.OutputPaths {
			switch op {
			case "stdout":
				wss = append(wss, zapcore.AddSync(os.Stdout))
			case "stderr":
				wss = append(wss, zapcore.AddSync(os.Stderr))
			default:
				//这里要是出错了就外部不知道
				if f, err := os.OpenFile(op, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
					wss = append(wss, zapcore.AddSync(f))
				}
			}
		}
	} else {
		if len(cfg.LogSplitPath) == 0 || cfg.LogSplitMaxSize <= 0 || cfg.LogSplitMaxAge <= 0 {
			wss = append(wss, zapcore.AddSync(os.Stdout))
		}
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		for _, op := range cfg.ErrorOutputPaths {
			switch op {
			case "stdout":
				wss = append(wss, zapcore.AddSync(os.Stdout))
			case "stderr":
				wss = append(wss, zapcore.AddSync(os.Stderr))
			default:
				//这里要是出错了就外部不知道
				if f, err := os.OpenFile(op, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
					wss = append(wss, zapcore.AddSync(f))
				}
			}
		}
	} else {
		if len(cfg.LogSplitPath) == 0 || cfg.LogSplitMaxSize <= 0 || cfg.LogSplitMaxAge <= 0 {
			wss = append(wss, zapcore.AddSync(os.Stderr))
		}
	}

	if len(cfg.LogSplitPath) > 0 && cfg.LogSplitMaxSize > 0 && cfg.LogSplitMaxAge > 0 {
		wss = append(wss, zapcore.AddSync(&lumberjack.Logger{
			Filename: cfg.LogSplitPath,    //分割⽇志⽂件路径
			MaxSize:  cfg.LogSplitMaxSize, //分割日志文件容量megabytes
			//MaxBackups: 3,                    //最多保留3个备份
			MaxAge:   cfg.LogSplitMaxAge,   //分割days
			Compress: cfg.LogSplitCompress, //是否压缩disabled by default
		}))
	}

	if cfg.Env != EnvProd {
		options = append(options, zap.Development())
	}
	options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewSampler(core, time.Second, 100, 100)
	}))
	options = append(options, zap.AddCaller())
	options = append(options, zap.AddCallerSkip(1))

	if cfg.TxtType == TxtJson {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(zapCfg),
			zapcore.NewMultiWriteSyncer(wss...),
			level,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapCfg),
			zapcore.NewMultiWriteSyncer(wss...),
			level,
		)
	}

	log = zap.New(core, options...)
	//log,_ = zap.Config {
	//	Level:       zap.NewAtomicLevelAt(lvl),
	//	Development: isDev,
	//	Sampling: &zap.SamplingConfig {
	//		Initial:    100,
	//		Thereafter: 100,
	//	},
	//	Encoding:         txtType,
	//	EncoderConfig:    zapcore.EncoderConfig{
	//		TimeKey:        cfg.TimeKey,
	//		LevelKey:       cfg.Level,
	//		NameKey:        cfg.NameKey,
	//		CallerKey:      cfg.CallerKey,
	//		MessageKey:     cfg.MessageKey,
	//		StacktraceKey:  cfg.StacktraceKey,
	//		LineEnding:     zapcore.DefaultLineEnding,
	//		EncodeLevel:    zapcore.LowercaseLevelEncoder,
	//		EncodeTime:     zapcore.ISO8601TimeEncoder,
	//		EncodeDuration: zapcore.StringDurationEncoder,//zapcore.SecondsDurationEncoder,//zapcore.StringDurationEncoder,
	//		EncodeCaller:   zapcore.ShortCallerEncoder,
	//	},
	//	OutputPaths:      outputPaths,
	//	ErrorOutputPaths: erroroutputPaths,
	//}.Build(zap.AddCaller(),zap.AddCallerSkip(1))
}*/
//func Debug(msg string) {
//	log.Debug(msg)
//}
//
//func Info(msg string) {
//	log.Info(msg)
//}
//
//func Warn(msg string) {
//	log.Warn(msg)
//}
//
//func Error(msg string) {
//	log.Error(msg)
//}
//
//func DPanic(msg string) {
//	log.DPanic(msg)
//}
//
//func Panic(msg string) {
//	log.Panic(msg)
//}
//
//func Fatal(msg string) {
//	log.Fatal(msg)
//}

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
