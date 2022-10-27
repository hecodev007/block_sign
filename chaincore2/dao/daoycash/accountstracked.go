package daoycash

import (
	"github.com/astaxie/beego/orm"
)

type AccountsTracked struct {
	Highindex    int64
	Txid         string
	Voutn        int
	Voutvalue    float64
	Voutaddress  string
	ScriptPubKey string
}

func NewAccountsTracked() *AccountsTracked {
	res := new(AccountsTracked)
	return res
}

// 删除指定高度数据
func RemoveAccountsTracked(height int64) {
	o := orm.NewOrm()
	o.Raw("delete from accounts_tracked where highindex >= ?", height).Exec()
}

// 删除跟踪数据
func DeleteAccountsTracked(addr string, txid string, n int) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update accounts_tracked set isdelete = 1 where vout_address = ? and txid = ? and vout_n = ?",
		addr, txid, n).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 插入跟踪数据
func (b *AccountsTracked) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into accounts_tracked(highindex, txid, vout_n, vout_value, vout_address, scriptpubkey) values(?,?,?,?,?,?)",
		b.Highindex, b.Txid, b.Voutn, b.Voutvalue, b.Voutaddress, b.ScriptPubKey).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
