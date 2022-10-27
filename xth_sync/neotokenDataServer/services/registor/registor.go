package registor

import (
	"neotokenDataServer/common"
	"neotokenDataServer/common/conf"
	"neotokenDataServer/services"
	"neotokenDataServer/services/neo" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"neo": neo.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"neo": neo.NewProcessor,
	}
}
