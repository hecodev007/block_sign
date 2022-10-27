package registor

import (
	"cfxDataServer/common"
	"cfxDataServer/common/conf"
	"cfxDataServer/services"
	btc "cfxDataServer/services/cfx" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"mob": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"mob": btc.NewProcessor,
	}
}
