package cfx

import (
	"cfxDataServer/common"
	"cfxDataServer/common/conf"
	"cfxDataServer/common/log"
	dao "cfxDataServer/models/po/cfx"
	"cfxDataServer/services"
	rpc "cfxDataServer/utils/cfx"
	"errors"
	"fmt"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/shopspring/decimal"
	"math/big"
	"sync"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(781001)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
	return count - 18, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	//if height != 2071702{
	//	time.Sleep(time.Second*10000)
	//	return nil, errors.New("123")
	//}
	//starttime := time.Now()
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Infof("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}
	//txs2,_:=json.Marshal(block)
	//log.Info(height,string(txs2))
	//log.Info(block.Transactions[1].Status,block.Transactions[1].TransactionIndex)
	//panic("")

	for _, blockHash := range block.RefereeHashes {
		//if len(block.RefereeHashes)>1{
		//	log.Info("blockHash",height,blockHash)
		//}
	GetBlockByHash:
		tempblock, err := s.GetBlockByHash(string(blockHash))
		if err != nil {
			log.Info(err.Error())
			goto GetBlockByHash
		}
		//txs,_:=json.Marshal(tempblock.Transactions)
		//log.Info(height,string(txs))
		block.Transactions = append(block.Transactions, tempblock.Transactions...)
	}
	//log.Info(height, bestHeight, len(block.Transactions), block.Hash)

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Transactions))
	//bs, _ := json.Marshal(block)
	//log.Info(string(bs))
	//if has, err := dao.BlockHashExist(string(block.Hash)); err != nil {
	//	return nil, fmt.Errorf("database err")
	//} else if has {
	//	return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	//}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height.ToInt().Int64(),
			Hash:              string(block.Hash),
			Previousblockhash: string(block.ParentHash),
			Nextblockhash:     "",
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	tmptxs := make(map[string]bool)
	for _, tx := range block.Transactions {
		if tmptxs[string(tx.Hash)] == true {
			continue
		}
		tmptxs[string(tx.Hash)] = true
		//log.Info(tx.Hash)
		if tx.To == nil {
			continue
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, s.watch, &tx, string(block.Hash), block.Height.ToInt().Int64()); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			//tx_json,_:=json.Marshal(txInfo)
			//log.Info(string(tx_json))
			lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)
			lock.Unlock()
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", block.Height.ToInt().Int64(), string(ts), "task")
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, watch *services.WatchControl, tx *rpc.Transaction, blockhash string, blockheight int64) (*dao.BlockTx, error) {
	//log.Info(blockheight)
	//RpcClient.IsUser(tx.To.String())
	//if tx.Status == nil {
	//	panic(fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" tx.Status == nil",blockhash,blockheight))
	//}
	//if tx.TransactionIndex == nil {
	//	return nil, errors.New(fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" tx.TransactionIndex == nil",blockhash,blockheight))
	//}
	//if tx.BlockHash == nil {
	//	log.Info((fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" tx.BlockHash == nil",blockhash,blockheight)))
	//	return nil, errors.New(fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" tx.BlockHash == nil",blockhash,blockheight))
	//}
	GetRawTransaction:
	tx2,err := RpcClient.GetRawTransaction(string(tx.Hash))
	if err != nil {
		log.Info(err.Error())
		goto GetRawTransaction
	}
	if tx2 == nil {

		//log.Info((fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" GetRawTranaction == nil",blockhash,blockheight)))
		return nil, errors.New(fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" GetRawTranaction == nil",blockhash,blockheight))
	}

	tohex,_ := tx.To.ToHex()
	if RpcClient.IsUser(tohex) {
		blocktx := &dao.BlockTx{
			Txid:            string(tx.Hash),
			CoinName:        conf.Cfg.Sync.Name,
			ContractAddress: "",
			FromAddress:     tx.From.MustGetBase32Address(),
			ToAddress:       tx.To.MustGetBase32Address(),
			BlockHeight:     blockheight,
			BlockHash:       blockhash,
			Amount:          decimal.NewFromBigInt(tx.Value.ToInt(), -18).String(),
			Status:          0,
			GasPrice:        tx.GasPrice.ToInt().Int64(),
			GasUsed:         21000,
			Fee:             decimal.NewFromBigInt(tx.GasPrice.ToInt(), -18).Mul(decimal.NewFromInt(21000)).String(),
			Nonce:           tx.Nonce.ToInt().Int64(),
			Input:           tx.Data,
			Logs:            "",
			Timestamp:       time.Now(),
			CreateTime:      time.Now(),
		}
		if blocktx.Status != 0 {
			return nil, errors.New("tx.status != success")
		}
		return blocktx, nil
	} else if RpcClient.IsContract(tohex) && RpcClient.IsTransfer(tx.Data)  && watch.IsContractExist(tx.To.MustGetBase32Address()){//

		to2, amount, err := RpcClient.ParseTransferData(tx.Data)

		if err != nil {
			return nil, err
		}
		to_base32,err := cfxaddress.NewFromHex(to2,1029)
		log.Info(tx.Data,string(tx.Hash),to_base32.MustGetBase32Address())
		if err != nil {
			return nil, err
		}
		txfee := "0"
		gasused := int64(0)
		//gasUsed := int64()0
		receipt, err := RpcClient.GetTransactionReceipt(string(tx.Hash))
		if err != nil {
			return nil, err
		}
		if receipt == nil {
			//panic("txhash: "+tx.Hash+" receipt== nil")
			txScan,err :=RpcClient.GetReceiptByscan(string(tx.Hash))
			if err != nil {
				log.Info(err.Error())
				return nil,err
			}
			txfee = txScan.GasFee.Shift(-18).String()
			gasused =txScan.GasUsed.IntPart()
			//成功交易继续
			if txScan.Hash == string(tx.Hash ) && txScan.Status==0{
				if txScan.GasCoveredBySponsor{
					txfee = "0"
				}
			}else if txScan.Hash == string(tx.Hash ) && txScan.Status==1 {
				return nil,errors.New("失败交易")
			} else {
				return nil,errors.New(fmt.Sprintf("blockhash:%v,blockheight:%v,txhash: "+string(tx.Hash)+" receipt== nil;交易未获取到receipt",blockhash,blockheight))
			}
		} else {
			txfee = decimal.NewFromBigInt((*big.Int)(receipt.GasFee), -18).String()
			if receipt.GasCoveredBySponsor{
				txfee = "0"
			}
			if len(receipt.Logs) == 0 {
				return nil, errors.New("tx.status != success")
			}
			gasused = receipt.GasUsed.ToInt().Int64()
			//logstr, _ := json.Marshal(receipt.Logs)
		}

		//log.Info("txhash: "+tx.Hash+" success")

		coinname, decm, err := watch.GetContractNameAndDecimal(tx.To.MustGetBase32Address())
		if err != nil {
			return nil, err
		}

		blocktx := &dao.BlockTx{
			Txid:            string(tx.Hash),
			CoinName:        coinname,
			ContractAddress: tx.To.MustGetBase32Address(),
			FromAddress:     tx.From.MustGetBase32Address(),
			ToAddress:       to_base32.MustGetBase32Address(),
			BlockHeight:     blockheight,
			BlockHash:       blockhash,
			Amount:          decimal.NewFromBigInt(amount, 0-int32(decm)).String(),
			Status:          0,
			GasPrice:        tx.GasPrice.ToInt().Int64(),
			GasUsed:         gasused,
			Fee:             txfee,//decimal.NewFromBigInt((*big.Int)(receipt.GasFee), -18).String(),
			Nonce:           tx.Nonce.ToInt().Int64(),
			Input:           tx.Data,
			Logs:            "",
			Timestamp:       time.Now(),
			CreateTime:      time.Now(),
		}
		return blocktx, nil
	}
	if !RpcClient.IsTransfer(tx.Data){
		return nil,errors.New("调用方法!=transfer"+tohex)
	}

	if !watch.IsContractExist(tx.To.MustGetBase32Address()) {
		return nil, errors.New("合约地址不存在"+tx.To.MustGetBase32Address())
	}
	return nil, nil

}
