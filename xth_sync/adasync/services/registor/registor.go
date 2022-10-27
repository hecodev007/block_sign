package registor

import (
	"adasync/common"
	"adasync/common/conf"
	"adasync/services"
	btc "adasync/services/biw" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		conf.Cfg.Sync.Name: btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		conf.Cfg.Sync.Name: btc.NewProcessor,
	}
}
