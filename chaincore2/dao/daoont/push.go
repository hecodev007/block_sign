package daoont

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type NotifyResult struct {
	Id        int64
	UserID    int64
	Txid      string
	Num       int
	Timestamp int64
	Result    int
	Content   string
}

func NewNotifyResult() *NotifyResult {
	res := new(NotifyResult)
	return res
}

// txid 获取交易数据
func (b *NotifyResult) Select(id int64) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select id, userid, txid, num, timestamp, result from notifyresult where id = ?", id).Values(&maps)
	if err == nil && nums > 0 {
		b.Id = common.StrToInt64(maps[0]["id"].(string))
		b.UserID = common.StrToInt64(maps[0]["userid"].(string))
		b.Txid = maps[0]["txid"].(string)
		b.Num = common.StrToInt(maps[0]["num"].(string))
		b.Timestamp = common.StrToInt64(maps[0]["timestamp"].(string))
		b.Result = common.StrToInt(maps[0]["result"].(string))

		return true, err
	}
	return false, err
}

// 插入交易数据, 返回id
func (b *NotifyResult) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into notifyresult(userid, txid, num, timestamp, result, content) values(?,?,?,?,?,?)",
		b.UserID, b.Txid, b.Num, common.TimeToStr(b.Timestamp), b.Result, b.Content).Exec()
	if err == nil {
		id, err := res.LastInsertId()
		return id, err
	}

	return 0, err
}

// 插入交易数据, 返回id
func (b *NotifyResult) Update() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update notifyresult set num = ?, timestamp = ?, content = ?, result = ? where id = ?", b.Num, common.TimeToStr(b.Timestamp), b.Content, b.Result, b.Id).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
