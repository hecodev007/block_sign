package cocos

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
)

// 解锁钱包
func UnlockWallet(pwd string) bool {
	respdata, err := common.RequestUrl("unlock", []interface{}{pwd}, beego.AppConfig.String("walleturl"))
	if err != nil {
		log.Error(err)
		return false
	}

	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {
		log.Debug(err)
		return false
	}

	if datas["result"] == nil {
		return false
	}
	return datas["result"].(bool)
}

// 钱包是否加锁
func IsLocked() (bool, error) {
	respdata, err := common.RequestUrl("is_locked", []interface{}{}, beego.AppConfig.String("walleturl"))
	if err != nil {
		log.Error(err)
		return false, err
	}

	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {
		log.Debug(err)
		return false, err
	}

	if datas["result"] == nil {
		return false, errors.New("req error")
	}

	return datas["result"].(bool), nil
}

// 获取原始memo
func GetRawMemo(data map[string]interface{}) string {
	respdata, err := common.RequestUrl("read_memo", []interface{}{data}, beego.AppConfig.String("walleturl"))
	if err != nil {
		log.Error(err)
		return ""
	}

	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {
		log.Debug(err)
		return ""
	}

	if datas["result"] == nil {
		return ""
	}

	return datas["result"].(string)
}

// 根据account_id 获取 account "1.2.26539" -> "waldo"
func GetAccountById(account_id string) string {
	respdata, err := common.RequestUrl("get_account", []interface{}{account_id}, beego.AppConfig.String("walleturl"))
	if err != nil {
		log.Error(err)
		return ""
	}

	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {
		log.Debug(err)
		return ""
	}

	if datas["result"] == nil {
		return ""
	}

	result := datas["result"].(map[string]interface{})
	return result["name"].(string)
}

// 根据account获取account_id 	"waldo" -> "1.2.26539"
func GetAccountIdByAccount(account string) string {
	respdata, err := common.RequestUrl("get_account_id", []interface{}{account}, beego.AppConfig.String("walleturl"))
	if err != nil {
		log.Error(err)
		return ""
	}

	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {
		log.Debug(err)
		return ""
	}

	if datas["result"] == nil {
		return ""
	}

	return datas["result"].(string)

}
