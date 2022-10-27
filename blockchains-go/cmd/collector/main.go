package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/bamzi/jobrunner"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/runtime/job"
	"github.com/robfig/cron/v3"
)

type collectorFunc func(cfg conf.Collect2) cron.Job

var collectorFuncMap = map[string]collectorFunc{
	"eth":     job.NewCollectEthJob,
	"mof":     job.NewCollectMofJob,
	"klay":    job.NewCollectKlayJob,
	"seek":    job.NewCollectSeekJob,
	"etc":     job.NewCollectEtcJob,
	"kava":    job.NewCollectKavaJob,
	"bnb":     job.NewCollectBNBJob,
	"zvc":     job.NewCollectZvcJob,
	"cds":     job.NewCollectCdsJob,
	"hx":      job.NewCollectHxJob,
	"ar":      job.NewCollectARJob,
	"crab":    job.NewCollectCringJob,
	"vet":     job.NewCollectVetJob,
	"mtr":     job.NewCollectMtrJob,
	"qtum":    job.NewCollectQtumJob,
	"dot":     job.NewCollectDotJob,
	"azero":   job.NewCollectAzeroJob,
	"nodle":   job.NewCollectNodleJob,
	"sgb-sgb": job.NewCollectSgbJob,
	"kar":     job.NewCollectKarJob,
	//"btm":         job.NewCollectBtmJob,
	"stx":         job.NewCollectStxNewJob,
	"ont":         job.NewCollectOntJob,
	"ksm":         job.NewCollectKsmJob,
	"bnc":         job.NewCollectBncJob,
	"crust":       job.NewCollectCrustJob,
	"hnt":         job.NewCollectHntJob,
	"celo":        job.NewCollectCeloJob,
	"fio":         job.NewCollectFioJob,
	"sol":         job.NewCollectSolnewJob,
	"ckb":         job.NewCollectCkbJob,
	"nas":         job.NewCollectNasJob,
	"bsc":         job.NewCollectBscJob,
	"fil":         job.NewCollectFilJob,
	"wd":          job.NewCollectWd_wdJob,
	"near":        job.NewCollectNearJob,
	"satcoin":     job.NewCollectSatcoinJob,
	"cfx":         job.NewCollectCfxJob,
	"star":        job.NewCollectStarJob,
	"fis":         job.NewCollectFisJob,
	"oneo":        job.NewCollectNeoJob,
	"atp":         job.NewCollectAtpJob,
	"cph-cph":     job.NewCollectCphJob,
	"trx":         job.NewCollectTrxJob,
	"pcx":         job.NewCollectChainXJob,
	"mw":          job.NewCollectMwJob,
	"algo":        job.NewCollectAlgoJob,
	"ori":         job.NewCollectOriJob,
	"usdt":        job.NewCollectUsdtJob,
	"heco":        job.NewCollectHecoJob,
	"biw":         job.NewCollectBiwJob,
	"hsc":         job.NewCollectHscJob,
	"dhx":         job.NewCollectDhxJob,
	"dom":         job.NewCollectDomJob,
	"wtc":         job.NewCollectWtcJob,
	"moac":        job.NewCollectMoacJob,
	"kai":         job.NewCollectKaiJob,
	"rbtc":        job.NewCollectRbtcJob,
	"sep20":       job.NewCollectSep20Job,
	"rei":         job.NewCollectReiJob,
	"dscc":        job.NewCollectDsccJob,
	"dscc1":       job.NewCollectDscc1Job,
	"brise-brise": job.NewCollectBriseJob,
	"ccn":         job.NewCollectCcnJob,
	"optim":       job.NewCollectOptimJob,
	"ftm":         job.NewCollectFtmJob,
	"welups":      job.NewCollectWelJob,
	"rose":        job.NewCollectRoseJob,
	"rev":         job.NewCollectRevJob,
	"tkm":         job.NewCollectTkmJob,
	"ron":         job.NewCollectRonJob,
	"one":         job.NewCollectOneJob,
	"neo":         job.NewCollectN3neoJob,
	"flow":        job.NewCollectFlowJob,
	"icp":         job.NewCollectIcpJob,
	"uenc":        job.NewCollectUencJob,
	"btm":         job.NewCollectBtmJob,
	"cspr":        job.NewCollectCsprJob,
	"matic-matic": job.NewCollectMaticJob,
	"okt":         job.NewCollectOktJob,
	"waves":       job.NewCollectWavesJob,
	"glmr":        job.NewCollectGlmrJob,
	"avaxcchain":  job.NewCollectAvaxcchainJob,
	"iotx":        job.NewCollectIotexJob,
	"aur":         job.NewCollectAurJob,
	"evmos":       job.NewCollectEvmosJob,
	"mob":         job.NewCollectMobJob,
	"deso":        job.NewCollectDesoJob,
	"lat":         job.NewCollectLatJob,
	"hbar":        job.NewCollectHbarJob,
}

func main() {
	var (
		//	err        error
		routineNum int
		cfgFile    string
		cfg        conf.CollectConfig
	)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()

	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "conf/app.toml", "set the toml config file")
	flag.Parse()

	if routineNum <= 0 || routineNum > 10 {
		routineNum = 4
	}
	runtime.GOMAXPROCS(runtime.NumCPU() * routineNum)
	if err := conf.LoadConfig3(cfgFile, &cfg); err != nil {
		panic(err)
	}
	conf.DecryptCfg(&cfg)
	// 数据库加载
	db.InitOrm2(cfg.DB)
	if cfg.DingName == "" {
		cfg.DingName = "coin-collect"
	}
	job.InitDingErrBot(cfg.DingName, cfg.DingToken)
	// 开启定时任务
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	for k, v := range cfg.Collectors {
		collectfunc, ok := collectorFuncMap[k]
		if !ok {
			panic(fmt.Errorf("don't find any collector %s", k))
		}

		jobrunner.Schedule(v.Spec, collectfunc(*v))
	}

	// starCronJob()// 添加定时归集任务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("get signal: %v. sever will stop and showdown !", sig)

	jobrunner.Stop()
	log.Println("server showdown !")
}

// func JobJson(c *gin.Context) {
//	// returns a map[string]interface{} that can be marshalled as JSON
//	c.JSON(200, jobrunner.StatusJson())
// }
//
// func JobHtml(c *gin.Context) {
//	// Returns the template data pre-parsed
//	c.HTML(200, "Status.html", jobrunner.StatusPage())
//
// }
//
// func initRouter(runmode string) *gin.Engine {
//
//	r := gin.New()
//	r.Use(gin.Recovery())
//	r.Use(middleware.GinCors())
//
//	if runmode == "dev" {
//		r.Use(gin.Logger())
//		gin.SetMode("debug")
//	}
//
//	// Resource to return the JSON data
//	r.GET("/jobrunner/json", JobJson)
//
//	// Load template file location relative to the current working directory
//	r.LoadHTMLGlob("./views/Status.html")
//	//routes.LoadHTMLFiles("./views/Status.html")
//
//	// Returns html page at given endpoint based on the loaded
//	// template from above
//	r.GET("/jobrunner/html", JobHtml)
//
//	return r
// }
