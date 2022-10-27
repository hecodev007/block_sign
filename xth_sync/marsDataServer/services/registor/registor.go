package registor

import (
	"marsDataServer/common"
	"marsDataServer/common/conf"
	"marsDataServer/services"
	"marsDataServer/services/ghost" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"mars": ghost.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"mars": ghost.NewProcessor,
	}
}
