package router

import (
	"context"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/middleware/opentracing"
	"custody-merchant-admin/middleware/pprof"
	"custody-merchant-admin/module/log"
	sy "custody-merchant-admin/module/sync_sys"
	"custody-merchant-admin/mqService"
	"custody-merchant-admin/router/web"
	"custody-merchant-admin/runtime/job"
	"github.com/bamzi/jobrunner"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmechov4"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type (
	// Host 结构体
	Host struct {
		Echo *echo.Echo
	}
)

var Echos *echo.Echo

// InitRoutes 初始化路由
// 路由集合
// 创建集合：host->router
func InitRoutes() map[string]*Host {
	// Hosts
	hosts := make(map[string]*Host)
	// Key-value
	hosts[Conf.Server.DomainWeb] = &Host{web.Routers()}
	hosts[Conf.Server.DomainApi] = &Host{web.Routers()}
	return hosts
}

// RunSubdomains 子域名部署
func RunSubdomains(confFilePath string) {

	// 配置初始化
	if err := InitConfig(confFilePath); err != nil {
		log.Panic(err)
	}
	// 全局日志级别
	log.SetLevel(GetLogLvl())
	// Server
	Echos = echo.New()
	// pprof
	Echos.Pre(pprof.Serve())
	Echos.Pre(mw.RemoveTrailingSlash())
	// 请求追踪
	// Elastic APM
	// Requires APM Server 6.5.0 or newer
	apm.DefaultTracer.Service.Name = Conf.Opentracing.ServiceName
	apm.DefaultTracer.Service.Version = Conf.App.Version
	Echos.Use(apmechov4.Middleware(
		apmechov4.WithRequestIgnorer(func(request *http.Request) bool {
			return false
		}),
	))
	// OpenTracing
	otCtf := opentracing.Configuration{
		Disabled: Conf.Opentracing.Disable,
		Type:     opentracing.TracerType(Conf.Opentracing.Type),
	}
	if closer := otCtf.InitGlobalTracer(
		opentracing.ServiceName(Conf.Opentracing.ServiceName),
		opentracing.Address(Conf.Opentracing.Address),
	); closer != nil {
		defer closer.Close()
	}
	// 日志级别
	Echos.Logger.SetLevel(GetLogLvl())
	// Secure, XSS/CSS HSTS
	Echos.Use(mw.SecureWithConfig(mw.DefaultSecureConfig))
	Echos.Use(mw.MethodOverride())
	// CORS
	//Echos.Use(mw.CORSWithConfig(mw.CORSConfig{
	//	AllowOrigins: []string{"https://" + Conf.Server.DomainWeb, "http://" + Conf.Server.DomainApi},
	//	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding, echo.HeaderAuthorization},
	//}))
	Echos.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAcceptEncoding,
			echo.HeaderAuthorization, global.XCaNonce, global.XCaTime, global.XCaSignStr},
	}))

	Echos.Use()
	hosts := InitRoutes()
	// 开启异步处理
	sy.InitSyncData()
	Echos.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		// 我们将解析这个 URL 示例，它包含了一个 scheme，认证信息，主机名，端口，路径，查询参数和片段。
		u, _err := url.Parse(c.Scheme() + "://" + req.Host)
		if _err != nil {
			Echos.Logger.Errorf("Request URL parse error:%v", _err)
		}
		host := hosts[u.Hostname()]
		if host == nil {
			Echos.Logger.Info("Host not found")
			err = echo.ErrNotFound
		} else {
			host.Echo.ServeHTTP(res, req)
		}
		return
	})

	// jobRunner()
	// mqConsume()

	if !Conf.Server.Graceful {
		Echos.Logger.Fatal(Echos.Start(Conf.Server.Addr))
	} else {
		// 优雅关闭
		// 开启程序
		go func() {
			if err := Echos.Start(Conf.Server.Addr); err != nil {
				// 报错关闭
				Echos.Logger.Errorf("Shutting down the server with error:%v", err)
			}
		}()
		// 优雅地中断程序
		// 设置超时10秒
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := Echos.Shutdown(ctx); err != nil {
			Echos.Logger.Fatal(err)
		}
	}
}

//定时任务开启
func jobRunner() {
	jobrunner.Start() // optional: jobrunner.Start(pool int, concurrent int) (10, 1)
	//定时更新币列表
	jobrunner.Schedule("@every 1h", job.WalletCoinListCallBackJob{})

}

// 开启MQ消费
func mqConsume() {
	go func() {
		mqService.RunConsume()
	}()
}
