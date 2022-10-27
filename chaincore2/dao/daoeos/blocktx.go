package daoeos

import (
	"github.com/astaxie/beego/orm"
)

type BlockTX struct {
	Id        int64
	Height int64
	Hash      string
	Txid      string
	Sysfee    float64
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
func (b *BlockTX) Select(txid string) (bool, error) {
	return false, nil
}

// 插入交易数据
func (b *BlockTX) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx(txid, height, hash, sys_fee, fromaccount, toaccount, amount, memo, contract) values(?,?,?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Sysfee, b.From, b.To, b.Amount, b.Memo, b.ContractAddress).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
