package neo

import (
	"encoding/json"
	"fmt"
	"neotokenDataServer/common"
	"neotokenDataServer/common/conf"
	"neotokenDataServer/common/log"
	dao "neotokenDataServer/models/po/neo"
	rpc "neotokenDataServer/utils/neo"
	"strconv"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.VinRollBack(height)
	dao.VoutRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(6113472)

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	return s.GetBlockCount()
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
	//starttime := time.Now()
retryGetBlock:
	block, err := s.GetBlockByHeight2(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlock
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	if has, err := dao.BlockHashExist(block.Hash); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height > s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.Hash,
			Previousblockhash: block.PreviousBlockHash,
			Nextblockhash:     block.NextBlockHash,
			Transactions:      len(block.Txs),
			Confirmations:     bestHeight - height, //block.Confirmations
			Time:              time.Unix(block.Time, 0),
		},
	}

	//处理区块内的交易
	if len(block.Txs) > 0 {
		for _, tx := range block.Txs {
			//过滤旷工交易
			if tx.Type != "InvocationTransaction" {
				continue
			}

			if txInfo, err := s.parseBlockRawTX(tx, block.Hash, height); err != nil {
				//log.Info(tx.Txid, "parseBlockRawTX err：", err.Error())
			} else {
				task.TxInfos = append(task.TxInfos, txInfo)
			}
		}
	}
	//tsjon, _ := json.Marshal(task)
	//log.Info(string(tsjon))
	//log.Infof("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(block.Txs), time.Since(starttime).Seconds())
	return task, nil
}

//解析交易
func (s *Scanner) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) (txInfo *TxInfo, err error) {

	if tx == nil || tx.Type != "InvocationTransaction" {
		return nil, fmt.Errorf("txid is null")
	}
	blockTx := &dao.BlockTx{
		Txid:      tx.Txid,
		Height:    height,
		Hash:      blockhash,
		Vincount:  len(tx.Vin),
		Voutcount: len(tx.Vout),
		Type:      tx.Type,
		Vmstate:   "HALT",
	}
	txInfo = &TxInfo{Tx: blockTx}
	//获取合约执行状态
getlog:
	txlog, err := s.RpcClient.GetTransactionLog(tx.Txid)
	if err != nil {
		log.Warn(err.Error() + tx.Txid)
		time.Sleep(time.Second * 3)
		goto getlog
	}
	if len(txlog.Executions) == 0 || txlog.Executions[0].Vmstate != "HALT" {
		return nil, fmt.Errorf("ship tx:%v", txlog.Executions[0].Vmstate)
	}
	for index, nt := range txlog.Executions[0].Notifications {
		if nt.State.Type != "Array" {
			continue
		}
		valueJson, _ := json.Marshal(nt.State.Value)
		values := make([]*rpc.Param, 0)
		if err := json.Unmarshal(valueJson, &values); err != nil {
			log.Warn(string(valueJson), err.Error())
			continue
		}

		if len(values) != 4 || values[0].Type != "ByteArray" || values[0].Value != "7472616e73666572" {
			continue
		}
		var amount int64
		if values[3].Type == "Integer" {
			if amount, err = strconv.ParseInt(values[3].Value.(string), 10, 64); err != nil {
				return nil, err
			}
		} else if values[3].Type == "ByteArray" {
			if amount, err = bytesToInt(values[3].Value.(string)); err != nil {
				log.Warn(values[0].Value.(string) + " " + values[3].Value.(string) + "  " + err.Error())
				return nil, err
			}
		} else {
			panic(tx.Txid)
			return nil, fmt.Errorf("Unknow type:%v", values[3].Type)
		}
		contractx := &dao.ContractTx{
			Txid:     tx.Txid,
			Height:   height,
			Hash:     blockhash,
			Contract: nt.Contract,
			Vmstate:  txlog.Executions[0].Vmstate,
			Index:    index,
			From:     BytesToNeoAddr(values[1].Value.(string)),
			To:       BytesToNeoAddr(values[2].Value.(string)),
			Value:    amount,
		}
		txInfo.Contractxs = append(txInfo.Contractxs, contractx)
	}

	return txInfo, nil
}
