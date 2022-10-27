package main

import (
	"github.com/astaxie/beego/plugins/auth"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/group-coldwallet/chaincore2/controllers/agent"
	"github.com/group-coldwallet/chaincore2/routers"
	"runtime"

	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
)

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

	routers.CommonInit()
	routers.AgentInit()

	//php测试使用
	agent.TestUrl = beego.AppConfig.DefaultString("testurl", "")

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
	beego.Run()
}
