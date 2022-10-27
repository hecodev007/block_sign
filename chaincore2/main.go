package main

import (
	"fmt"
	"github.com/astaxie/beego/plugins/auth"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/group-coldwallet/chaincore2/routers"
	"github.com/group-coldwallet/chaincore2/service/ar"
	"github.com/group-coldwallet/chaincore2/service/bnb"
	"github.com/group-coldwallet/chaincore2/service/chainx"
	"github.com/group-coldwallet/chaincore2/service/ckb"
	"github.com/group-coldwallet/chaincore2/service/cocos"
	"github.com/group-coldwallet/chaincore2/service/dash"
	"github.com/group-coldwallet/chaincore2/service/dcr"
	"github.com/group-coldwallet/chaincore2/service/dhx"
	"github.com/group-coldwallet/chaincore2/service/eos"
	"github.com/group-coldwallet/chaincore2/service/fibos"
	"github.com/group-coldwallet/chaincore2/service/hc"
	"github.com/group-coldwallet/chaincore2/service/kakao"
	"github.com/group-coldwallet/chaincore2/service/ksm"
	"github.com/group-coldwallet/chaincore2/service/mdu"
	"github.com/group-coldwallet/chaincore2/service/neo"
	"github.com/group-coldwallet/chaincore2/service/ont"
	"github.com/group-coldwallet/chaincore2/service/qtum"
	"github.com/group-coldwallet/chaincore2/service/rsk"
	"github.com/group-coldwallet/chaincore2/service/ruby"
	"github.com/group-coldwallet/chaincore2/service/seek"
	"github.com/group-coldwallet/chaincore2/service/stacks"
	"github.com/group-coldwallet/chaincore2/service/ve"
	"github.com/group-coldwallet/chaincore2/service/ycash"
	"github.com/group-coldwallet/chaincore2/service/zvc"
	"github.com/group-coldwallet/common/log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	//gsrpc "github.com/centrifuge/go-substrate-rpc-client"
)

func init() {
	debug_log_path := "./debug.log"
	error_log_path := "./error.log"
	if runtime.GOOS == "linux" {
		_, fileName := filepath.Split(os.Args[0])
		logpath := fmt.Sprintf("/data/logs/%s/", fileName)
		os.MkdirAll(logpath, 0644)
		debug_log_path = fmt.Sprintf("%sdebug.log", logpath)
		error_log_path = fmt.Sprintf("%serror.log", logpath)
	}

	cfg := &log.Logcfg{
		Level:            log.LvlDebug,
		Env:              log.EnvDevelopment,
		TxtType:          log.TxtJson,
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		OutputPaths:      []string{"stdout", debug_log_path},
		ErrorOutputPaths: []string{"stderr", error_log_path},
		//OutputPaths:  nil,
		//ErrorOutputPaths: nil,
	}

	log.InitLog(cfg)
}

func InitDB() {
	maxCPU := runtime.NumCPU()
	syncdsn := beego.AppConfig.String("syncdsn")
	userdsn := beego.AppConfig.String("userdsn")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	err := orm.RegisterDataBase("default", "mysql", syncdsn)
	if err != nil {
		panic("InitDB syncdsn:" + err.Error())
	}
	err = orm.RegisterDataBase("user", "mysql", userdsn)
	if err != nil {
		panic("InitDB userdsn:" + err.Error())
	}
	orm.SetMaxIdleConns("default", maxCPU*2)
	orm.SetMaxOpenConns("default", maxCPU*4)
	ormdebug, _ := beego.AppConfig.Bool("ormdebug")
	orm.Debug = ormdebug

	// 注册model模型
	//orm.RegisterModel(new(models.User))
	//调用 RunCommand 执行 orm 命令。
	//orm.RunCommand()
}

func SecretAuth(username, password string) bool {
	// The username and password parameters comes from the request header,
	// make a database lookup to make sure the username/password pair exist
	// and return true if they do, false if they dont.

	// To keep this example simple, lets just hardcode "hello" and "world" as username,password
	if username == "hello" && password == "world" {
		return true
	}
	return false
}

func main() {
	log.Info("main")
	//api, err := gsrpc.NewSubstrateAPI("wss://kusama-rpc.polkadot.io/")
	//api, err := gsrpc.NewSubstrateAPI("ws://ksm.rylink.io:30944")
	//if err != nil {
	//	log.Debug(err)
	//	return
	//}
	//hash, _ := api.RPC.Chain.GetBlockHashLatest()
	//log.Debug(hash.Hex())
	//head, _ := api.RPC.Chain.GetHeaderLatest()
	//log.Debug(head.Number, head.ParentHash.Hex())
	//bhash, _ := api.RPC.Chain.GetBlockHash(88674)
	//log.Debug(bhash.Hex())
	//
	//metaData, _ := api.RPC.State.GetMetadata(bhash)
	//log.Debug(metaData)
	//log.Debug(api.RPC.State.GetMetadataLatest())
	////sigblk, _ := api.RPC.Chain.GetBlockLatest()
	////log.Debug(sigblk)
	//return
	if len(os.Args) > 1 {
		beego.LoadAppConfig("ini", "conf/"+os.Args[1])
		log.Debug("load config", "conf/"+os.Args[1])
	}
	maxCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(maxCPU)

	/*
		// 读取配置中心
		if beego.AppConfig.DefaultBool("enableetcd", false) {
			addrs := strings.Split(beego.AppConfig.String("etcdhost"), ",")
			client, err := NewEtcdClient(addrs, beego.AppConfig.String("etcduser"), beego.AppConfig.String("etcdpass"))
			if err != nil {
				log.Error(err)
				return
			}
			defer client.client.Close()

			config, err := client.GetConfig(beego.AppConfig.String("etcdkey"))
			var _etcconf map[string]interface{}
			if err := json.Unmarshal([]byte(config), &_etcconf); err != nil {
				for k, v := range _etcconf {
					if k != "" {
						beego.AppConfig.Set(k, v.(string))
					}
				}
			}
		}
	*/

	// mq
	//if !common.InitMQ() {
	//	return
	//}
	if beego.AppConfig.DefaultBool("enableauth", true) {
		beego.InsertFilter("*", beego.BeforeRouter, auth.Basic(beego.AppConfig.DefaultString("username", "username"), beego.AppConfig.DefaultString("password", "password")))
	}

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	// 启用日志
	beego.BConfig.Log.AccessLogs = true

	if beego.AppConfig.DefaultBool("enabledb", true) {
		InitDB()
	}

	// 史上最陋的代码，先这样吧，上线要紧
	coin := strings.ToLower(beego.AppConfig.String("coin"))
	fmt.Println(coin)
	switch coin {
	case "agent":
		routers.AgentInit()
	case "addrmanager":
		routers.AddrManagerInit()
	case "bnb":
		bnb.InitWatchAddress()
		bnb.InitPush()
		bnb.InitContract()
		bnb.InitSync()
		routers.CommonInit()
		routers.BnbInit()
		bnb.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			bnb.StartSync()
		}
	case "fo":
		fibos.InitWatchAddress()
		fibos.InitPush()
		fibos.InitContract()
		routers.CommonInit()
		routers.FibosInit()
		fibos.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			fibos.InitSync()
			fibos.StartSync()
		}
	case "eos":
		eos.InitWatchAddress()
		eos.InitPush()
		eos.InitContract()
		routers.CommonInit()
		routers.EosInit()
		fibos.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			eos.InitSync()
			eos.StartSync()
		}
	case "mdu":
		mdu.InitWatchAddress()
		mdu.InitPush()
		mdu.InitContract()
		routers.CommonInit()
		routers.MduInit()
		mdu.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			mdu.InitSync()
			mdu.StartSync()
		}
	case "ar":
		ar.InitWatchAddress()
		ar.InitPush()
		ar.InitContract()
		routers.CommonInit()
		routers.ArInit()
		ar.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ar.InitSync()
			ar.StartSync()
		}

	case "vet":
		ve.InitWatchAddress()
		ve.InitPush()
		ve.InitContract()
		routers.CommonInit()
		routers.VeInit()
		ve.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ve.InitSync()
			ve.StartSync()
		}

	case "rsk":
		rsk.InitWatchAddress()
		rsk.InitPush()
		rsk.InitContract()
		routers.CommonInit()
		routers.RifInit()
		rsk.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			rsk.InitSync()
			rsk.StartSync()
		}
		//rsk.Test()
		//rsk.StartBlockAmountTimer()
	case "pcx":
		chainx.InitWatchAddress()
		chainx.InitPush()
		chainx.InitSync()
		routers.CommonInit()
		routers.ChainXInit()
		chainx.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			chainx.StartSync()
		}
	case "cocos":
		cocos.InitWatchAddress()
		cocos.InitContract()
		cocos.InitPush()
		cocos.InitSync()
		routers.CommonInit()
		routers.CocosInit()
		cocos.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			cocos.StartSync()
		}
	case "dash":
		dash.InitWatchAddress()
		dash.InitPush()
		dash.InitSync()
		routers.CommonInit()
		routers.DashInit()
		dash.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			dash.StartSync()
		}
	case "dcr":
		dcr.InitWatchAddress()
		dcr.InitPush()
		dcr.InitContract()
		dcr.InitSync()
		routers.CommonInit()
		routers.DcrInit()
		dcr.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			dcr.StartSync()
		}
	case "hc":
		hc.InitWatchAddress()
		hc.InitPush()
		hc.InitContract()
		hc.InitSync()
		routers.CommonInit()
		routers.HcInit()
		hc.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			hc.StartSync()
		}
	case "klay":
		kakao.InitWatchAddress()
		kakao.InitPush()
		kakao.InitContract()
		kakao.InitSync()
		routers.CommonInit()
		routers.KakaoInit()
		kakao.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			kakao.StartSync()
		}
	case "ksm":
		ksm.InitWatchAddress()
		ksm.InitPush()
		ksm.InitContract()
		ksm.InitSync()
		routers.CommonInit()
		routers.KsmInit()
		ksm.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ksm.StartSync()
		}
	case "dhx":
		dhx.InitWatchAddress()
		dhx.InitPush()
		dhx.InitContract()
		dhx.InitSync()
		routers.CommonInit()
		routers.DhxInit()
		dhx.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			log.Info("dhx.StartSync()")
			dhx.StartSync()
		}
	case "neo":
		neo.InitWatchAddress()
		neo.InitPush()
		neo.InitContract()
		neo.InitSync()
		routers.CommonInit()
		routers.NeoInit()
		neo.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			neo.StartSync()
		}
		//neo.StartBlockAmountTimer()
	case "ont":
		ont.InitWatchAddress()
		ont.InitPush()
		ont.InitContract()
		ont.InitSync()
		routers.CommonInit()
		routers.OntInit()
		ont.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ont.StartSync()
		}
	case "qtum":
		qtum.InitWatchAddress()
		qtum.InitPush()
		qtum.InitContract()
		qtum.InitSync()
		routers.CommonInit()
		routers.QtumInit()
		qtum.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			qtum.StartSync()
		}
	case "rub":
		ruby.InitWatchAddress()
		ruby.InitPush()
		ruby.InitContract()
		ruby.InitSync()
		routers.CommonInit()
		routers.RubyInit()
		ruby.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ruby.StartSync()
		}
	case "seek":
		seek.InitWatchAddress()
		seek.InitPush()
		seek.InitContract()
		seek.InitSync()
		routers.CommonInit()
		routers.SeekInit()
		seek.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			seek.StartSync()
		}
	case "stx": // stacks
		stacks.InitWatchAddress()
		stacks.InitPush()
		stacks.InitContract()
		stacks.InitSync()
		routers.CommonInit()
		routers.StacksInit()
		stacks.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			stacks.StartSync()
		}
	case "yec":
		ycash.InitWatchAddress()
		ycash.InitPush()
		ycash.InitSync()
		routers.CommonInit()
		routers.YcashInit()
		ycash.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ycash.StartSync()
		}
	case "zec":
		ycash.InitWatchAddress()
		ycash.InitPush()
		ycash.InitSync()
		routers.CommonInit()
		routers.YcashInit()
		ycash.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ycash.StartSync()
		}
	case "zvc":
		zvc.InitWatchAddress()
		zvc.InitPush()
		zvc.InitContract()
		zvc.InitSync()
		routers.CommonInit()
		routers.ZvcInit()
		zvc.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			zvc.StartSync()
		}
	case "ckb", "ckt":
		ckb.InitWatchAddress()
		ckb.InitPush()
		ckb.InitContract()
		ckb.InitSync()
		routers.CommonInit()
		routers.CkbInit()
		ckb.RunPush()
		if beego.AppConfig.DefaultBool("enablesync", true) {
			// 初始化同步服务
			ckb.StartSync()
		}
	default:
		break
	}

	beego.Run()
}

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o=xxx
