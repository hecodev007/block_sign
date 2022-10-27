package registor

import (
	"xrpDataServer/common"
	"xrpDataServer/common/conf"
	"xrpDataServer/services"
	"xrpDataServer/services/xrp" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"xrp": xrp.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"xrp": xrp.NewProcessor,
	}
}
