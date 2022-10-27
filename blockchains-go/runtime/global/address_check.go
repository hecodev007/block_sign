package global

var CheckAddressServer map[string]string

//初始化
//todo 准备废弃
func InitCheckAddressServer() {
	servers := make(map[string]string, 0)
	//servers["btc"] = "http://47.244.140.180:9999/api/v1/btc/validateaddress?address=%s"
	//servers["usdt"] = "http://47.244.140.180:9999/api/v1/usdt/validateaddress?address=%s"
	//servers["zec"] = "http://47.244.140.180:9999/api/v1/zec/validateaddress?address=%s"
	servers["ltc"] = "http://47.244.140.180:9999/api/v1/ltc/validateaddress?address=%s"
	//servers["cocos"] = "http://52.195.18.33:8091/v1/cocosbcx/vaildaddress?account=%s"
	//servers["mdu"] = "http://52.195.18.33:8091/v1/cocosbcx/vaildaddress?account=%s"
	servers["zvc"] = "http://13.231.250.0:10015/api/v1/validaddr"
	//servers["zvc"] = "http://192.169.2.157:10015/api/v1/validaddr"
	servers["klay"] = "http://localhost:18854/klay/validaddress?address=%s"
	CheckAddressServer = servers
}
