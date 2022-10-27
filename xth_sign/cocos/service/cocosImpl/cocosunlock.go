package cocosImpl

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/spf13/viper"
	"github.com/group-coldwallet/cocos/common"
	"github.com/group-coldwallet/cocos/model/bo"
	"github.com/group-coldwallet/cocos/service/cocosImpl/rpc"
)

func Unlock() {
	unlockpassword := viper.GetString("rpc.unlockpassword")
	rpcCon := rpc.Unlock(unlockpassword)

	ret := common.HTTPRPC(rpcCon)
	if ret.Code != 0 {
		panic(fmt.Sprintf("没有成功解锁钱包，连接rpc服务器失败:%d", ret.Code))
		return
	}
	var res bo.RpcResponse
	json.Unmarshal([]byte(ret.Body), &res)
	logs.Debug(res.Result, ":", res.Id, ":", res.JsonRpc, ":", res.Error)
	if res.Error != nil {
		panic("解锁钱包失败，请检查解锁密码是否正确")
	}
}
