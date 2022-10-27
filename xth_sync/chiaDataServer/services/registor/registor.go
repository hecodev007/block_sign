package registor

import (
	"chiaDataServer/common"
	"chiaDataServer/common/conf"
	"chiaDataServer/services"
	btc "chiaDataServer/services/yotta" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"yta": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"yta": btc.NewProcessor,
	}
}
