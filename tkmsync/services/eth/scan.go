package eth

import (
	"encoding/json"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"math/big"
	"rsksync/common"
	"rsksync/common/log"
	"rsksync/conf"
	dao "rsksync/models/po/eth"
	"rsksync/utils"
	"rsksync/utils/eth"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	*eth.RpcClient
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	//如果启动eth，顺便启动定制加载的合约
	err := InitEthClient(node.Url)
	if err != nil {
		panic(err)
	}
	return &Scanner{
		RpcClient: eth.NewRpcClient(node.Url),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)

	dao.DeleteBlockTX(height)
}

//爬数据
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.BlockNumber()
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.GetMaxBlockIndex()
}

//批量扫描多个区块
func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	starttime := time.Now()
	count := endHeight - startHeight
	taskmap := &sync.Map{}
	wg := &sync.WaitGroup{}

	wg.Add(int(count))
	for i := int64(0); i < count; i++ {
		height := startHeight + i
		go func(w *sync.WaitGroup) {
			if task, err := s.ScanIrreverseBlock(height, bestHeight); err == nil {
				taskmap.Store(height, task)
			}
			w.Done()
		}(wg)
	}
	wg.Wait()
	log.Debugf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return taskmap
}

func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByNumber(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &EthProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Number,
			Hash:           block.Hash,
			FrontBlockHash: block.ParentHash,
			Timestamp:      time.Unix(block.Timestamp, 0),
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
		txInfos: make([][]*dao.BlockTX, 0),
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tmp := range block.Transactions {

				tx := tmp
				go s.batchParseTx(&tx, bestHeight, block.Timestamp, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Transactions {
				blockTx, err := s.parseBlockTX(&tx, bestHeight, block.Timestamp)
				if err == nil {
					task.txInfos = append(task.txInfos, blockTx)
				}
			}
		}
	}
	log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *eth.Transaction, bestHeight, blockTimestamp int64, task *EthProcTask, w *sync.WaitGroup) {
	defer w.Done()
	blockTx, err := s.parseBlockTX(tx, bestHeight, blockTimestamp)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *eth.Transaction, bestHeight, blockTimestamp int64) ([]*dao.BlockTX, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	res := make([]*dao.BlockTX, 0)
	if !s.IsContractTx(tx) {

		//主币transfer的处理逻辑
		txReceipt, err := s.GetTransactionReceipt(tx.Hash)
		if err != nil {
			return nil, err
		}
		if txReceipt.Status != "0x1" {
			log.Infof("%s 无效交易", txReceipt.TransactionHash)
			return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
		}

		blocktx := &dao.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasUsed:     tx.Gas,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     eth.WEI,
			Timestamp:   time.Unix(blockTimestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
		}
		res = append(res, blocktx)
	} else {
		//toAddr, amt, err := eth.ERC20{}.ParseTransferData(tx.Input)
		//"sta" "asko"，"dego"，临时处理 不使用这种方式入账
		//"sta" "asko"，"dego"，"bonk"，"mq","pis"临时处理 不使用这种方式入账
		//if err == nil  && (
		//	tx.To != "0xa7de087329bfcda5639247f96140f9dabe3deed1" ||
		//	tx.To != "0xeeee2a622330e6d2036691e983dee87330588603"){
		if strings.ToLower(tx.To) != "0xa7de087329bfcda5639247f96140f9dabe3deed1" &&
			strings.ToLower(tx.To) != "0xeeee2a622330e6d2036691e983dee87330588603" &&
			strings.ToLower(tx.To) != "0x88ef27e69108b2633f8e1c184cc37940a075cc02" &&
			strings.ToLower(tx.To) != "0x6d6506e6f438ede269877a0a720026559110b7d5" &&
			strings.ToLower(tx.To) != "0xc26d79f8dcb5bbbc52e4cf17c6fb800a9aad39a7" &&
			strings.ToLower(tx.To) != "0x834ce7ad163ab3be0c5fd4e0a81e67ac8f51e00c" {
			//erc2o tx
			//blocktx := &dao.BlockTX{
			//	BlockHeight:     tx.BlockNumber,
			//	BlockHash:       tx.BlockHash,
			//	Txid:            tx.Hash,
			//	FromAddress:     tx.From,
			//	Nonce:           tx.Nonce,
			//	GasUsed:         tx.Gas,
			//	GasPrice:        tx.GasPrice.Int64(),
			//	Input:           tx.Input,
			//	CoinName:        s.conf.Name,
			//	Decimal:         eth.WEI,
			//	Timestamp:       time.Unix(blockTimestamp, 0),
			//	Amount:          decimal.NewFromBigInt(amt, 0),
			//	ToAddress:       toAddr,
			//	ContractAddress: tx.To,
			//}
			//res = append(res, blocktx)
			if errPTR := s.parseTxReceipt(tx, blockTimestamp, &res); errPTR != nil {
				return nil, errPTR
			}

		} else if strings.ToLower(tx.To) == "0x1c2349acbb7f83d07577692c75b6d7654899bf10" {
			//todo 由于没有详细查看该程序逻辑，临时先使用硬编码解析mykey eth合约的代码
			txReceipt, err := s.GetTransactionReceipt(tx.Hash)
			if err != nil {
				log.Errorf("%s 交易获取异常", txReceipt.TransactionHash)
				return nil, err
			}
			if txReceipt.Status != "0x1" {
				log.Infof("%s 无效交易", txReceipt.TransactionHash)
				return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
			}
			//log.Info("txReceipt.Removed:",txReceipt.Removed)
			//if txReceipt.Removed {
			//	log.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
			//	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
			//}
			log.Infof("mykey eth合约入账解析: %s，Logs：%d", tx.Hash, len(txReceipt.Logs))
			for _, lg := range txReceipt.Logs {
				if lg.Address != "0x1c2349acbb7f83d07577692c75b6d7654899bf10" {
					continue
				}
				if len(lg.Topics) == 0 || lg.Topics[0] != "0x3efc190d59645f005a5974aa84aa94401ad997938870e7b2aa74a45138ad679b" {
					continue
				}
				blocktx := &dao.BlockTX{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					Nonce:           tx.Nonce,
					GasUsed:         txReceipt.GasUsed,
					GasPrice:        tx.GasPrice.Int64(),
					Input:           tx.Input,
					CoinName:        s.conf.Name,
					Decimal:         eth.WEI,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: lg.Address,
				}

				haTopicsHash := make([]common2.Hash, 0)
				for _, vt := range lg.Topics {
					haTopicsHash = append(haTopicsHash, common2.HexToHash(vt))
				}
				vLog := types.Log{
					Address:     common2.HexToAddress(lg.Address),
					Topics:      haTopicsHash,
					Data:        common2.FromHex(lg.Data),
					BlockNumber: uint64(tx.BlockNumber),
					TxHash:      common2.HexToHash(txReceipt.TransactionHash),
					TxIndex:     uint(txReceipt.TransactionIndex),
					BlockHash:   common2.HexToHash(txReceipt.BlockHash),
					Index:       uint(txReceipt.LogIndex),
					Removed:     txReceipt.Removed,
				}
				sender, receiver, amountFloatStr, txid, err := MyKeyProcessTransferLogic(vLog)
				log.Infof("mykey 解析数据解析：sender：%s,receiver:%s,am:%s", sender, receiver, amountFloatStr)
				if err != nil {
					log.Infof("mykey 解析异常 err:%s", err.Error())
					continue
				}
				if txid != tx.Hash {
					log.Infof("mykey txid 不一致")
					continue
				}

				am, _ := decimal.NewFromString(amountFloatStr)
				if sender == "" || receiver == "" || am.IsZero() {
					log.Infof("mykey txid 解析数据解析不全：sender：%s,receiver:%s,am:%s", sender, receiver, am.String())
					continue
				}
				blocktx.Amount = am.Shift(18) //扩大18变成int
				blocktx.FromAddress = sender
				blocktx.ToAddress = receiver
				blocktx.ContractAddress = "" //合约清空
				res = append(res, blocktx)
				log.Infof("添加mykey eth交易，txid：%s", txid)
			}
		} else {
			if errPTR := s.parseTxReceipt(tx, blockTimestamp, &res); errPTR != nil {
				return nil, errPTR
			}

			////contract tx
			////log.Infof("sta ,asko dego:%s",tx.Hash)
			//txReceipt, err := s.GetTransactionReceipt(tx.Hash)
			//if err != nil {
			//	return nil, err
			//}
			//if txReceipt.Status != "0x1" {
			//	log.Infof("%s 无效交易", txReceipt.TransactionHash)
			//	return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
			//}
			////if txReceipt.Removed {
			////	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
			////}
			//for _, lg := range txReceipt.Logs {
			//	blocktx := &dao.BlockTX{
			//		BlockHeight:     tx.BlockNumber,
			//		BlockHash:       tx.BlockHash,
			//		Txid:            tx.Hash,
			//		Nonce:           tx.Nonce,
			//		GasUsed:         txReceipt.GasUsed,
			//		GasPrice:        tx.GasPrice.Int64(),
			//		Input:           tx.Input,
			//		CoinName:        s.conf.Name,
			//		Decimal:         eth.WEI,
			//		Timestamp:       time.Unix(blockTimestamp, 0),
			//		ContractAddress: lg.Address,
			//	}
			//	if lg.Data == "" || len(lg.Data) < 3 {
			//		continue
			//	}
			//	tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
			//	blocktx.Amount = decimal.NewFromBigInt(tmp, 0)
			//	if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			//		continue
			//	}
			//	if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			//		blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
			//		blocktx.ToAddress = "0x" + lg.Topics[2][26:66]
			//	} else {
			//		continue
			//	}
			//	res = append(res, blocktx)
			//}
		}
	}
	return res, nil
}

func (s *Scanner) parseTxReceipt(tx *eth.Transaction, blockTimestamp int64, res *[]*dao.BlockTX) error {
	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return err
	}
	if txReceipt.Status != "0x1" {
		log.Infof("%s 无效交易", txReceipt.TransactionHash)
		return fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
	}
	//if txReceipt.Removed {
	//	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
	//}
	for i, lg := range txReceipt.Logs {
		blockTx := &dao.BlockTX{
			BlockHeight:     tx.BlockNumber,
			BlockHash:       tx.BlockHash,
			Txid:            tx.Hash,
			Nonce:           tx.Nonce,
			GasUsed:         txReceipt.GasUsed,
			GasPrice:        tx.GasPrice.Int64(),
			Input:           tx.Input,
			CoinName:        s.conf.Name,
			Decimal:         eth.WEI,
			Timestamp:       time.Unix(blockTimestamp, 0),
			ContractAddress: lg.Address,
			CreateTime:      time.Now(),
		}
		sta, staErr := utils.ParseInt(txReceipt.Status)
		if staErr != nil {
			log.Errorf("tx Log[%d] status parse err:%s", i, staErr.Error())
		}
		blockTx.Status = sta

		// 保存Logs对应的索引数据
		btys, jmErr := json.Marshal(txReceipt.Logs[i])
		if jmErr != nil {
			log.Errorf("tx Log[%d] json marshal err:%s", i, jmErr.Error())
		} else {
			blockTx.Logs = string(btys)
		}

		//没有输出日志数据，认为是非合法的交易
		//status=2 表示失败
		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			blockTx.Status = 2
		}

		if lg.Data == "" || len(lg.Data) < 3 {
			continue
		}
		tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
		blockTx.Amount = decimal.NewFromBigInt(tmp, 0)
		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			continue
		}
		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			blockTx.FromAddress = "0x" + lg.Topics[1][26:66]
			blockTx.ToAddress = "0x" + lg.Topics[2][26:66]
		} else {
			continue
		}
		*res = append(*res, blockTx)
	}
	return nil
}
