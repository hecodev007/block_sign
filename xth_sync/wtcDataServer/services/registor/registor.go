package registor

import (
	"wtcDataServer/common"
	"wtcDataServer/common/conf"
	"wtcDataServer/services"
	btc "wtcDataServer/services/wtc" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"wtc": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"wtc": btc.NewProcessor,
	}
}
