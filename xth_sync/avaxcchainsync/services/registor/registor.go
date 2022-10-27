package registor

import (
	"avaxcchainsync/common"
	"avaxcchainsync/common/conf"
	"avaxcchainsync/services"
	btc "avaxcchainsync/services/wtc" //导入执行其init
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
