package registor

import (
	"solsync/common"
	"solsync/common/conf"
	"solsync/services"
	"solsync/services/ccn"
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"ccn": ccn.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"ccn": ccn.NewProcessor,
	}
}
