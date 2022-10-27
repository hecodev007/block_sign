package rsk

import (
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	"time"
)

type BlockTX struct {
	Id              int64
	CoinName        string
	Txid            string
	ContractAddress string
	FromAddress     string
	ToAddress       string
	BlockHeight     int64
	BlockHash       string
	Amount          string
	Status          int8 //'0代表 失败,1代表成功,2代表上链成功但交易失败',
	GasUsed         int64
	GasPrice        int64
	Nonce           int64
	Input           string
	Logs            string
	Decimal         int8
	Timestamp       string
	CreateTime      string
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

//res, err := o.Raw("insert into block_tx(" +
//"coin_name, txid, contract_address, from_address, to_address, block_height, block_hash, amount, status,gas_used,gas_price,nonce,input,logs,decimal,timestamp,create_time) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
//b.CoinName, b.Txid, b.ContractAddress, b.FromAddress, b.ToAddress, b.BlockHeight, b.BlockHash, b.Amount, b.Status, b.GasUsed, b.GasPrice, b.Nonce, b.Input, b.Logs, b.Decimal, b.Timestamp, b.CreateTime).Exec()


// 插入交易数据
func (b *BlockTX) Insert() (int64, error) {
	o := orm.NewOrm()
	nowStr := common.TimeToStr(time.Now().Unix())
	res, err := o.Raw("insert into block_tx(" +
	"coin_name, txid, contract_address, from_address, to_address, block_height, block_hash, amount, status,gas_used,gas_price,nonce,input,logs,decimals,timestamp,create_time) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	b.CoinName, b.Txid, b.ContractAddress, b.FromAddress, b.ToAddress, b.BlockHeight, b.BlockHash, b.Amount, b.Status, b.GasUsed, b.GasPrice, b.Nonce, b.Input, b.Logs, b.Decimal, b.Timestamp,nowStr).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}
	return 0, err
}

