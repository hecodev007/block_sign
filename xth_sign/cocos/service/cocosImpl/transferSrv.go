package cocosImpl

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/group-coldwallet/cocos/common"
	"github.com/group-coldwallet/cocos/model/bo"
)

/*
返回值 针对http错误
*/

func TransferSrv(from string, to string, num string, memo string) bo.TransferReturn {
	transferRPC := fmt.Sprintf(`{"jsonrpc":"2.0", "method":"transfer", "params": ["%s","%s","%s","COCOS",["%s",false],true], "id":"2"}`, from, to, num, memo)
	logs.Debug(transferRPC)
	ret := common.HTTPRPC(transferRPC)
	var tr bo.TransferReturn
	tr.Code = ret.Code
	if ret.Code != 0 {
		return tr
	}
	//下面是解析http返回的json数据 由于
	logs.Debug(ret.Body)
	err := json.Unmarshal([]byte(ret.Body), &tr.TRR)
	if err != nil {
		logs.Debug(err.Error())
		ret.Code = -2
		return tr
	}
	logs.Debug(ret.Body)
	logs.Debug("=======================")
	return tr
}
