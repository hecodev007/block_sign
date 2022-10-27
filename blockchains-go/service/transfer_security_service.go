package service

import "github.com/shopspring/decimal"

type TransferSecurityService interface {

	//验证币种是否支持
	VerifyCoin(coinName string, mchId int) (bool, error)

	//验证币种关闭还是开放
	VerifyCoinPermission(coinName string, mchId int) (bool, error)

	//验证币种精度
	VerifyCoinDecimal(coinName string, amount decimal.Decimal) (bool, error)

	////ip黑名单
	//IsBlacklistIP(ip string, mchId int) (bool, error)
	//
	////ip白名单
	//IsWhitelistIP(ip string, mchId int) (bool, error)

	//验证api权限
	VerifyApiPermission(path, coinName, ip string, mchId int) (bool, error)

	//风险验证
	//单笔出账限额，每小时出账限额，每日出账限额
	VerifyRisk(coinName string, amount decimal.Decimal, mchId int) (bool, error)

	//验证商户余额
	//param ok：true 满足出账  false 不满足出账
	//param mchBalance： 商户数据库余额
	//param coinName：主链币
	//param err：错误描述
	VerifyMchBalance(coinName, contractAddress string, transferAmount decimal.Decimal, mchId int) (ok bool, mchBalance decimal.Decimal, err error)

	//验证是否是重复订单,apply表
	IsDuplicateApplyOrder(outOrderNo string, mchName string) (bool, error)

	//验证是否已经分配地址
	IsAssignAddress(coinName string, mchId int) (bool, error)

	//合法地址验证
	VerifyAddress(address, coinName string) (bool, error)

	//地址是内部还是外部地址
	IsInsideAddress(address string) (bool, error)

	//验证订单是否正在order表执行，或者执行完成
	IsRunningOrder(outOrderNo, coinName string, mchId int) (bool, error)
}
