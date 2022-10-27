package cocosImpl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/cocos/common"
	"github.com/group-coldwallet/cocos/model/bo"
	"github.com/group-coldwallet/cocos/service/cocosImpl/rpc"
	"github.com/spf13/viper"
	"net/http"
)

func CreateAccount(c *gin.Context) {
	var req bo.CreateAccountReq
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    2001,
			"message": "传入的参数错误",
		})
		return
	}

	//申请公私钥
	accountname := req.Account
	logs.Debug("accountname:", accountname)
	owner := suggestBrainKey()
	if owner == nil {
		createAccountHTTPError(c, "请求失败")
		return
	}
	logs.Debug("owner:")
	logs.Debug(owner)
	active := owner
	if active == nil {
		createAccountHTTPError(c, "请求失败")
		return
	}
	logs.Debug("active")
	logs.Debug(active)
	//注册账户
	ret := registerAccount(accountname, owner.PubKey, active.PubKey)
	if ret != 0 {
		createAccountHTTPError(c, "注册账户失败")
		logs.Debug("注册账户失败")
		return
	}
	logs.Debug("注册账户成功")
	//导入账户
	importKey(accountname, active.WifPrivKey)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "注册账户成功",
		"data": accountname,
	})
}

func createAccountHTTPError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "600",
		"message": message,
		"data":    "",
	})
}

func createAccountJsonError(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "600",
		"message": "json解析出错",
		"data":    "",
	})
}

func suggestBrainKey() *bo.BrainKey {
	rpcContent := rpc.SuggestBrainKey()
	ret := common.HTTPRPC(rpcContent)
	if ret.Code != 0 {
		return nil
	}
	sbres := bo.RpcResponse{}
	err := json.Unmarshal([]byte(ret.Body), &sbres)
	if err != nil {
		logs.Debug(err.Error())
		return nil
	}
	bk := new(bo.BrainKey)
	bkmapinterface := sbres.Result.(map[string]interface{})
	bk.BrainPrivKey = fmt.Sprintf("%v", bkmapinterface["brain_priv_key"])
	bk.WifPrivKey = fmt.Sprintf("%v", bkmapinterface["wif_priv_key"])
	bk.PubKey = fmt.Sprintf("%v", bkmapinterface["pub_key"])
	return bk
}

func registerAccount(accountname string, owner string, active string) int {
	rpcContent := fmt.Sprintf(`{"jsonrpc":"2.0", "method":"register_account", "params": ["%s","%s","%s","%s",true], "id":"2"}`, accountname, owner, active, viper.GetString("account.viper"))
	logs.Debug(rpcContent)
	ret := common.HTTPRPC(rpcContent)
	logs.Debug(ret.Body)
	if ret.Code != 0 {
		logs.Debug("rpc请求出错")
		return -1
	}
	r := bo.RegisterResponse{}
	err := json.Unmarshal([]byte(ret.Body), &r)
	if err != nil {
		logs.Debug("解析json出错")
		return -1
	}
	logs.Debug(r.Error)
	if r.Error != nil {
		logs.Debug(r)
		return -1
	}
	return 0
}

func importKey(account string, key string) {
	rpcContent := fmt.Sprintf(`{"jsonrpc":"2.0", "method":"import_key", "params": ["%s","%s"], "id":"2"}`, account, key)
	ret := common.HTTPRPC(rpcContent)
	logs.Debug(ret.Body)
}

func GetAccount(account string) ([]byte, error) {
	rpcStr := fmt.Sprintf(`{"jsonrpc": "2.0", "id":"2", "method": "get_account", "params": ["%s"] }`, account)
	logs.Debug(rpcStr)
	ret := common.HTTPRPC(rpcStr)
	var br bo.BalanceReturn
	br.Code = ret.Code
	if ret.Code != 0 {
		return nil, errors.New("请求失败")
	}
	logs.Debug(ret.Body)
	return []byte(ret.Body), nil
}
