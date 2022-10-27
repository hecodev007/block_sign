package registor

import (
	"ghostDataServerNew/common"
	"ghostDataServerNew/common/conf"
	"ghostDataServerNew/services"
	"ghostDataServerNew/services/ghost" //导入执行其init
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"ghost": ghost.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"ghost": ghost.NewProcessor,
	}
}
