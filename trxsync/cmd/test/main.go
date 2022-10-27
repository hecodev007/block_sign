package main

import (
	"flag"
	"fmt"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/trxsync/common"
	"github.com/group-coldwallet/trxsync/conf"
	"github.com/group-coldwallet/trxsync/db"
	"github.com/group-coldwallet/trxsync/router"
	"github.com/group-coldwallet/trxsync/services"
	"github.com/group-coldwallet/trxsync/utils"
	"net/http"
	"runtime"
	"time"
)

var (
	routineNum int
	cfgFile    string
	cfg        conf.Config
	scanMap    map[string]ScanFunc
)

type ScanFunc func(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner

func init() {
	scanMap = make(map[string]ScanFunc)
	//注册IScanner

}
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()
	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "", "set the toml conf file")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	// 1. 初始化log
	utils.InitLog()
	// 2。 初始化配置文件
	err := conf.LoadConfig(cfgFile, &cfg)
	if err != nil {
		panic(err)
	}
	// 3。 初始化db数据库
	err = db.InitSyncDB(cfg.DataBases["sync"])
	if err != nil {
		log.Panic(fmt.Sprintf("init sync db err %v", err))
		return
	}
	log.Info("ConnectDB sync db success ")
	//链接用户地址数据库
	err = db.InitUserDB(cfg.DataBases["user"])
	if err != nil {
		log.Panic("init user db err %v", err)
	}
	log.Info("ConnectDB user db success")

	//4。 加载监听地址
	watcher, err := common.NewWatchControl(cfg.Sync.Name)
	if err != nil {
		log.Panic("new watch control err %v", err)
	}
	log.Info("NewWatchControl success ")

	scan, ok := scanMap[cfg.Sync.Name]
	if !ok {
		log.Errorf("暂不支持该币种的数据服务：%s", cfg.Sync.Name)
		return
	}
	//5。 初始化扫块服务
	bs := services.NewBaseService(&cfg, scan(cfg, cfg.Nodes[cfg.Sync.Name]), watcher)
	bs.Init()
	bs.Start()

	// 6。初始化推送服务
	r, _, err := router.InitRouter(bs)
	if err != nil {
		log.Errorf("init router error: %v", err)
		return
	}
	//7. 启动http服务
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
