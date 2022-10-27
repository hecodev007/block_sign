package registor

import (
	"stxDataServer/common"
	"stxDataServer/common/conf"
	"stxDataServer/services"
	btc "stxDataServer/services/stx" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"stx": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"stx": btc.NewProcessor,
	}
}
