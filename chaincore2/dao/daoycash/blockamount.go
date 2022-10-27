package daoycash

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type BlockAmount struct {
	Id      int64
	Address string
	Amount  int64
}

// new
func NewBlockAmount() *BlockAmount {
	res := new(BlockAmount)
	return res
}

// update
func UpdateBlockAmount(addr string, amount int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_amount set amount = amount + ? where address = ?", amount, addr).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// find and set
func FindSetBlockAmount(addr string, amount int64) (int64, error) {
	blockamount := NewBlockAmount()
	ret, err := blockamount.Find(addr)
	if err != nil {
		beego.Error(err)
		return 0, err
	}
	if ret {
		return blockamount.Update(amount)
	} else {
		blockamount.Address = addr
		blockamount.Amount = amount
		return blockamount.Insert()
	}
}

// 查找资产地址是否存在
func (b *BlockAmount) Find(addr string) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select id, address, amount from block_amount where address = ?", addr).Values(&maps)
	if err == nil && nums > 0 {
		b.Id = common.StrToInt64(maps[0]["id"].(string))
		b.Address = maps[0]["address"].(string)
		b.Amount = common.StrToInt64(maps[0]["amount"].(string))
		return true, err
	}
	return false, err
}

// 写入资产
func (b *BlockAmount) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_amount(address, amount) values(?,?)",
		b.Address, b.Amount).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 更新资产
func (b *BlockAmount) Update(val int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_amount set amount = amount + ? where id = ?",
		val, b.Id).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 设置资产
func (b *BlockAmount) Set(val int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_amount set amount = ? where id = ?", val, b.Id).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
