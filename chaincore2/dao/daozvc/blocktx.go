package daozvc

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
)

type BlockTX struct {
	Id       int64
	Height   int64
	Hash     string
	Txid     string
	GasLimit int64
	GasPrice int64
	From     string
	To       string
	Amount   string
	Memo     string
	Contract string
	GasUsed  float64
}

func NewBlockTX() *BlockTX {
	res := new(BlockTX)
	return res
}

// 删除区块
func DeleteBlockTX(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_tx where height >= ?", height).Exec()
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
	nums, err := o.Raw("select txid, height, hash, contract_address, fromaddress, toaddress,amount,gaslimit,gasprice,gasused,memo from block_tx where txid = ?", txid).Values(&maps)
	if err == nil && nums > 0 {
		b.Txid = maps[0]["txid"].(string)
		b.Height = common.StrToInt64(maps[0]["height"].(string))
		b.Hash = maps[0]["hash"].(string)
		b.Contract = maps[0]["contract_address"].(string)
		b.From = maps[0]["fromaddress"].(string)
		b.To = maps[0]["toaddress"].(string)
		b.Amount = maps[0]["amount"].(string)
		b.GasLimit = common.StrToInt64(maps[0]["gaslimit"].(string))
		b.GasPrice = common.StrToInt64(maps[0]["gasprice"].(string))
		b.GasUsed = common.StrToFloat64(maps[0]["gasused"].(string))
		b.Memo = maps[0]["memo"].(string)
		return true, err
	}
	return false, err
}

// 插入交易数据
func (b *BlockTX) Insert() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_tx(txid, height, hash, contract_address, fromaddress, toaddress,amount,gaslimit,gasprice,gasused,memo) values(?,?,?,?,?,?,?,?,?,?,?)",
		b.Txid, b.Height, b.Hash, b.Contract, b.From, b.To, b.Amount, b.GasLimit, b.GasPrice, b.GasUsed, b.Memo).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}
