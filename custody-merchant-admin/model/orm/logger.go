package orm

import (
	"custody-merchant-admin/module/log"
)

type Logger struct {
}

// Print format & print log
func (logger Logger) Print(values ...interface{}) {
	// @TODO
	// 日志格式化解析
	log.Debugf("orm log:%v \n", values)
}
