package eth

import (
	"bytes"
	"github.com/shopspring/decimal"
	"rsksync/db"
	"strings"
	"time"
)

type BlockTX struct {
	Id              int64           `json:"id,omitempty" gorm:"column:id"`
	CoinName        string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	Txid            string          `json:"txid,omitempty" gorm:"column:txid"`
	ContractAddress string          `json:"contract_address,omitempty" gorm:"column:contract_address"`
	FromAddress     string          `json:"from_address,omitempty" gorm:"column:from_address"`
	ToAddress       string          `json:"to_address,omitempty" gorm:"column:to_address"`
	BlockHeight     int64           `json:"block_height,omitempty" gorm:"column:block_height"`
	BlockHash       string          `json:"block_hash,omitempty" gorm:"column:block_hash"`
	Amount          decimal.Decimal `json:"amount,omitempty" gorm:"column:amount"`
	Status          int             `json:"status,omitempty" gorm:"column:status"`
	GasUsed         int64           `json:"gas_used,omitempty" gorm:"column:gas_used"`
	GasPrice        int64           `json:"gas_price,omitempty" gorm:"column:gas_price"`
	Nonce           int             `json:"nonce,omitempty" gorm:"column:nonce"`
	Input           string          `json:"input,omitempty" gorm:"column:input"`
	Decimal         int             `json:"decimal,omitempty" gorm:"column:decimal"`
	Logs            string          `json:"logs,omitempty" gorm:"column:logs"`
	Timestamp       time.Time       `json:"timestamp,omitempty" gorm:"column:timestamp"`
	CreateTime      time.Time       `json:"create_time,omitempty" gorm:"column:create_time"`
	ToAmount        decimal.Decimal `json:"toAmount" gorm:"-"` //sta这个币种会销毁币种.临时加结构处理

}

func (o *BlockTX) TableName() string {
	return "block_tx"
}

func BatchInsertBlockTX(bs []*BlockTX) error {
	rn := len(bs)
	if rn == 0 {
		return nil
	}

	vals := make([]interface{}, 0, len(bs)*20)
	for _, b := range bs {
		vals = append(vals, b.CoinName, b.Txid, b.ContractAddress, b.FromAddress, b.ToAddress, b.BlockHeight, b.BlockHash,
			b.Amount, b.Status, b.GasUsed, b.GasPrice, b.Nonce, b.Input, b.Decimal, b.Logs, b.Timestamp, b.CreateTime)
	}
	buf := bytes.NewBufferString(`INSERT INTO block_tx(coin_name,txid,contract_address,from_address,to_address,block_height,block_hash,amount,status,gas_used,gas_price,nonce,input,decimal,logs,timestamp,create_time) VALUES `)
	buf.WriteString(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`, rn))
	buf.Truncate(buf.Len() - 1)
	err := db.SyncDB.DB.Exec(buf.String(), vals...).Error
	if err != nil {
		return err
	}
	return nil
}

// 删除区块
func DeleteBlockTX(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_tx where block_height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// index 根据区块高度索引获取交易数据
func SelectBlockTXsByIndex(blkheight int64) (bs []*BlockTX, err error) {

	if err = db.SyncDB.DB.Where(" block_height = ? ", blkheight).Find(bs).Error; err != nil {
		return nil, err
	}
	return bs, nil
}

// hash 获取交易数据
func SelecBlockTxByHash(hash string) (*BlockTX, error) {
	b := &BlockTX{}
	if err := db.SyncDB.DB.Where(" txid = ? ", hash).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// txid 获取交易数据
//func Select(txid string)  (*BlockTX,error)  {
//	//o := orm.NewOrm()
//	//var maps []orm.Params
//	//nums, err := o.Raw("select txid, block_height, block_hash, contract_address, from_address, to_address, nonce, gas_used, gas_price, amount, input, status, successed from contract_tx where txid = ?", txid).Values(&maps)
//	//if err == nil && nums > 0 {
//	//	b.Txid = maps[0]["txid"].(string)
//	//	b.BlockHeight = common.StrToInt64(maps[0]["block_height"].(string))
//	//	b.BlockHash = maps[0]["block_hash"].(string)
//	//	b.ContractAddress = maps[0]["contract_address"].(string)
//	//	b.FromAddress = maps[0]["from_address"].(string)
//	//	b.ToAddress = maps[0]["to_address"].(string)
//	//	b.Amount = maps[0]["amount"].(decimal.Decimal)
//	//	b.Nonce = maps[0]["nonce"].(int)
//	//	b.GasPrice = maps[0]["gas_price"].(int64)
//	//	b.GasUsed = maps[0]["gas_used"].(int64)
//	//	b.Input = maps[0]["input"].(string)
//	//	b.Status = maps[0]["status"].(int)
//	//	b.Success = maps[0]["success"].(int)
//	//	return nil
//	//}
//	//return err
//	b := &BlockTX{}
//	if err := MysqlDB.Where(" txid = ? ", hash).First(b).Error; err != nil {
//		return nil, err
//	}
//	return b, nil
//}

// 插入交易数据
func InsertBlockTX(b *BlockTX) (int64, error) {

	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
