package po

import (
	"rsksync/db"
	"time"
)

type NotifyResult struct {
	Id        int64     `json:"id,omitempty" gorm:"column:id"`
	UserID    int64     `json:"userid,omitempty" gorm:"column:userid"`
	Height    int64     `json:"height,omitempty" gorm:"column:height"`
	Txid      string    `json:"txid,omitempty" gorm:"column:txid"`
	Num       int       `json:"num,omitempty" gorm:"column:num"`
	Timestamp time.Time `json:"timestamp,omitempty" gorm:"column:timestamp"`
	Result    int       `json:"result,omitempty" gorm:"column:result"`
	Content   string    `json:"content,omitempty" gorm:"column:content"`
	Type      int       `json:"type,omitempty" gorm:"column:type"`
}

func (o *NotifyResult) TableName() string {
	return "notifyresult"
}

// txid 获取交易数据
func SelectNotifyResult(id int64) (*NotifyResult, error) {
	n := &NotifyResult{}
	if err := db.SyncDB.DB.Where(" id = ? ", id).First(n).Error; err != nil {
		return nil, err
	}
	return n, nil
}

// 插入交易数据, 返回id
func InsertNotifyResult(n *NotifyResult) (int64, error) {
	if err := db.SyncDB.DB.Create(n).Error; err != nil {
		return 0, err
	}
	return n.Id, nil
}

// 插入交易数据, 返回id
func UpdateNotifyResult(n *NotifyResult) error {
	if err := db.SyncDB.DB.Save(n).Error; err != nil {
		return err
	}
	return nil
}

// 高度 获取关注用户
func SelectWatchHeight(height int64) (map[int64]int, error) {

	var ws []*NotifyResult
	//o := orm.NewOrm()
	//var maps []orm.Params
	//nums, err := o.Raw("select userid from user_watch_confir where height = ?", height).Values(&maps)
	//if err == nil && nums > 0 {
	//	for i := 0; i < len(maps); i++ {
	//		uid := common.StrToInt64(maps[i]["userid"].(string))
	//		users = append(users, uid)
	//	}
	//}
	users := make(map[int64]int)
	err := db.SyncDB.DB.Select([]string{"userid"}).Where("height = ? ", height).Find(&ws).Error
	if err != nil || len(ws) == 0 {
		return nil, err
	}

	for i := 0; i < len(ws); i++ {
		users[ws[i].UserID] = 1
	}
	return users, err
}
