文件目录结构：
    
    |---custody-merchant
        |---config      配置文件
        |---global      全局配置文件
        |---internal    web项目服务
        |---middleware  中间件
        |---model       项目模型，实体
        |---module      项目模块
        |---proto       grpc模块
        |---router      路由管理
        |---util        工具
        |---go.mod      包版本管理
        |---main.go     项目启动

框架使用：

- go: 1.16
- echo:v4
- redis
- rabbitmq
- mysql
- opentracing
- protoc/grpc

配置文件介绍：

    |---config
        |---config.go     配置文件读取转为实体

配置文件toml格式

```text
[配置名]
param1 = ..
param2 = ..
```

在config.go文件下添加：

```go
type config struct {
	...
	// 应用配置 
	// App：配置名 app：toml里你设置的配置名称
	App app
}
// app：toml里你设置的配置名称
type app struct {
    // toml：toml里你设置的参数名	
    Param1  string `toml:"Param1"`
    Param1  string `toml:"Param2"`
}
```
                    domain    
前端 <--> | controller <--> service | <--> model <--> db 

打包程序
```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
```


配置文件
```json
钱包接口路由配置
[blockchaincustody]
base_url = ""
coin_list = "127.0.0.1:10086/v3/isInsideAddress"
create_mch = ""
reset_mch = ""
get_mch = ""
verify_param = ""
create_address = ""
withdraw = ""
balance = ""
chain_status = ""
```

#### mysql同步表

admin_package_pay

admin_package_trade

admin_package

admin_record

service_finance