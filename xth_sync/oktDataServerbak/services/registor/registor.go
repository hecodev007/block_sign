package registor

import (
	"oktDataServer/common"
	"oktDataServer/common/conf"
	"oktDataServer/services"
	btc "oktDataServer/services/okt" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"okt": btc.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"okt": btc.NewProcessor,
	}
}
