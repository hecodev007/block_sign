package daoksm

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type WatchHeight struct {
	Id     int64
	UserID int64
	Height int64
}

func NewWatchHeight() *WatchHeight {
	res := new(WatchHeight)
	return res
}

// 高度 获取关注用户
func SelectWatchHeight(height int64) ([]int64, error) {
	var users []int64
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select userid from user_watch_confir where height = ?", height).Values(&maps)
	if err == nil && nums > 0 {
		for i := 0; i < len(maps); i++ {
			uid := common.StrToInt64(maps[i]["userid"].(string))
			users = append(users, uid)
		}
	}
	return users, err
}

// 插入, 返回id
func InsertWatchHeight(uid int64, height int64) error {
	o := orm.NewOrm()
	_, err := o.Raw("insert into user_watch_confir(userid, height) values(?,?)", uid, height).Exec()
	return err
}
