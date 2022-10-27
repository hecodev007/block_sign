**目录介绍**
|----middleware
    |----binder             绑定器
    |----cache              缓存
    |----captcha            验证码
    |----metrics            请求监控
    |----multitemplate      多样模板
    |----opentracing        开放追踪
    |----pongo2echo         模板引擎
    |----pprof              匹配授权     
    |----rabbitmq           MQ 限流、削峰，异步通知
    |----session            session缓存，redis，cookie
    |----staticbin          静态箱 静态文件路径


**Group Level**

创建新组时，您可以仅为该组注册中间件。
例如，您可以拥有一个通过为其注册 BasicAuth 中间件而受到保护的管理员组。
```go
e := echo.New()
admin := e.Group("/admin", middleware.BasicAuth())
```
您还可以在通过 admin.Use() 创建组后添加中间件。

**Route Level**

定义新路由时，您可以选择为其注册中间件。
```go

e := echo.New()
e.GET("/", <Handler>, <Middleware...>)

```

Skipping Middleware
在某些情况下，您希望根据某些条件跳过中间件，因为每个中间件都有一个选项来定义函数 Skipper func(c echo.Context) bool。

```go

e := echo.New()
e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	Skipper: func(c echo.Context) bool {
		if strings.HasPrefix(c.Request().Host, "localhost") {
			return true
		}
		return false
	},
}))

```
当请求主机以 localhost 开始时，上面的示例跳过 Logger 中间件。
