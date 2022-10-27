package registor

import (
	"dogeDataServer/common"
	"dogeDataServer/common/conf"
	"dogeDataServer/services"
	"dogeDataServer/services/doge" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"doge": doge.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"doge": doge.NewProcessor,
	}
}
