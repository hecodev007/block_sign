package po

import (
	"github.com/group-coldwallet/scanning-service/db"
)

type BlockTX struct {
	Id              int64  `json:"id,omitempty" gorm:"column:id"`
	Txid            string `json:"txid,omitempty" gorm:"column:txid"`
	Height          int64  `json:"height,omitempty" gorm:"column:height"`
	Hash            string `json:"hash,omitempty" gorm:"column:hash"`
	From            string `json:"from,omitempty" gorm:"column:from"`
	To              string `json:"to,omitempty" gorm:"column:to"`
	Amount          string `json:"amount,omitempty" gorm:"column:amount"`
	SysFee          string `json:"sys_fee,omitempty" gorm:"column:sys_fee"`
	Memo            string `json:"memo,omitempty" gorm:"column:memo"`
	ContractAddress string `json:"contract_address,omitempty" gorm:"column:contract_address"`
}

func (o *BlockTX) TableName() string {
	return "block_tx"
}

// 删除区块
func DeleteBlockTX(height int64) error {

	err := db.SyncDB.DB.Exec("delete from block_tx where height >= ?", height).Error
	if err != nil {
		return err
	}

	return nil
}

// hash 获取交易数据
func SelecBlockTxByTxid(hash string) (*BlockTX, error) {
	b := &BlockTX{}
	if err := db.SyncDB.DB.Where(" txid = ? ", hash).First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

// 插入交易数据
func InsertBlockTX(b *BlockTX) (int64, error) {

	if err := db.SyncDB.DB.Create(b).Error; err != nil {
		return 0, err
	}
	return b.Id, nil
}
