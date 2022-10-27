package registor

import (
	"dotDataServer/common"
	"dotDataServer/common/conf"
	"dotDataServer/services"
	btc "dotDataServer/services/dot" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"dot": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"dot": btc.NewProcessor,
	}
}
