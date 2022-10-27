package daoar

import (
	"github.com/astaxie/beego/orm"
)

type BlockTX struct {
	Id        int64
	Height int64
	Hash      string
	Txid      string
	Sysfee    string
	From  	string
	To 		string
	Amount 	string
	Memo 	string
	ContractAddress string
}

func NewBlockTX() *BlockTX {
	res := new(BlockTX)
	return res
}

// 删除区块
func DeleteFromBlockTX(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx where height >= ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 删除区块
func DeleteBlockTX(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx where height = ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// index 根据区块高度索引获取交易数据
func (b *BlockTX) SelectByIndex(index int64) error {
	return nil
}

// hash 获取交易数据
func (b *BlockTX) SelectByHash(hash string) error {
	return nil
}

// txid 获取交易数据
func (b *BlockTX) Exist(txid string) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select height from block_tx where txid = ?", txid).Values(&maps)
	if err!= nil {
		return false, nil
	}
	if nums > 0 {
		return true, err
	}
	return false, nil
}

// 插入交易数据
func (b *BlockTX) Update() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update  block_tx set  height = ?,hash = ?,sys_fee = ?,fromaccount = ?,toaccount = ?,amount = ? where txid= ?",
		b.Height, b.Hash, b.Sysfee, b.From, b.To, b.Amount,b.Txid).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 插入交易数据
func (b *BlockTX) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx(txid, height, hash, sys_fee, fromaccount, toaccount, amount, memo, contractaddress) values(?,?,?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Sysfee, b.From, b.To, b.Amount, b.Memo, b.ContractAddress).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
