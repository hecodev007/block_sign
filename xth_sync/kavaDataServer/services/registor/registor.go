package registor

import (
	"kavaDataServer/common"
	"kavaDataServer/common/conf"
	"kavaDataServer/services"
	btc "kavaDataServer/services/atom" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"kava": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"kava": btc.NewProcessor,
	}
}
