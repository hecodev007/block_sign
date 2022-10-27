package brise

import (
	"domsync/common"
	"domsync/common/conf"
	dao "domsync/models/po/brise"
	"domsync/services"
	"domsync/utils"
	"domsync/utils/brise"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"sync"
	"time"
)

type Scanner struct {
	client *brise.RpcClient
	lock   *sync.Mutex
	conf   conf.SyncConfig
	Watch  *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	//如果启动eth，顺便启动定制加载的合约
	err := InitEthClient(node.Url)
	if err != nil {
		panic(err)
	}
	return &Scanner{
		client: brise.NewRpcClient(node.Url),
		lock:   &sync.Mutex{},
		conf:   conf.Sync,
		Watch:  watch,
	}
	//flowClient, err := client.New(node.Url, grpc.WithInsecure())
	//if err != nil {
	//	panic(err)
	//}
	//
	//return &Scanner{
	//	client: flowClient,
	//	lock:   &sync.Mutex{},
	//	conf:   conf.Sync,
	//	Watch:  watch,
	//}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	_, err := dao.BlockRollBack(height)
	if err != nil {
		panic(err.Error())
	}
	_, err = dao.TxRollBack(height)
	if err != nil {
		panic(err.Error())
	}
}

func (s *Scanner) Init() error {
	if s.conf.EnableRollback {
		s.Rollback(s.conf.RollHeight)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(60612620)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.client.BlockNumber()
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
	starttime := time.Now()
	//log.Printf("scanBlock %d ", height)
	block, err := s.client.GetBlockByNumber(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			//Height: int64(block.Height),
			//Hash:   block.ID.String(),
			////Hash:              "0x" + block.ID.String(),
			////Previousblockhash: "0x" + block.ParentID.String(),
			//Previousblockhash: block.ParentID.String(),
			//Timestamp:         block.Timestamp,
			//Transactions:      len(block.CollectionGuarantees),
			//Confirmations:     bestHeight - height + 1,
			//Createtime:        time.Now(),
			Height:            block.Number,
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Timestamp:         time.Unix(block.Timestamp, 0),
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息
	//for _, tx := range block.Transactions {
	//	workpool.Incr()
	//	go func(tx *brise.Transaction, task *ProcTask) {
	//		defer workpool.Dec()
	//		blockTx, err := s.parseBlockTX(tx, bestHeight, block.Timestamp)
	//		if err == nil {
	//			task.TxInfos = append(task.TxInfos, blockTx...)
	//		}
	//	}(&tx, task)
	//
	//}
	//workpool.Wait()
	workpool := utils.NewWorkPool(10)
	for _, tx := range block.Transactions {
		workpool.Incr()
		go func(tx brise.Transaction, task *ProcTask) {
			defer workpool.Dec()
			blockTx, err := s.parseBlockTX(tx, bestHeight, block.Timestamp)
			if err == nil {
				task.TxInfos = append(task.TxInfos, blockTx...)
			}
		}(tx, task)
	}
	workpool.Wait()
	_ = starttime
	log.Printf("scanBlock %d ,%d,used time : %f 's", height, len(block.Transactions), time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx brise.Transaction, height, blockTimestamp int64, blockHash string, task *ProcTask) {
	blockTxs, err := s.parseBlockTX(tx, height, blockTimestamp)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		//log.Printf(err.Error())
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx brise.Transaction, bestHeight, blockTimestamp int64) ([]*dao.BlockTx, error) {
	log.Printf("%d,%s ", tx.BlockNumber, tx.Hash)
	//if tx.Hash != "0x985485b1f74530bb5e6fd6b729e5168e80015225b5ebacdb209bed3367c595d1" {
	//	return nil, fmt.Errorf("不是关注的地址")
	//}
	res := make([]*dao.BlockTx, 0)
	txReceipt, err := s.client.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return nil, err
	}
	if txReceipt.Status != "0x1" {
		log.Printf("%s 无效交易", txReceipt.TransactionHash)
		return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
	}

	if !s.client.IsContractTx(&tx) {
		log.Printf("普通交易：%s", tx.Hash)
		blocktx := &dao.BlockTx{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasUsed:     txReceipt.GasUsed,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     brise.WEI,
			Timestamp:   time.Unix(blockTimestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
		}
		res = append(res, blocktx)
	} else { //合约交易，通过
		log.Printf("合约交易：%s", tx.Hash)
		if errPTR := s.parseTxReceipt(tx, blockTimestamp, &res, txReceipt); errPTR != nil {
			return nil, errPTR
		}
	}
	return res, nil
}

func (s *Scanner) parseTxReceipt(tx brise.Transaction, blockTimestamp int64, res *[]*dao.BlockTx, txReceipt *brise.TransactionReceipt) error {
	for i, lg := range txReceipt.Logs {
		contractInfo, err := s.Watch.GetContract(lg.Address)
		if err != nil { //不是关注的合约地址则跳过
			//log.Printf("dont't have care of watch contract : %s \n", lg.Address)
			continue
		}
		blockTx := &dao.BlockTx{
			BlockHeight:     tx.BlockNumber,
			BlockHash:       tx.BlockHash,
			Txid:            tx.Hash,
			Nonce:           tx.Nonce,
			GasUsed:         txReceipt.GasUsed,
			GasPrice:        tx.GasPrice.Int64(),
			Input:           tx.Input,
			Timestamp:       time.Unix(blockTimestamp, 0),
			ContractAddress: lg.Address,
			CreateTime:      time.Now(),
		}
		if lg.Removed { //log被删除
			continue
		}
		sta, staErr := utils.ParseInt(txReceipt.Status)
		if staErr != nil {
			log.Printf("tx Log[%d] status parse err:%s", i, staErr.Error())
		}
		blockTx.Status = sta
		blockTx.ContractAddress = lg.Address
		blockTx.Decimal = contractInfo.Decimal
		blockTx.CoinName = contractInfo.Name
		// 保存Logs对应的索引数据
		btys, jmErr := json.Marshal(txReceipt.Logs[i])
		if jmErr != nil {
			log.Printf("tx Log[%d] json marshal err:%s", i, jmErr.Error())
		} else {
			blockTx.Logs = string(btys)
		}
		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			continue
		}
		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			blockTx.Status = 2 //status=2 表示失败
		}
		if lg.Data == "" || len(lg.Data) < 3 {
			continue
		}
		tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
		blockTx.Amount = decimal.NewFromBigInt(tmp, 0)
		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" && lg.Topics[1] != "0x0000000000000000000000000000000000000000000000000000000000000000" && blockTx.Amount.String() != "0" {
			//转账的交易,0的地址算不算上去，还有转账的是aur还是tri币，这个怎么看，这些地址都是合约地址啊---合约地址的进度
			blockTx.FromAddress = "0x" + lg.Topics[1][26:66]
			blockTx.ToAddress = "0x" + lg.Topics[2][26:66]
			blockTx.ContractAddress = lg.Address //合约地址
			log.Printf("%s,%s,%s,%s", lg.Data, blockTx.Amount.Shift(int32(-contractInfo.Decimal)).String(), blockTx.FromAddress, blockTx.ToAddress)
		} else {
			continue
		}
		*res = append(*res, blockTx)
	}
	return nil
}
