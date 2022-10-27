package main

import (
	"custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/router"
	"custody-merchant-admin/runtime/job"
	"flag"
	"fmt"
	"github.com/bamzi/jobrunner"
	"os"
)

var (
	confFilePath string
	cmdHelp      bool
)

// 初始化
func init() {
	ph, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("配置文件路径找不到"))
	}
	// 配置配置文件所在位置
	flag.StringVar(&confFilePath, "c", ph+"/"+config.DefaultConfigFile, "配置文件路径")
	// 赋予bool
	flag.BoolVar(&cmdHelp, "h", false, "帮助")
	flag.Parse()

}

func main() {
	if cmdHelp {
		// 输出默认值
		flag.PrintDefaults()
		return
	}
	// 日志打印
	log.Debugf("run with conf:%s", confFilePath)
	// 子域名部署
	router.RunSubdomains(confFilePath)

	jobRunner()
}

//定时任务开启
func jobRunner() {
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	////定时更新币列表
	jobrunner.Schedule("@every 0.5h", job.WalletCoinListCallBackJob{})
	j := job.WalletCoinListCallBackJob{}
	j.Run()
}
