package dingbot

import "fmt"

//发送钉钉告警文本信息
//func DingWarnInfoTpl(coinName, orderId, statusDesc string) []byte {
//	return []byte(fmt.Sprintf("币种名称:%s,订单号:%s,异常状态:%s", coinName, orderId, statusDesc))
//}

//发送钉钉告警文本信息
func DingWarnInfoTpl(mchName, coinName, outOrderNo, statusDesc string) []byte {
	return []byte(fmt.Sprintf("商家名称:%s,币种名称:%s,外部订单号:%s,异常状态:%s", mchName, coinName, outOrderNo, statusDesc))
}

//发送钉钉告警文本信息
func DingFinanceInfoTpl(errStr string) []byte {
	return []byte(fmt.Sprintf("通知Finance业务异常,ERROR:%s", errStr))
}
