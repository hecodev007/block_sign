package registor

import (
	"atomDataServer/common"
	"atomDataServer/common/conf"
	"atomDataServer/services"
	btc "atomDataServer/services/atom" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"atom": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"atom": btc.NewProcessor,
	}
}
