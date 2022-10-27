# ucaserver

1. 启动读取ab加密文件 获取私钥
2. 启动http服务


#### 文件名

* 入口： main.go


#### swgger重新生成
* swag init


#### 程序目录结构

<pre>
/api                # controller
    /v1             # v1版本
/conf               # 配置文件
/doc                # swagger文档
/middleware         # 中间件
/model              # Model
    /bo             # 业务对象  
    /global         # 全局对象
    /vo             # 值对象  （response）
/pem                # 密钥文件    
/router             # Router
/pkg                # 工具类库以及业务辅助类
/service            # 业务型功能服务，主要处理业务逻辑实现
/script             # 类脚本代码，通常是完成一次性的任务
/tmp                # 临时缓存目录（类log）
/main.go

</pre>

