package dao

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"xorm.io/builder"
)

type FcFixTotalAmt struct {
	Amount string `json:"amount" xorm:"not null INT"`
}

func UpdateFcFixAddressDiscard() error {
	task := &entity.FcFixAddress{Status: 2}
	_, err := db.Conn.Where(builder.In("status", []int{0, 1})).Update(task)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFcFixAddressById(id int) error {
	model := &entity.FcFixAddress{Payed: 1}
	_, err := db.Conn.Where("id = ?", id).Update(model)
	if err != nil {
		return err
	}
	return nil
}

func InsertFcFixAddressBatch(models []entity.FcFixAddress) {
	db.Conn.Insert(models)
}

func FcFixTotalAmount() (*FcFixTotalAmt, error) {
	fee := &FcFixTotalAmt{}
	has, err := db.Conn.SQL("SELECT SUM(amount) as amount FROM fc_fix_address where status = ?", 1).Get(fee)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return fee, nil
}

func FcFixAddressActiveList() ([]*entity.FcFixAddress, error) {
	results := make([]*entity.FcFixAddress, 0)
	err := db.Conn.Table("fc_fix_address").Where("status = ?", 1).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcFindFixAddressList() ([]string, error) {
	res := make([]string, 0)
	if err := db.Conn.Table("fc_fix_address").Cols("address").Where("status = 1").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
