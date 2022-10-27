package registor

import (
	"starDataServer/common"
	"starDataServer/common/conf"
	"starDataServer/services"
	"starDataServer/services/fil"
)

type ScanFunc func(conf.Config, conf.NodeConfig) common.Scanner
type ProcFunc func(conf.Config, conf.NodeConfig, *services.WatchControl) common.Processor

var ScanFuncMap map[string]ScanFunc
var ProcFuncMap map[string]ProcFunc

func init() {
	ScanFuncMap = map[string]ScanFunc{
		"star": fil.NewScanner,
	}

	ProcFuncMap = map[string]ProcFunc{
		"star": fil.NewProcessor,
	}
}
