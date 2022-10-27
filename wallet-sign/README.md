# Wallet sign service

## 编写service
    1. 在services/v1下创建脚本。例如：btc.go
    2. 在脚本下创建结构体.随意取名，例如： BtcService
    3. 创建初始化结构体函数。*** 注意： 使用BaseService去创建，并且方法名字为：币种名字大写+Service
        例如： func (bs *BaseService)BTCService()*BtcService{return &BtcService{}}
    4. 继承IService接口

## 编写api 
    1. 在routers/apis/v1下创建脚本。例如：btc.go
    2. 创建对应Api接口体，并继承BaseApi
    3. 继承接口Apis

## 初始化api
    在routers/apis/api.go switch下添加对应的分支
    
## 编写配置文件
    添加对应币种的配置文件，可参照gxc的去做

## 启动程序
    修改./conf/config.toml下的cointype为对应的币种
    配置./conf/config.toml下的walletType为hot or cold
    