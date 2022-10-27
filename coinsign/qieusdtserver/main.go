package main

import (
	"flag"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/api"
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/group-coldwalle/coinsign/qieusdtserver/server"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service"
	"github.com/group-coldwalle/coinsign/qieusdtserver/service/usdtfile"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	log "github.com/sirupsen/logrus"
	//"github.com/group-coldwalle/coinsign/qieusdtserver/api"
	//"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"runtime"
)

var (
	v  bool
	pm int
	c  string
)

var (
	cfg *config.GlobalConfig
)

func GetConfig() *config.GlobalConfig {
	return cfg
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	parseFlag()
	deploySys()
	loadConfig()
	initLog()
	initService()
	intCommand()
	intApi()

	//加载以前历史遗留的旧文件
	//config.ReadOldeCsv("./files/old.csv")
	config.ReadOldeCsv(cfg.OldAddressFile)

	//加载新版本生成文件目录
	usdtfile.ReadNewFolder(cfg.LoadAddressPath)
	startServer()
}

//获取系统参数
func parseFlag() {
	flag.BoolVar(&v, "v", false, "build info")
	flag.IntVar(&pm, "pm", 10, "each cpu's routine num")
	flag.StringVar(&c, "c", "", "set the yaml config file")
	flag.Parse()
}

//配置系统运行参数
func deploySys() {
	if v {
		//这里写死,等项目以后完善这个信息
		fmt.Printf("version 1.0")
		return
	}
	//限制goroutine在合理的范围
	if pm <= 0 || pm > 10 {
		pm = 4
	}
	runtime.GOMAXPROCS(runtime.NumCPU() * pm)
}

//读入配置文件
func loadConfig() {
	var err error
	if cfg, err = config.LoadConfig(c); err != nil {
		panic(err)
	}
}

//读入配置文件
func initLog() {
	util.ConfigLogger(cfg.LogCfg)
}

func initService() {
	service.InitTranscation(cfg)
}

func intCommand() {
	err := usdtfile.SetDefaultFileConfig(cfg)
	if err != nil {
		panic(err)
	}
}

func intApi() {
	api.SetConfig(cfg)
}

//运行api 服务器
func startServer() {
	server.Run(cfg)
}

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o=usdt_linux
//go build -ldflags "-s -w"  -o=usdt_mac
