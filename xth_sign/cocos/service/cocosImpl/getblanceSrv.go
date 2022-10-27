package cocosImpl

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/group-coldwallet/cocos/common"
	"github.com/group-coldwallet/cocos/model/bo"
)

func GetBlanceServ(account string) bo.BalanceReturn {
	rpcStr := fmt.Sprintf(`{"jsonrpc":"2.0", "method":"list_account_balances", "params": ["%s"], "id":"2"}`, account)
	logs.Debug(rpcStr)
	ret := common.HTTPRPC(rpcStr)
	var br bo.BalanceReturn
	br.Code = ret.Code
	if ret.Code != 0 {
		return br
	}
	logs.Debug(ret.Body)
	err := json.Unmarshal([]byte(ret.Body), &br.BRR)
	logs.Debug(br)
	if err != nil {
		logs.Debug(err.Error())
		br.Code = -1
		return br
	}
	return br
}
