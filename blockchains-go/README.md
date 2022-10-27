## HooCustody go版本

### 初始化项目 

当我们使用go build，go test以及go list时，go会自动得更新go.mod文件，将依赖关系写入其中。
如果你想手动处理依赖关系，那么使用如下的命令：
<pre>
 go mod tidy
</pre>

### 第三方私有库
需要提前先导入到本地，详细可以查看 `go.mod` 的配置

* 内部common项目，目前使用了log
> go get github.com/group-coldwallet/blockchains-go




### 项目结构
<pre>
blockchains-go/
├── conf
├── middleware
├── model
├── entity
├── dao
├── pkg
├── routers
└── runtime
└── service
</pre>

* conf：用于存储配置文件
* middleware：应用中间件
* model：应用业务模型
* entity：数据库模型
* dao：数据库层
* pkg：第三方包(util，http错误代码定义之类的)
* routers 路由
* runtime 应用运行时数据
* service 业务逻辑处理

### 生成数据库entity
 https://gitea.com/xorm/cmd
 
 转换为驼峰结构，注意xorm的名称映射功能设置，或者对应更改比较特殊的tag

 xorm 名称映射规则 https://www.kancloud.cn/xormplus/xorm/167084

* 安装工具
```
go get xorm.io/cmd/xorm
```
* 编译
```
cd gitea.com/xorm/cmd/xorm && go build
```
* 生成命令（读取entity下面的模板，生成实体类）
```
xorm reverse mysql "root:123456@tcp(127.0.0.1:3306)/data?charset=utf8" ./entity ./entity
```
### 测试钉钉
