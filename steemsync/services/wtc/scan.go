package wtc

import (
	"fmt"
	"steemsync/common"
	"steemsync/common/conf"
	"steemsync/common/log"
	dao "steemsync/models/po/yotta"
	"steemsync/services"
	"steemsync/utils"
	"steemsync/utils/rpc"
	"steemsync/utils/rpc/transports/rpcclient"
	"steemsync/utils/rpc/types"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.Client
	lock  *sync.Mutex
	conf  conf.SyncConfig
	Watch *services.WatchControl
	//IrreverseBlock map[int64]common.ProcTask
	TaskCaches sync.Map
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	t := rpcclient.NewRpcClient(node.Url)
	clt, err := rpc.NewClient(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &Scanner{
		Client:     clt,
		lock:       &sync.Mutex{},
		conf:       conf.Sync,
		Watch:      watch,
		TaskCaches: sync.Map{},
	}
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

//var i = int64(14717020)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	props, err := s.Client.Database.GetDynamicGlobalProperties()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return int64(props.LastIrreversibleBlockNum), err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (t common.ProcTask, err error) {
	//task, ok := s.IrreverseBlock[height]
	taskBest, ok := s.TaskCaches.Load(bestHeight)
	if !ok {
		taskBest, err = s.scanBlock(bestHeight, bestHeight)
		if err != nil {
			return nil, err
		}
		s.TaskCaches.Store(bestHeight, taskBest)
	}
	for h := bestHeight - 1; h >= height; h-- {
		//		log.Info(h)
		taskH, ok := s.TaskCaches.Load(h)
		if ok { //如果有block缓存且hash一致则退出,否者扫节点
			taskHer, ok := s.TaskCaches.Load(h + 1)
			if !ok {
				panic("")
			}
			if taskHer.(common.ProcTask).ParentHash() == taskH.(common.ProcTask).GetBlockHash() {
				break
			}
		}

		taskHnew, err := s.scanBlock(h, bestHeight)
		if err != nil {
			return nil, err
		}

		if ok {
			log.Info(h, "分叉恢复", taskH.(common.ProcTask).GetBlockHash(), taskHnew.GetBlockHash())
		}
		s.TaskCaches.Store(h, taskHnew)
	}
	taskInter, ok := s.TaskCaches.Load(height)
	if !ok {
		panic("")
	}
	taskInter.(common.ProcTask).SetBestHeight(bestHeight, false)
	return taskInter.(common.ProcTask), nil
	//taskInter, ok := s.TaskCaches.Load(height)
	//if ok {
	//	task := taskInter.(common.ProcTask)
	//	block, err := s.GetBlockByNumber(height, false)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if block.Hash == task.GetBlockHash() {
	//		task.SetBestHeight(bestHeight, false)
	//		return task, nil
	//	}
	//}
	//task, err := s.scanBlock(height, bestHeight)
	//if err != nil {
	//	return task, err
	//}
	//s.TaskCaches.Store(height, task)
	//for i := height - 150; i < height-100; i++ {
	//	s.TaskCaches.Delete(i)
	//}
	//return task, nil
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	taskInter, ok := s.TaskCaches.Load(height)
	if ok {
		s.TaskCaches.Delete(height)
		task := taskInter.(common.ProcTask)
		block, err := s.Client.Database.GetBlock(uint32(height))
		if err != nil {
			return nil, err
		}
		if block.TransactionMerkleRoot == task.GetBlockHash() {
			task.SetBestHeight(bestHeight, true)
			return task, nil
		}
	}
	task, err := s.scanBlock(height, bestHeight)
	return task, err
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	st := time.Now()
	//log.Infof("scanBlock %d ", height)
getblocknumber:
	block, err := s.Client.Database.GetBlock(uint32(height))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	if block.Number == 0 {
		goto getblocknumber
	}
	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height: int64(block.Number),
			Hash:   block.TransactionMerkleRoot,
			//	Previousblockhash: block.ParentHash,
			Timestamp:     *block.Timestamp.Time,
			Transactions:  len(block.Transactions),
			Confirmations: bestHeight - height + 1,
			Createtime:    time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.Irreversible = true
	}
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息
	operations, err := s.Client.Database.GetOpsInBlock(uint32(height), false)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	for _, operation := range operations {

		switch op := operation.Data().(type) {
		case *types.TransferOperation:
			if op.From == "steemgoapi" || op.To == "steemgoapi" {
				fmt.Printf("Transfer from %v,to %v,memo %v,amount %v\n", op.From, op.To, op.Memo, op.Amount)
			}

		}
		workpool.Incr()
		go func(tx *types.OperationObject, task *ProcTask) {

			defer workpool.Dec()
			s.batchParseTx(tx, task)

		}(operation, task)

	}
	workpool.Wait()
	_ = st
	//log.Infof("scanBlock %v ,耗时 : %v", height, time.Since(st))
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *types.OperationObject, task *ProcTask) {
	if _, ok := tx.Operation.Data().(*types.TransferOperation); !ok {
		return
	}
	blockTxs, err := s.parseBlockTX(tx, task.Block)
	if blockTxs != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	}
	if err != nil {
		//log.Info(tx.Hash, err.Error())
		//panic("")
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *types.OperationObject, block *dao.BlockInfo) (blocktxs []*dao.BlockTx, err error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	//Watch := s.Watch
	//rpc := s.RpcClient
	//var blocktxs = make([]*dao.BlockTx, 0)
	operation := tx.Operation.Data().(*types.TransferOperation)
	amounts := strings.Split(operation.Amount, "")
	amount, _ := decimal.NewFromString(amounts[0])
	//eth 转账交易
	if amount.GreaterThan(decimal.Zero) && (s.Watch.IsWatchAddressExist(operation.From) || s.Watch.IsWatchAddressExist(operation.To)) {
		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.TransactionID,
			BlockHeight:     block.Height,
			ContractAddress: "",
			FromAddress:     operation.From,
			ToAddress:       operation.To,
			Amount:          amount,
			//Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
			BlockHash: block.Hash,
			Status:    "success",
			Timestamp: block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}

	return
}

//
//func (s *Scanner) IsContractTx(tx *rpc.Transaction) bool {
//	from := tx.From
//	to := tx.To
//	input := tx.Input
//
//	hoo_from := s.Watch.IsWatchAddressExist(from)
//	hoo_contract := s.Watch.IsContractExist(to)
//	hoo_to := s.Watch.IsWatchAddressExist(to)
//
//	if hoo_from && len(input) < 10 {
//		return false
//	}
//
//	if hoo_to {
//		return false
//	}
//
//	if hoo_contract {
//		return true
//	}
//	return false
//}
//
//// 解析交易
//func parseBlockTxInternal(Watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.Transaction) ([]*dao.BlockTx, error) {
//
//	if tx == nil {
//		return nil, fmt.Errorf("tx is null")
//	}
//	var blocktxs = make([]*dao.BlockTx, 0)
//TransactionReceipt:
//	receipt, err := rpc.TransactionReceipt(tx.Hash)
//	if err != nil {
//		log.Error(err.Error())
//		time.Sleep(time.Second * 10)
//		goto TransactionReceipt
//	}
//
//	if receipt.Status != 1 {
//		return nil, errors.New("失败的交易")
//	}
//	log.Info(String(tx))
//	if tx.Value.ToInt().Cmp(big.NewInt(0)) != 0 {
//		log.Info(tx.To)
//		if Watch.IsWatchAddressExist(tx.From) || Watch.IsWatchAddressExist(tx.To) {
//			blocktx := dao.BlockTx{
//				CoinName:    conf.Cfg.Sync.CoinName,
//				Txid:        tx.Hash,
//				BlockHeight: int64(tx.BlockNumber),
//				BlockHash:   tx.BlockHash,
//				FromAddress: tx.From,
//				ToAddress:   tx.To,
//				Amount:      decimal.NewFromBigInt(tx.Value.ToInt(), -18),
//				Fee:         decimal.NewFromBigInt(big.NewInt(0).Mul(receipt.GasUsed.ToInt(), tx.GasPrice.ToInt()), -18),
//				Status:      "success",
//				Timestamp:   time.Now(),
//			}
//			blocktxs = append(blocktxs, &blocktx)
//		}
//	}
//	log.Info(String(receipt.Logs))
//	for _, tlog := range receipt.Logs {
//		if !Watch.IsContractExist(tlog.Address) {
//			log.Info(tlog.Address)
//			continue
//		}
//		if len(tlog.Topics) != 3 || tlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
//			continue
//		}
//		from := "0x" + receipt.Logs[0].Topics[1][26:66]
//		to := "0x" + receipt.Logs[0].Topics[2][26:66]
//		//log.Info(receipt.Logs[0].Topics[1], from, to)
//		//log.Info(from, to, tx.To)
//		if !Watch.IsWatchAddressExist(from) && !Watch.IsWatchAddressExist(to) {
//			log.Info(from, to)
//			continue
//		}
//		contract, _ := Watch.GetContract(tx.To)
//		d, err := hex.DecodeString(strings.TrimPrefix(tlog.Data, "0x"))
//		if err != nil {
//			panic(tx.Hash + err.Error())
//		}
//		amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
//		blocktx := dao.BlockTx{
//			CoinName:        contract.Name,
//			Txid:            tx.Hash,
//			BlockHeight:     int64(tx.BlockNumber),
//			ContractAddress: tlog.Address,
//			FromAddress:     from,
//			ToAddress:       to,
//			Amount:          amount,
//			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
//			BlockHash:       tx.BlockHash,
//			Status:          "success",
//			Timestamp:       time.Now(),
//		}
//		blocktxs = append(blocktxs, &blocktx)
//	}
//	return blocktxs, nil
//}
