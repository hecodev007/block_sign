package dao

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
)

//插入数据
func FcTransactionsStateInsert(fs *entity.FcTransactionsState) error {
	_, err := db.Conn.Insert(fs)
	return err
}

//查询单条
func FcTransactionsStateGetOne(outOrderId string) (*entity.FcTransactionsState, error) {
	dataRow := &entity.FcTransactionsState{}
	has, err := db.Conn.Where("out_orderid = ?", outOrderId).Get(dataRow)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return dataRow, nil
}

//修改数据,并且增加错误次数
//可选项更新msg，callbackMsg
func FcTransactionsStateUpdateAddErr(id int, status entity.FcTransactionsStateCode, msg, callbackMsg string) error {
	dataRow := &entity.FcTransactionsState{}
	has, err := db.Conn.Id(id).Get(dataRow)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("Not Fount!")
	}
	if dataRow.Status == entity.FcTransactionsStatesSuccess {
		return fmt.Errorf("ID :%d,Already successful", id)
	}
	if msg != "" {
		dataRow.Msg = msg
	}
	if callbackMsg != "" {
		dataRow.CallbackMsg = callbackMsg
	}
	if status == entity.FcTransactionsStatesSuccess {
		dataRow.PushStatus = 0
	} else if status == entity.FcTransactionsStatesFail {
		dataRow.PushStatus = 2
	}

	dataRow.Status = status
	dataRow.RetryNum = dataRow.RetryNum + 1
	dataRow.Lastmodify = util.GetChinaTimeNow()
	affected, err := db.Conn.Id(dataRow.Id).
		Cols("callback_msg", "msg", "lastmodify", "push_status", "status", "retry_num").
		Update(dataRow)
	if affected == 0 {
		err = errors.New("No Data Update")
	}
	return err
}

func FcTransactionsStateUpdateState(id int, status entity.FcTransactionsStateCode, callbackMsg string) error {
	dataRow := &entity.FcTransactionsState{}
	has, err := db.Conn.Id(id).Get(dataRow)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("Not Fount!")
	}
	if callbackMsg != "" {
		dataRow.CallbackMsg = callbackMsg
	}
	dataRow.Status = status
	dataRow.Lastmodify = util.GetChinaTimeNow()
	if status == entity.FcTransactionsStatesSuccess {
		dataRow.PushStatus = 0
	} else if status == entity.FcTransactionsStatesFail {
		dataRow.PushStatus = 2
	}
	affected, err := db.Conn.Id(dataRow.Id).
		Cols("callback_msg", "lastmodify", "push_status", "status").
		Update(dataRow)
	if affected == 0 {
		err = errors.New("No Data Update")
	}
	return err
}

//寻找可以推送的记录
func FcTransactionsStateFindByStatus(status entity.FcTransactionsStateCode, lessThanRetryNum, limit int) (
	[]*entity.FcTransactionsState, error) {
	results := make([]*entity.FcTransactionsState, 0)
	err := db.Conn.Where("status = ? and retry_num < ?", status, lessThanRetryNum).OrderBy("id asc").Limit(limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
