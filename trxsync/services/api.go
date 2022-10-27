package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/trxsync/common"
	"github.com/group-coldwallet/trxsync/models"
	"github.com/group-coldwallet/trxsync/models/po"
	"github.com/shopspring/decimal"
	"strings"
)

func (bs *BaseService) Info() (string, int64, error) {
	height, err := po.GetMaxBlockIndex()
	if err != nil {
		return "", 0, fmt.Errorf("get max block height error: %v", err)
	}
	return bs.Cfg.Sync.Name, height, nil
}

func (bs *BaseService) RepushTx(userId int64, txid string, height int64) (bool, error) {
	if userId <= 0 {
		return false, fmt.Errorf("use id is less than 0,userId=%d", userId)
	}
	if txid == "" {
		return false, errors.New("txid is null")
	}
	//if height<=0 {
	//	return fmt.Errorf("height is less than 0,height=%d",height)
	//}
	var (
		blockData *common.BlockData
		txData    *common.TxData
		err       error
	)
	if height == 0 {
		// 根据txid获取高度
		height, err = bs.scan.GetHeightByTxid(txid)
		if err != nil {
			return false, fmt.Errorf("补推错误，请添加高度重试： %v", err)
		}
	}
	log.Infof("补推处理高度： %d", height)
	if height > 0 {
		blockData, err = bs.scan.GetBlockByHeight(height)
		if err != nil {
			return false, fmt.Errorf("get block data error,err: %v", err)
		}
		hasTx := false

		if len(blockData.TxDatas) > 0 {
			for _, td := range blockData.TxDatas {
				if strings.TrimPrefix(td.Txid, "0x") == strings.TrimPrefix(txid, "0x") {
					hasTx = true
					//****** 需要把金额转换*********//
					var coinDecimal int32
					//第一步，判断是否有监听的合约
					if td.ContractAddress != "" {
						contractInfo, isHaveContract := bs.isContractTx(td.ContractAddress)
						if !isHaveContract {
							//没这监听这个合约
							continue
						}
						coinDecimal = int32(contractInfo.Decimal)
					} else {
						coinDecimal = td.MainDecimal
					}

					//2. ******更改amount以及fee的精度**********
					amount, _ := decimal.NewFromString(td.Amount)
					td.Amount = amount.Shift(-coinDecimal).String()
					fee, _ := decimal.NewFromString(td.Fee)
					td.Fee = fee.Shift(-td.MainDecimal).String()
					txData = td
					break
				}
			}
		}

		if len(blockData.TxIds) > 0 {
			for _, bd := range blockData.TxIds {
				if strings.TrimPrefix(bd, "0x") == strings.TrimPrefix(txid, "0x") {
					hasTx = true
					txData, err = bs.scan.GetTxData(blockData, txid, bs.isWatchAddress, bs.isContractTx)
					if err != nil {
						return false, fmt.Errorf("parse transaction error,tx id=%s,err: %v", txid, err)
					}
					if txData == nil {
						return false, fmt.Errorf("没有相关的交易：%s", txid)
					}
					break
				}
			}
		}

		if !hasTx {
			return false, fmt.Errorf("do not find this txid: %s", txid)
		}
	}

	log.Infof("tx data height: %d", txData.Height)
	if blockData == nil && txData.Height > 0 {
		log.Infof("================>高度处理： %d", txData.Height)
		blockData, err = bs.scan.GetBlockByHeight(txData.Height)
		if err != nil {
			return false, fmt.Errorf("get block data error,err: %v", err)
		}
		hasTx := false
		for _, bd := range blockData.TxIds {
			if bd == txid {
				hasTx = true
			}
		}
		if !hasTx {
			return false, fmt.Errorf("do not find this txid: %s", txid)
		}
	}
	if txData.IsFakeTx {
		return false, fmt.Errorf("发现一笔假充值,txid=%s", txid)
	}
	var (
		pushTxs      []models.PushAccountTx
		tmpWatchList map[string]bool = make(map[string]bool)
	)

	if txData.ToAddr != "" && bs.Watcher.IsWatchAddressExist(txData.ToAddr) {
		tmpWatchList[txData.ToAddr] = true
	}
	if txData.FromAddr != "" && bs.Watcher.IsWatchAddressExist(txData.FromAddr) {
		tmpWatchList[txData.FromAddr] = true
	}

	if len(tmpWatchList) > 0 {
		var pushtx models.PushAccountTx
		pushtx.From = txData.FromAddr
		pushtx.To = txData.ToAddr
		pushtx.Amount = txData.Amount
		pushtx.Fee = txData.Fee
		pushtx.Txid = txid
		pushtx.Memo = txData.Memo
		pushtx.Contract = txData.ContractAddress
		pushTxs = append(pushTxs, pushtx)
		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.CoinName = bs.Cfg.Sync.Name
		pushBlockTx.Height = blockData.Height
		pushBlockTx.Hash = blockData.Hash
		pushBlockTx.Confirmations = bs.Cfg.Sync.Confirmations + 1
		pushBlockTx.Time = common.Int64ToTime(blockData.Timestamp).Unix()
		pushBlockTx.Txs = pushTxs
		pushdata, err := json.Marshal(&pushBlockTx)
		if err != nil {
			return false, err
		}
		log.Infof("PushData: %s", string(pushdata))
		// 添加推送
		bs.AddPushTask(blockData.Height, txData.Txid, tmpWatchList, pushdata)
	} else {
		return true, fmt.Errorf("没有发现监听的地址，txid=%s", txid)
	}
	//如果成功 把	数据写入数据库
	bt, _ := po.SelecBlockTxByTxid(txid)
	if bt == nil {
		blockTx := &po.BlockTX{
			Height:          blockData.Height,
			Hash:            blockData.Hash,
			Txid:            txData.Txid,
			From:            txData.FromAddr,
			To:              txData.ToAddr,
			Amount:          txData.Amount,
			SysFee:          txData.Fee,
			Memo:            txData.Memo,
			ContractAddress: txData.ContractAddress,
		}
		po.InsertBlockTX(blockTx)
	}
	return false, nil
}
