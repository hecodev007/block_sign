package chainx

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: address
var WatchAddressList map[string][]*models.UserAddressInfo = make(map[string][]*models.UserAddressInfo)

// 用户通知信息列表
var UserWatchList map[int64]*models.UserAddressInfo = make(map[int64]*models.UserAddressInfo)

func InitWatchAddress() {
	if !beego.AppConfig.DefaultBool("enablewatchaddress", false) {
		return
	}

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
				addrInfo := new(models.UserAddressInfo)
				addrInfo.UserID = common.StrToInt64(maps[i]["user_id"].(string))
				addrInfo.Address = maps[i]["address"].(string)
				addrInfo.NotifyUrl = UserWatchList[addrInfo.UserID].NotifyUrl
				WatchAddressList[addrInfo.Address] = append(WatchAddressList[addrInfo.Address], addrInfo)
			}
		}
	}
}

func InsertWatchAddress(uid int64, address string, url string) {
	if !beego.AppConfig.DefaultBool("enablewatchaddress", false) {
		return
	}

	if WatchAddressList[address] != nil {
		find := false
		for i := 0; i < len(WatchAddressList[address]); i++ {
			if WatchAddressList[address][i].UserID == uid {
				find = true
				WatchAddressList[address][i].NotifyUrl = url
				break
			}
		}
		if find {
			return
		}
	}

	addrInfo := new(models.UserAddressInfo)
	addrInfo.UserID = uid
	addrInfo.Address = address
	addrInfo.NotifyUrl = url
	WatchAddressList[address] = append(WatchAddressList[addrInfo.Address], addrInfo)

	if UserWatchList[uid] == nil {
		UserWatchList[uid] = addrInfo
	}
}

func RemoveWatchAddress(address string) {
	delete(WatchAddressList, address)
}

func RemoveWatchAddressByUserId(uid int64, address string) {
	for i := 0; i < len(WatchAddressList[address]); i++ {
		if WatchAddressList[address][i].UserID == uid {
			WatchAddressList[address] = append(WatchAddressList[address][:i], WatchAddressList[address][i+1:]...)
			break
		}
	}
}

func RemoveUserWatch(uid int64) {
	delete(UserWatchList, uid)
}

func UpdateWatchAddress(uid int64, url string) {
	if !beego.AppConfig.DefaultBool("enablewatchaddress", false) {
		return
	}

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
