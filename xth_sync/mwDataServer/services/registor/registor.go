package registor

import (
	"mwDataServer/common"
	"mwDataServer/common/conf"
	"mwDataServer/services"
	btc "mwDataServer/services/atp" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"mw": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"mw": btc.NewProcessor,
	}
}
