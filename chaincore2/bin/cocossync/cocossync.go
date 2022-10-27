package main

import (
	"fmt"
	"github.com/astaxie/beego/plugins/auth"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/group-coldwallet/chaincore2/routers"
	coinservice "github.com/group-coldwallet/chaincore2/service/cocos"
	"github.com/group-coldwallet/common/log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
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
	//maxCPU := runtime.NumCPU()
	//syncdsn := beego.AppConfig.String("syncdsn")
	//userdsn := beego.AppConfig.String("userdsn")
	//orm.RegisterDriver("mysql", orm.DRMySQL)
	//orm.RegisterDataBase("default", "mysql", syncdsn)
	//orm.RegisterDataBase("user", "mysql", userdsn)
	//orm.SetMaxIdleConns("default", maxCPU*2)
	//orm.SetMaxOpenConns("default", maxCPU*4)
	//ormdebug, _ := beego.AppConfig.Bool("ormdebug")
	//orm.Debug = ormdebug

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
	//beego.LoadAppConfig("ini", "conf/app2.conf")
	maxCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(maxCPU)
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
	if beego.AppConfig.String("runmode") == "dev" {
		beego.BConfig.Log.AccessLogs = true
	}

	InitDB()

	// 初始化
	if !coinservice.InitWatchAddress() {
		return
	}
	coinservice.InitPush()
	coinservice.InitContract()
	routers.CommonInit()
	routers.CocosInit()

	// 启用日志
	beego.BConfig.Log.AccessLogs = true
	coinservice.RunPush()

	if beego.AppConfig.DefaultBool("enablesync", true) {
		// 初始化同步服务
		coinservice.InitSync()
		coinservice.StartSync()
	}

	beego.Run()
}
