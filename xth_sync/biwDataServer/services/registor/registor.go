package registor

import (
	"biwDataServer/common"
	"biwDataServer/common/conf"
	"biwDataServer/services"
	btc "biwDataServer/services/biw" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"biw": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"biw": btc.NewProcessor,
	}
}
