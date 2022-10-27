package daohc

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type BlockTX struct {
	Id        int64
	Height    int64
	Hash      string
	Txid      string
	Sysfee    float64
	Vincount  int
	Voutcount int
	Coinbase  int // 是否是矿工交易
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
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select txid, height, hash, sys_fee, vincount, voutcount, coinbase from block_tx where txid = ?", txid).Values(&maps)
	if err == nil && nums > 0 {
		b.Txid = maps[0]["txid"].(string)
		b.Height = common.StrToInt64(maps[0]["height"].(string))
		b.Hash = maps[0]["hash"].(string)
		b.Sysfee = common.StrToFloat64(maps[0]["sys_fee"].(string))
		b.Vincount = common.StrToInt(maps[0]["vincount"].(string))
		b.Voutcount = common.StrToInt(maps[0]["voutcount"].(string))
		b.Coinbase = common.StrToInt(maps[0]["coinbase"].(string))
		return true, err
	}
	return false, err
}

// 插入交易数据
func (b *BlockTX) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx(txid, height, hash, sys_fee, vincount, voutcount, coinbase) values(?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Sysfee, b.Vincount, b.Voutcount, b.Coinbase).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
