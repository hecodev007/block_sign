package registor

import (
	"avaxDataServer/common"
	"avaxDataServer/conf"
	"avaxDataServer/services"
	"avaxDataServer/services/avax" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"avax": avax.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"avax": avax.NewProcessor,
	}
}
