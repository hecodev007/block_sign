package model

import (
	"btcont/common/db"
	"errors"

	"github.com/shopspring/decimal"
)

type FcAddressAmount struct {
	Id            int64           `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinId        int             `json:"coin_id" xorm:"not null default 0 INT(10)"`
	CoinType      string          `json:"coin_type" xorm:"not null comment('币种名称') unique(coin_address) VARCHAR(100)"`
	Address       string          `json:"address" xorm:"not null comment('地址') unique(coin_address) VARCHAR(100)"`
	Amount        decimal.Decimal `json:"amount" xorm:"not null default 0.000000000000000000 comment('当前余额') DECIMAL(60,24)"`
	ForzenAmount  string          `json:"forzen_amount" xorm:"not null default 0.000000000000000000 comment('冻结金额') DECIMAL(60,24)"`
	Type          int             `json:"type" xorm:"not null default 0 comment('地址类型 1 冷地址 2 用户地址  3 手续费地址') TINYINT(3)"`
	AppId         int64           `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	PendingAmount string          `json:"pending_amount" xorm:"not null default 0.000000000000000000 comment('发送中的金额') DECIMAL(60,24)"`
}

func (fc *FcAddressAmount) Get(coin_type, addr string) (bool, error) {
	return db.SyncConn.Where("coin_type=? and address=?", coin_type, addr).Get(fc)
}

func (fc *FcAddressAmount) All(coin_type string) ([]*FcAddressAmount, error) {
	list := make([]*FcAddressAmount, 0)
	//db.SyncConn.ShowSQL(true)
	err := db.SyncConn.Where("coin_type=?", coin_type).OrderBy("amount desc").Find(&list)
	return list, err
}

func (fc *FcAddressAmount) AllAssert(coin_type string) ([]*FcAddressAmount, error) {
	list := make([]*FcAddressAmount, 0)
	//db.SyncConn.ShowSQL(true)
	err := db.SyncConn.Where("coin_type=?", coin_type).Where("amount != ?", "0").OrderBy("amount desc").Find(&list)
	return list, err
}
func (fc *FcAddressAmount) SetMount(coin_type, addr, amount string) error {
	ret, err := db.SyncConn.Exec("update `fc_address_amount` set amount=? where coin_type=? and address=?", amount, coin_type, addr)
	if err != nil {
		return err
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("更新出错,影响0条数据")
	}
	if affected > 1 {
		return errors.New("更新出错,影响多条条数据")
	}
	return nil
}

func (fc *FcAddressAmount) test(coin_type, addr, amount string) error {
	ret, err := db.SyncConn.Exec("update `fc_address_amount` set amount=? where coin_type=? and address=?", amount, coin_type, addr)
	//db.SyncConn.Update()
	if err != nil {
		return err
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("更新出错,影响0条数据")
	}
	if affected > 1 {
		return errors.New("更新出错,影响多条条数据")
	}
	return nil
}
