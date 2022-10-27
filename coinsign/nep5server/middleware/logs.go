package middleware

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

//gin log中间件设置,使用第三方库logrus
func GinLogger(logPath string, logSPath string, logName string, loglevel string) gin.HandlerFunc {
	//MkdirAll会创建一个名为path的目录以及任何必要的父项，并返回nil，否则返回错误。
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	baseLogPath := path.Join(logPath,
		logName)
	sLogPath := path.Join(logSPath,
		logName)
	//设定log模板
	log.SetFormatter(&log.JSONFormatter{})
	//log.SetFormatter(joonix.NewFormatter())
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(sLogPath),         // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(15*24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", err)
	}

	//log.SetFormatter(&log.TextFormatter{})
	switch level := loglevel; level {
	/*
		如果日志级别不是debug就不要打印日志到控制台了
	*/
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetOutput(os.Stderr)
	case "info":
		setNull()
		log.SetLevel(log.InfoLevel)
	case "warn":
		setNull()
		log.SetLevel(log.WarnLevel)
	case "error":
		setNull()
		log.SetLevel(log.ErrorLevel)
	default:
		setNull()
		log.SetLevel(log.InfoLevel)
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer,
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.JSONFormatter{})
	log.AddHook(lfHook)

	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		end := time.Now()
		//执行时间
		latency := end.Sub(start)

		path := c.Request.URL.Path

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// 这里是指定日志打印出来的格式。分别是状态码，执行时间,请求ip,请求方法,请求路由
		log.Infof("| %3d | %13v | %15s | %s | %s ",
			statusCode,
			latency,
			clientIP,
			method, path,
		)
	}
}

func setNull() {
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}
	writer := bufio.NewWriter(src)
	log.SetOutput(writer)
}
