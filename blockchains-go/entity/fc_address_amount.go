package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

type FcAddressAmount struct {
	Id            int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinId        int    `json:"coin_id" xorm:"not null default 0 INT(10)"`
	CoinType      string `json:"coin_type" xorm:"not null comment('币种名称') unique(coin_address) VARCHAR(100)"`
	Address       string `json:"address" xorm:"not null comment('地址') unique(coin_address) VARCHAR(190)"`
	Amount        string `json:"amount" xorm:"not null default 0.000000000000000000 comment('当前余额') DECIMAL(60,24)"`
	ForzenAmount  string `json:"forzen_amount" xorm:"not null default 0.000000000000000000 comment('冻结金额') DECIMAL(60,24)"`
	Type          int    `json:"type" xorm:"not null default 0 comment('地址类型 1 冷地址 2 用户地址  3 手续费地址') TINYINT(3)"`
	AppId         int64  `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	PendingAmount string `json:"pending_amount" xorm:"not null default 0.000000000000000000 comment('发送中的金额') DECIMAL(60,24)"`
}

func (o *FcAddressAmount) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
func (o *FcAddressAmount) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcAddressAmount) Update(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Update(o)
}
func (o FcAddressAmount) Exist(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Exist(o)
}
func (o FcAddressAmount) Count(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Count(o)
}
func (o FcAddressAmount) Delete(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Delete(o)
}
func (o FcAddressAmount) Find(cond builder.Cond, limit int) ([]*FcAddressAmount, error) {
	res := make([]*FcAddressAmount, 0)
	if err := db.Conn.Where(cond).Desc("id").Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o FcAddressAmount) FindAddress(cond builder.Cond, limit int) ([]string, error) {
	res := make([]string, 0)
	if err := db.Conn.Table("fc_address_amount").Cols("address").Where(cond).Desc("amount").Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}

//write by jun 2020/4/29
func (o FcAddressAmount) FindAddressAndAmount(cond builder.Cond, limit int) ([]*FcAddressAmount, error) {
	res := make([]*FcAddressAmount, 0)
	if err := db.Conn.Table("fc_address_amount").Where(cond).Desc("amount").Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o FcAddressAmount) SumAmount(cond builder.Cond) (decimal.Decimal, error) {
	total, err := db.Conn.Where(cond).Sum(&o, "amount")
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(total), nil
}

/*
2021-02-25
升序查找地址
write by flynn
*/
func (o FcAddressAmount) FindAddressAndAmountByAsc(cond builder.Cond, limit int) ([]*FcAddressAmount, error) {
	res := make([]*FcAddressAmount, 0)
	if err := db.Conn.Table("fc_address_amount").Where(cond).Asc("amount").Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
