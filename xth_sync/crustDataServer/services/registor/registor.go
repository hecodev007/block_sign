package registor

import (
	"crustDataServer/common"
	"crustDataServer/common/conf"
	"crustDataServer/services"
	"crustDataServer/services/crust"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"crust": crust.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"crust": crust.NewProcessor,
	}
}
