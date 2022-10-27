package registor

import (
	"domsync/common"
	"domsync/common/conf"
	"domsync/services"
	"domsync/services/dom"
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		//conf.Cfg.Sync.Name: brise.NewScanner,
		conf.Cfg.Sync.Name: dom.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		//conf.Cfg.Sync.Name: brise.NewProcessor,
		conf.Cfg.Sync.Name: dom.NewProcessor,
	}
}
