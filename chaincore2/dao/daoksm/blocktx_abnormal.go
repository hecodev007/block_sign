package daoksm

import (
	"github.com/astaxie/beego/orm"
)

//异常数据
type BlockTXAbnormal struct {
	Id              int64
	Height          int64
	Hash            string
	Txid            string
	Sysfee          string // 带精度
	From            string
	To              string
	Amount          string // 带精度
	Memo            string
	ContractAddress string
	SucInfo         string
}

func NewBlockTXAbnormal() *BlockTXAbnormal {
	res := new(BlockTXAbnormal)
	return res
}

// 删除区块
func DeleteFromBlockTXAbnormal(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx_abnomal where height >= ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 删除区块
func DeleteBlockTXAbnormal(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx_abnomal where height = ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// index 根据区块高度索引获取交易数据
func (b *BlockTXAbnormal) SelectByIndex(index int64) error {
	return nil
}

// hash 获取交易数据
func (b *BlockTXAbnormal) SelectByHash(hash string) error {
	return nil
}

// txid 获取交易数据
func (b *BlockTXAbnormal) Select(txid string) (bool, error) {
	return false, nil
}

func (b *BlockTXAbnormal) Exist(txid string) (bool, error) {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select height from block_tx_abnomal where txid = ?", txid).Values(&maps)
	if err!= nil {
		return false, nil
	}
	if nums > 0 {
		return true, err
	}
	return false, nil
}


// 插入交易数据
func (b *BlockTXAbnormal) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx_abnomal(txid, height, hash, sys_fee, fromaccount, toaccount, amount, memo, contractaddress,suc_info) values(?,?,?,?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Sysfee, b.From, b.To, b.Amount, b.Memo, b.ContractAddress,b.SucInfo).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
