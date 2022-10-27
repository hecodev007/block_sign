package daoycash

import (
	"github.com/astaxie/beego/orm"
)

type BlockTXVin struct {
	Id             	int64
	Height      	int64
	Hash           	string
	Txid           	string
	Vintxid        	string
	VinVoutindex   	int
	Address 	   	string
	Amount 		   	int64
}

func NewBlockTXVin() *BlockTXVin {
	res := new(BlockTXVin)
	return res
}

// 删除区块
func DeleteBlockTXVin(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx_vin where height >= ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// index 根据区块高度索引获取交易数据
func (b *BlockTXVin) SelectByIndex(index int64) error {
	return nil
}

// hash 获取交易数据
func (b *BlockTXVin) SelectByHash(hash string) error {
	return nil
}

// txid 获取交易数据
func (b *BlockTXVin) Select(txid string) error {
	return nil
}

// 插入交易数据
func (b *BlockTXVin) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx_vin(txid, height, hash, vin_txid, vin_voutindex) values(?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Vintxid, b.VinVoutindex).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
