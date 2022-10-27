package daoqtum

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type BlockTXVin struct {
	Id           int64
	Height       int64
	Hash         string
	Txid         string
	Vintxid      string
	VinVoutindex int
	Address      string
	Amount       int64
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
func (b *BlockTXVin) Select(txid string) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select txid, height, hash, vin_txid, vin_voutindex from block_tx_vin where txid = ?", txid).Values(&maps)
	if err == nil && nums > 0 {
		for i := 0; i < len(maps); i++ {
			b.Txid = maps[i]["txid"].(string)
			b.Height = common.StrToInt64(maps[i]["height"].(string))
			b.Hash = maps[i]["hash"].(string)
			b.Vintxid = maps[i]["vin_txid"].(string)
			b.VinVoutindex = common.StrToInt(maps[i]["vin_voutindex"].(string))
		}
		return true, nil
	}
	return false, err
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
