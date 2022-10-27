package common

import (
	"github.com/spf13/viper"
	"github.com/group-coldwallet/cocos/model/bo"
)

func HTTPRPC(rpcContent string) bo.HttpReturn {
	rpcurl := viper.GetString("rpc.url")
	ret := HttpRequest("post", rpcurl, []byte(rpcContent))
	return *ret
}
