package cocos

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
)

// 关注地址列表, key: account_id
var WatchAddressList map[string][]*models.UserAddressInfo = make(map[string][]*models.UserAddressInfo)

// 用户通知信息列表
var UserWatchList map[int64]*models.UserAddressInfo = make(map[int64]*models.UserAddressInfo)

// account_id 对应的 account_id, key:account
var AccountMap map[string]string = make(map[string]string)

func InitWatchAddress() bool {
	o := orm.NewOrm()
	o.Using("user")
	{
		var maps []orm.Params
		nums, err := o.Raw("select id, trx_notify_url from users").Values(&maps)
		if err == nil && nums > 0 {
			for i := 0; i < int(nums); i++ {
				addrInfo := new(models.UserAddressInfo)
				addrInfo.UserID = common.StrToInt64(maps[i]["id"].(string))
				addrInfo.NotifyUrl = maps[i]["trx_notify_url"].(string)
				UserWatchList[addrInfo.UserID] = addrInfo
			}
		}
	}
	{
		var maps []orm.Params
		nums, err := o.Raw("select user_id, address from addresses where coin_type = ? and status = 'used'", beego.AppConfig.String("coin")).Values(&maps)
		if err == nil && nums > 0 {
			for i := 0; i < int(nums); i++ {
				account_id := GetAccountIdByAccount(maps[i]["address"].(string))
				log.Debug(account_id)
				if account_id == "" {
					return false
				}
				addrInfo := new(models.UserAddressInfo)
				addrInfo.UserID = common.StrToInt64(maps[i]["user_id"].(string))
				addrInfo.Address = maps[i]["address"].(string)
				addrInfo.NotifyUrl = UserWatchList[addrInfo.UserID].NotifyUrl
				addrInfo.AccountID = account_id

				AccountMap[account_id] = addrInfo.Address
				WatchAddressList[account_id] = append(WatchAddressList[account_id], addrInfo)
			}
		}
	}
	return true
}

func InsertWatchAddress(uid int64, account string, url string) bool {
	account_id := GetAccountIdByAccount(account)
	if account_id == "" {
		return false
	}

	if WatchAddressList[account_id] != nil {
		find := false
		for i := 0; i < len(WatchAddressList[account_id]); i++ {
			if WatchAddressList[account_id][i].UserID == uid {
				find = true
				WatchAddressList[account_id][i].NotifyUrl = url
				break
			}
		}
		if find {
			return true
		}
	}

	addrInfo := new(models.UserAddressInfo)
	addrInfo.UserID = uid
	addrInfo.Address = account
	addrInfo.NotifyUrl = url
	addrInfo.AccountID = account_id

	if UserWatchList[uid] == nil {
		UserWatchList[uid] = addrInfo
	}

	WatchAddressList[account_id] = append(WatchAddressList[account_id], addrInfo)
	AccountMap[account_id] = addrInfo.Address

	return true
}

func RemoveWatchAddress(address string) {
	delete(WatchAddressList, address)
}

func RemoveWatchAddressByUserId(uid int64, account string) bool {
	account_id := GetAccountIdByAccount(account)
	if account_id == "" {
		return false
	}

	for i := 0; i < len(WatchAddressList[account_id]); i++ {
		if WatchAddressList[account_id][i].UserID == uid {
			WatchAddressList[account_id] = append(WatchAddressList[account_id][:i], WatchAddressList[account_id][i+1:]...)
			break
		}
	}
	return true
}

func RemoveUserWatch(uid int64) {
	delete(UserWatchList, uid)
}

func UpdateWatchAddress(uid int64, url string) {
	if UserWatchList[uid] != nil {
		UserWatchList[uid].NotifyUrl = url
	}

	for _, v := range WatchAddressList {
		for i := 0; i < len(v); i++ {
			if v[i].UserID == uid {
				v[i].NotifyUrl = url
			}
		}
	}
}
