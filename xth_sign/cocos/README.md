### cocos接口
<pre>
/v1/cocosbcx/transfer     POST
{
  "applyid": 117,
  "outerorderno": "btc_yxc_006",
  "orderno": "dhptest3",
  "mchname": "hoo",
  "coinname": "cocos",
  "toaddress": "dhptest3",
  "toamount": "1.12345",
  "memo": ""
}
返回值
{
    "code": 0,  #0 表示正确返回，其它值表示错误
    "data": ,  # txid
    "message": ""  //备注信息
}
/v1/cocosbcx/getblance    GET   获取系统默认账户的余额
{
    "code": 0,
    "data": 99659.03231,
    "message": "1.3.0"
}

### 5. 程序目录结构

<pre>
/launcher            # 各类需要初始化的启动器
/internal            # 应用程序的封装的代码，比如某个结构体json解析，http返回错误代码，某个应用私有的代码放到 /internal/myapp/ 目录下，多个应用通用的公共的代码，放到 /internal/common 之类的目录
/api                 # 路由接口实现
/pkg                 # 一些通用的可以被其他项目所使用的代码，放到这个目录下面
/config              # 配置文件 
/middleware          # 中间件
/model               # Model
    /constants       # 常量定义
    /bo              # 业务对象 封装业务逻辑对象，结合po,vo进行业务操作  
    /po              # 持久对象（数据库映射，数据库交互）
    /vo              # 值对象，通常用于response层响应数据   
/router              # Router(Controller)
/script              # 类脚本代码，通常是完成一次性的任务
/service             # 业务实现
/doc                 # API文档
/docs                # swagger文档生成目录
/assets              # 静态资源（js，css，image）
/test                # 功能性测试
/tmp                 # 临时缓存目录（log）
/util                # 工具类库
/main.go             # 入口文件