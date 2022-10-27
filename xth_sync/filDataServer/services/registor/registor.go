package registor

import (
	"filDataServer/common"
	"filDataServer/common/conf"
	"filDataServer/services"
	"filDataServer/services/fil"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"fil": fil.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"fil": fil.NewProcessor,
	}
}
