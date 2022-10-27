package registor

import (
	"hdxsync/common"
	"hdxsync/common/conf"
	"hdxsync/services"
	btc "hdxsync/services/dot" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"hdx": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"hdx": btc.NewProcessor,
	}
}
