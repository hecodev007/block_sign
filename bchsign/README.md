# BCH源码签名

###GIT仓库
http://example.com/v1/postal_services


##API文档列表
* [文档](doc/trx.md)


### 命名规范

* 所有返回的JSON数据的key，按驼峰书写，首字母小写(_id应补写为id)
* URI中的资源命名，字词之间以下划线分隔，如 http://example.com/v1/postal_services

## GO项目规范


### 1. **尽量满足 `golint`  和 `go tool vet`** 目前暂不做强制要求


### 2. 命名规则

#### 文件名

* 入口： main.go
* 驼峰，首字母小写

#### 函数或方法

* 若函数或方法为判断类型（返回值主要为 bool 类型），则名称应以 Has, Is, Can 或 Allow 等判断性动词开头
* 函数命名通常是： 动词+名词， 如findUsers()

#### 常量

* 常量均需使用首字母大写的驼峰命名法，如 ServiceType

#### 变量

* 变量命名基本上遵循相应的英文表达或简写
* 遇到特有名词时，需要遵循以下规则：
  * 如果变量为私有，且特有名词为首个单词，则使用小写，如 apiClient
  * 其它情况都应当使用该名词原有的写法，如 APIClient

#### 测试

* TestXXX

### 3. Import 规范

1.在非测试文件（*_test.go）中，禁止使用 . 来简化导入包的对象调用

2.禁止使用相对路径导入（./subpackage），所有导入路径必须符合 go get 标准

### 4. 注释规范

* 包头前注释说明此包的作用
* // 注释

### 5. 程序目录结构

<pre>
/api                 # 路由接口实现
    /helper          # API JSON数据的helper函数
/conf                # 配置文件 
/i18n
/middleware          # 中间件
/model               # Model
    /globalparam          # 全局变量
    /constparam           # 常量
    /bo              # 业务对象  
    /po              # 持久对象（数据库映射，数据库交互）
    /vo              # 值对象  （response）   
/pem                 # 密钥文件
/router              # Router(Controller)
/script              # 类脚本代码，通常是完成一次性的任务
/service             # 业务实现
/doc                 # API文档
/static              # 静态文件（js，css，image）
/test                # 功能性测试
/tmp                 # 临时缓存目录（类log）
/util                # 工具类库
/main.go             # 入口文件

</pre>

### 6. 代码指导

* model中一些表示状态的字段，都必须要用的常量来表示

  如User.gender uint8 ， 那么得定义 MALE=1, FEMALE=2

* model常用数据库查询命名参考

  > InsertResource() 增加一条数据
  
  > BatchResources() 批量增加

  > FindResources() 查找多个
  
  > FindResource()  查找一个
  
  > FindResourceBy*() 按某字段查找一个
  
  > UpdateResource()  更新字段
  
  > UpdateSetResource()  只做更新操作 相当于update(selector, {"$set": update})
  
  > UpsertResource()  upsert操作
  
  > DeleteResource()  删除

* 当resources时是单复数同行时

  > InsertAllResource() 增加多条数据
  
  > FindAllResource() 查找多个
  
  > UpdateAllResource() 更新多条数据某字段
  
  > UpdateSetAllResource() 更新多条数据某字段

* router(controller)命名参考

   > [GET] /resources            >  ListResources()
  
   > [POST] /resources           >  CreateResource()
  
   > [GET] /resouces/:id         >  ShowResource()
  
   > [PUT/PATCH] /resources/:id  >  UpdateResource()
  
   > [DELETE] /resources/:id     >  DeleteResource()
  
   > [GET] /resouces/:id/edit    >  EditResource()
  
   > [GET] /resources/:id/new    >  NewResource()

### 7. 配置文件

   默认配置文件在 conf/application.yml文件下
   配置信息的优先级是 application.yml < 系统环境变量 < 命令行参数

## GIT 规范

### Commit message格式

```
commit -m "<type>: (<scope>) <subject>
// 空一行
<body>
// 空一行
<footer>"
```

其中`<type>`和`<subject>`必须

`<type>` 用于说明 commit 的类别，只允许使用下面7个标识

* feat：新功能（feature）
* fix：修补bug
* docs：文档（documentation）
* style： 格式（不影响代码运行的变动）
* refactor：重构（即不是新增功能，也不是修改bug的代码变动）
* test：增加测试
* chore：构建过程或辅助工具的变动

详情见: http://www.ruanyifeng.com/blog/2016/01/commit_message_change_log.html

### 分支规范

master -> hostfix -> release -> develop -> feature
详情见: http://nvie.com/posts/a-successful-git-branching-model/

### 其他

在develop分支提交代码前，用git pull --rebase命令拉取合并最新代码,避免产生额外类似这些无意义的merge commit:Merge branch 'develop' of 127.0.0.1:hoo/openapi into develop
