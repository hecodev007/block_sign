package util

import (
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"path"
	"time"
)

// config logrus log to local filesystem, with file rotation
func ConfigLogger(cfg config.LogConfig) {
	baseLogPath := path.Join(cfg.LogPath, cfg.LogName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(24*time.Hour),       // 文件最大保存时间
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", err)
	}
	var f log.Formatter
	if cfg.Formatter == "json" {
		f = &log.JSONFormatter{}
	} else {
		f = &log.TextFormatter{}
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer, // 为不同级别设置不同的输出目的
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, f)
	log.AddHook(lfHook)
}
