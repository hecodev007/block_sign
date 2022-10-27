package runtime

import (
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
)

func InitGlobalReload() {
	InitGlobal()
	// 加载全局配置，通过redis通知prepare服务
	global.NotifyPrepare()
}

//初始化各项运行时启动的全局参数读取
func InitGlobal() {
	//检查地址使用的全局参数
	global.InitCheckAddressServer()

	//币种精度验证使用的全局参数
	global.InitCoinDecimal()

	//商户权限验证使用的全局参数
	global.InitMchApiAuth()

	//交易使用的全局参数
	global.InitTransferModel()

	//冷热钱包类型
	global.InitWalletType()

	//加载商户基本信息
	global.InitMchBaseInfo()

	global.InitMchService()

}

func InitIM() {
	dingding.InitDingBot()
}

func InitDingRole(model string) {
	dingding.InitDingRols(model)
}
