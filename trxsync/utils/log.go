package utils

import (
	"github.com/group-coldwallet/common/log"
)

func InitLog() {
	logCfg := &log.Logcfg{
		Level:           log.LvlDebug,
		Env:             log.EnvDevelopment,
		TxtType:         log.TxtJson,
		TimeKey:         "time",
		LevelKey:        "level",
		NameKey:         "logger",
		CallerKey:       "caller",
		MessageKey:      "msg",
		StacktraceKey:   "stacktrace",
		LogSplitPath:    "",
		LogSplitMaxSize: 1,
		//LogSplitMaxBackups int   //只保留多少个log日志文件先隐藏
		LogSplitMaxAge:   7,
		LogSplitCompress: false,
		OutputPaths:      []string{"stdout", ""},
		ErrorOutputPaths: []string{""},
		//OutputPaths:  nil,
		//ErrorOutputPaths: nil,
	}
	log.InitLog(logCfg)
}
