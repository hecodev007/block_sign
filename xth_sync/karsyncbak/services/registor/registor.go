package registor

import (
	"karsync/common"
	"karsync/common/conf"
	"karsync/services"
	btc "karsync/services/dot" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"kar-kar": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"kar-kar": btc.NewProcessor,
	}
}
