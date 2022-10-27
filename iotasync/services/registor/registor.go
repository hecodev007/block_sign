package registor

import (
	"iotasync/common"
	"iotasync/common/conf"
	"iotasync/services"
	"iotasync/services/iota"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"iota": iota.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"iota": iota.NewProcessor,
	}
}
