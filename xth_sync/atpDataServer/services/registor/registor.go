package registor

import (
	"atpDataServer/common"
	"atpDataServer/common/conf"
	"atpDataServer/services"
	btc "atpDataServer/services/atp" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"atp": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"atp": btc.NewProcessor,
	}
}
