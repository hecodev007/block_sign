package registor

import (
	"algoDataServer/common"
	"algoDataServer/common/conf"
	"algoDataServer/services"
	btc "algoDataServer/services/atp" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"algo": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"algo": btc.NewProcessor,
	}
}
