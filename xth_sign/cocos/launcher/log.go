package launcher

import (
	"github.com/astaxie/beego/logs"
)

func CustomLogInfo() {
	logs.SetLogger(logs.AdapterConsole)
	logs.SetLogger(logs.AdapterFile, `{"filename":"debug.log","level":6,"maxlines":1000,"maxsize":1000,"daily":true,"maxdays":10,"color":true}`)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
}
