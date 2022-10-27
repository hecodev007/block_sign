package registor

import (
	"moacDataServer/common"
	"moacDataServer/common/conf"
	"moacDataServer/services"
	btc "moacDataServer/services/wtc" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"moac": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"moac": btc.NewProcessor,
	}
}
