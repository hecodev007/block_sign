## 更新迭代日志

### 说明
暂时不改动莫源所编辑文件，新起server切换
交易service为transfer_service.go
各个币种交易实现在service---》transfer目录

### 新币种增加策略
1. application.toml 配置[transfer] [wallettype]，如果是热钱包还需要配置[hotservers]
2. 实现各个币种交易实现在service/transfer
3. 在runtime/job/init_service.go初始化币种
4. 在service/security/security_service.go初始化币种

5. 在runtime/global/address_check.go中添加验证地址服务器，如果无需调用这些服务器，可自行在各种验证方法处理（后续转移到配置文件）


### 2020年3月3号

该版本尚未实际运行，注释掉了main文件中的一些加载项，可以先先各自币种交易在service/transfer目录

币种交易模型需要首先在配置文件填写，会根据交易模型进行保存对应的结构，内部订单ID规则尚未研究，下次更新补上

新增下单API，前置认证，穿插大量验证 === 》router/api/v1/transfer.go

添加BTC交易支持 ===》 service/transfer/btc.go

轮询任务 交易任务实现，支持可视化，支持删除指定ID===》 pkg/job/transfer.go


### 2020年3月3号 （2）
新增下单保存的orderNo，使用uuid

新添加的币种交易需要在pkg/job/init_service.go进行初始化

调整路径权限验证

新增LTC交易 ===》 service/transfer/ltc.go

### 2020年3月4号

下单测试调整逻辑完成

定时轮询测试 抓取订单发送walletserver完成

新增 重复订单判断热钱包币种

新增 正在执行订单判断热钱包币种

新增 刷新全局配置API

新增 热钱包map配置

新增 重试异常状态为8的订单

sign签名尚未测试，其他验证已经通过

### 2020年3月6日
修改sign签名，完成通用匹配验证，测试通过

修复热钱包查询出账地址

修改重复订单验证