package avax

import (
	"avaxDataServer/common/log"
	"avaxDataServer/db"
	"avaxDataServer/utils/avax"
	"errors"
	"time"
)

type BlockTxVout struct {
	Id                     int       `xorm:"not null pk autoincr INT(20)"`
	Outputid               string    `xorm:"VARCHAR(45)"`
	Transactionid          string    `xorm:"not null VARCHAR(100)"`
	Outputindex            int       `xorm:"not null INT(20)"`
	Assetid                string    `xorm:"VARCHAR(45)"`
	Outputtype             int       `xorm:"INT(20)"`
	Amount                 string    `xorm:"not null DECIMAL(50,0)"`
	Locktime               int       `xorm:"INT(20)"`
	Threshold              int       `xorm:"INT(20)"`
	Address                string    `xorm:"not null VARCHAR(100)"`
	Createdat              time.Time `xorm:"TIMESTAMP"`
	SpendTxid              string    `xorm:"not null VARCHAR(100)"`
	Redeemingtransactionid string    `xorm:"VARCHAR(45)"`
	Height                 int       `xorm:"INT(20)"`
	Timestamp              time.Time `xorm:"TIMESTAMP"`
	Forked                 int       `xorm:"comment('是否分叉成孤块') INT(20)"`
}

func DeleteBlockTXVout(h int64) error {
	txout := new(BlockTxVout)
	_, err := db.SyncConn.Where("height >=?", h).Delete(txout)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func BatchInsertBlockTXVouts(txouts []*avax.Output, height int64) error {
	blockTxOuts := make([]*BlockTxVout, 0, len(txouts))
	for _, v := range txouts {
		blockTxout := new(BlockTxVout)
		blockTxout.Outputid = v.ID
		blockTxout.Transactionid = v.TransactionID
		blockTxout.Outputindex = int(v.OutputIndex)
		blockTxout.Assetid = v.AssetID
		blockTxout.Outputtype = int(v.OutputType)
		blockTxout.Amount = v.Amount
		blockTxout.Locktime = int(v.Locktime)
		blockTxout.Threshold = int(v.Threshold)

		if len(v.Addresses) == 1 {
			blockTxout.Address = string(v.Addresses[0])
		}
		blockTxout.Createdat = v.CreatedAt
		blockTxout.Redeemingtransactionid = v.RedeemingTransactionID
		blockTxout.Height = int(height)
		blockTxout.Timestamp = time.Now()
		blockTxOuts = append(blockTxOuts, blockTxout)
	}
	_, err := db.SyncConn.Insert(blockTxOuts)
	return err
}
func BatchInsertBlockTXins(txins []*avax.Input, height int64) error {
	blockTxOuts := make([]*BlockTxVout, 0, len(txins))
	for _, value := range txins {
		v := value.Output
		blockTxout := new(BlockTxVout)
		blockTxout.Outputid = v.ID
		blockTxout.Transactionid = v.TransactionID
		blockTxout.Outputindex = int(v.OutputIndex)
		blockTxout.Assetid = v.AssetID
		blockTxout.Outputtype = int(v.OutputType)
		blockTxout.Amount = v.Amount
		blockTxout.Locktime = int(v.Locktime)
		blockTxout.Threshold = int(v.Threshold)

		if len(v.Addresses) == 1 {
			blockTxout.Address = string(v.Addresses[0])
		}
		blockTxout.Createdat = v.CreatedAt
		blockTxout.Redeemingtransactionid = v.RedeemingTransactionID
		blockTxout.Height = int(height)
		blockTxout.Timestamp = time.Now()
		blockTxOuts = append(blockTxOuts, blockTxout)
	}
	_, err := db.SyncConn.Insert(blockTxOuts)
	return err
}

func BatchUpdateBlockTXVouts(txouts []*BlockTxVout) error {
	for _, v := range txouts {
		if _, err := db.SyncConn.Update(v); err != nil {
			return err
		}
	}
	return nil
}
func SelectBlockTXVout(txid string, index uint64) (*BlockTxVout, error) {
	vout := new(BlockTxVout)
	has, err := db.SyncConn.Where("Transactionid=? and index=?", txid, index).Get(vout)
	if !has {
		return nil, errors.New("not find")
	}
	return vout, err
}
