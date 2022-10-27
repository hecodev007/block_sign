package registor

import (
	"bchaDataServer/common"
	"bchaDataServer/common/conf"
	"bchaDataServer/services"
	btc "bchaDataServer/services/biw" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"bcha": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"bcha": btc.NewProcessor,
	}
}
