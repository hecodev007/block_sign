package registor

import (
	"wdsync/common"
	"wdsync/common/conf"
	"wdsync/services"
	"wdsync/services/fil"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		conf.Cfg.Sync.Name: fil.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		conf.Cfg.Sync.Name: fil.NewProcessor,
	}
}
