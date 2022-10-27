package registor

import (
	"nyzoDataServer/common"
	"nyzoDataServer/common/conf"
	"nyzoDataServer/services"
	btc "nyzoDataServer/services/nyzo" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"nyzo": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"nyzo": btc.NewProcessor,
	}
}
