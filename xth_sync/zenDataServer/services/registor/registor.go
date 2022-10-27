package registor

import (
	"zenDataServer/common"
	"zenDataServer/common/conf"
	"zenDataServer/services"
	btc "zenDataServer/services/zen" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"zen": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"zen": btc.NewProcessor,
	}
}
