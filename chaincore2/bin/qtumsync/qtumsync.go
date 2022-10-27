package main

import (
	"github.com/astaxie/beego/plugins/auth"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/group-coldwallet/chaincore2/routers"
	service "github.com/group-coldwallet/chaincore2/service/qtum"
	"runtime"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func InitDB() {
	maxCPU := runtime.NumCPU()
	syncdsn := beego.AppConfig.String("syncdsn")
	userdsn := beego.AppConfig.String("userdsn")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", syncdsn)
	orm.RegisterDataBase("user", "mysql", userdsn)
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

	InitDB()

	// 初始化
	service.InitWatchAddress()
	service.InitPush()
	service.InitContract()
	service.InitSync()
	routers.CommonInit()
	routers.QtumInit()

	// 启用日志
	beego.BConfig.Log.AccessLogs = true
	service.RunPush()

	if beego.AppConfig.DefaultBool("enablesync", true) {
		// 初始化同步服务
		service.StartSync()
	}

	beego.Run()
}
