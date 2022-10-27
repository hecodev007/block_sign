package po

import (
	"atpDataServer/common/db"
	"time"
)

type NotifyResult struct {
	Id        int64     `json:"id,omitempty" gorm:"column:id"`
	Userid    int64     `json:"userid,omitempty" gorm:"column:userid"`
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
func SelectNotifyResult(id int64) (ret *NotifyResult, err error) {
	ret = &NotifyResult{}
	_, err = db.SyncConn.Where("id=?", id).Get(ret)
	return
}

// 插入交易数据, 返回id
func InsertNotifyResult(n *NotifyResult) (id int64, err error) {
	_, err = db.SyncConn.InsertOne(n)
	return n.Id, err
}

// 插入交易数据, 返回id
func UpdateNotifyResult(n *NotifyResult) error {
	//db.SyncConn.ShowSQL(true)
	//defer db.SyncConn.ShowSQL(false)
	n.Timestamp = time.Now()
	if affected, err := db.SyncConn.Where("id=?", n.Id).Update(n); err != nil {
		return err
	} else if affected > 0 {
		return nil
	} else {
		_, err = db.SyncConn.InsertOne(n)
		return err
	}
}

// 高度 获取关注用户
func SelectWatchHeight(height int64) (map[int64]bool, error) {

	var ws []*NotifyResult
	if err := db.SyncConn.Where("height=?", height).Find(&ws); err != nil {
		return nil, err
	}
	users := make(map[int64]bool)
	for i := 0; i < len(ws); i++ {
		users[ws[i].Userid] = true
	}
	return users, nil
}
