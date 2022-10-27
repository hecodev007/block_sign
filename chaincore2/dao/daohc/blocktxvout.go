package daohc

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type BlockTXVout struct {
	Id          int64
	Height      int64
	Hash        string
	Txid        string
	Voutn       int
	Voutvalue   int64
	Voutaddress string

	Invaild int
	Status  int
}

func NewBlockTXVout() *BlockTXVout {
	res := new(BlockTXVout)
	return res
}

// 删除区块
func DeleteFromBlockTXVout(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx_vout where height >= ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 删除区块
func DeleteBlockTXVout(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx_vout where height = ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// index 根据区块高度索引获取交易数据
func (b *BlockTXVout) SelectByIndex(index int64) error {
	return nil
}

// hash 获取交易数据
func (b *BlockTXVout) SelectByHash(hash string) error {
	return nil
}

// txid 获取是否存在
func (b *BlockTXVout) Count(txid string, voutn int) int64 {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select height from block_tx_vout where txid = ? and vout_n = ?", txid, voutn).Values(&maps)
	if err == nil && nums > 0 {
		return nums
	}
	return 0
}

// txid 获取交易数据
func (b *BlockTXVout) Select(txid string, voutn int64) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select txid, height, hash, vout_n, vout_value, vout_address, invaild, status from block_tx_vout where txid = ? and vout_n = ?", txid, voutn).Values(&maps)
	if err == nil && nums > 0 {
		b.Txid = maps[0]["txid"].(string)
		b.Height = common.StrToInt64(maps[0]["height"].(string))
		b.Hash = maps[0]["hash"].(string)
		b.Voutn = common.StrToInt(maps[0]["vout_n"].(string))
		b.Voutvalue = common.StrToInt64(maps[0]["vout_value"].(string))
		b.Voutaddress = maps[0]["vout_address"].(string)

		b.Invaild = common.StrToInt(maps[0]["invaild"].(string))
		b.Status = common.StrToInt(maps[0]["status"].(string))
		return true, err
	}
	return false, err
}

// 插入交易数据
func (b *BlockTXVout) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx_vout(txid, height, hash, vout_n, vout_value, vout_address, invaild, status) values(?,?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Voutn, b.Voutvalue, b.Voutaddress, b.Invaild, b.Status).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 设置是否有效
func (b *BlockTXVout) UpdateInvaild(val int, txid string, voutn int) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_tx_vout set invaild = ? where txid = ? and vout_n = ?", val, txid, voutn).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 设置状态
func (b *BlockTXVout) UpdateStatus(status int, txid string, voutn int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_tx_vout set status = ? where txid = ? and vout_n = ?", status, txid, voutn).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
