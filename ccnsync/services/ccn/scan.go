package ccn

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"solsync/common"
	"solsync/common/conf"
	"solsync/common/log"
	dao "solsync/models/po/yotta"
	"solsync/services"
	rpc "solsync/utils/wtc"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	Watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
		Watch:     watch,
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

//var i = int64(60612620)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.BlockNumber() //获取到的是区块个数
	return count, err
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

func getUnCareTask(bestHeight, height int64) common.ProcTask {
	return &ProcTask{BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:     height,
			Timestamp:  time.Now(),
			Createtime: time.Now(),
		},
	}
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()
	//log.Infof("scanBlock %d ", height)

	getBlockCount := 0
	getBlockStatusCount := 0
GetBlock:
	block, err := s.BlockByNumber(height)
	if err != nil && getBlockCount < 5 {
		getBlockCount++
		time.Sleep(1 * time.Second)
		goto GetBlock
		//log.Infof("GetBlockByNumber %d  , err : %v", height, err)
		//return getUnCareTask(bestHeight, height), nil
	}
GetBlockStatus:
	status, err := s.BlockStatus(block.Blocks[0].Hash)
	if err != nil && getBlockStatusCount < 5 {
		getBlockStatusCount++
		time.Sleep(1 * time.Second)
		goto GetBlockStatus
		//log.Infof("BlockStatus %d  , err : %v", height, err)
		//return getUnCareTask(bestHeight, height), nil
	}
	if status.BlockState.IsStable != 1 {
		log.Infof("%d: IsStable不为1, 不是稳定的区块, 不处理", height)
		return getUnCareTask(bestHeight, height), nil
	}
	if status.BlockState.Type != 2 {
		log.Infof("2.不关心的交易类型, 不处理: %d", height)
		return getUnCareTask(bestHeight, height), nil
	}
	if status.BlockState.StableContent.Status != 0 {
		log.Infof("%d: 错误的的区块, 不处理", height)
		return getUnCareTask(bestHeight, height), nil
	}

	if len(block.Blocks) != 1 {
		log.Infof("GetBlockStatus %d  , err : %v", height, "block长度不为1")
		return getUnCareTask(bestHeight, height), nil
	}
	if block.Blocks[0].Type != 2 {
		log.Infof("1.不关心的交易类型, 不处理: %d", height)
		return getUnCareTask(bestHeight, height), nil
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            status.BlockState.StableContent.StableIndex,
			Hash:              block.Blocks[0].Hash,
			Previousblockhash: block.Blocks[0].Content.Previous,
			Timestamp:         time.Unix(status.BlockState.StableContent.StableTimestamp, 0),
			Transactions:      1,
			Confirmations:     12,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	s.batchParseTx(&block.Blocks[0], task)
	//workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息
	//for i, _ := range block.Transactions {
	//	workpool.Incr()
	//	receipt, err := s.TransactionByHash(block.Transactions[i])
	//	if err != nil {
	//		log.Infof("txid: %s, Get TransactionByHash err: %s",block.Transactions[i],err.Error())
	//		continue
	//	}
	//	go func(tx *rpc.Transaction, task *ProcTask) {
	//		defer workpool.Dec()
	//		s.batchParseTx(tx, task)
	//	}(receipt, task)
	//
	//}
	//workpool.Wait()
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *rpc.TxCCN, task *ProcTask) {

	blockTxs, err := parseBlockTX(s.Watch, s.RpcClient, tx, task.Block)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		//log.Infof(err.Error())
	}
}

// 解析交易
func parseBlockTX(watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.TxCCN, block *dao.BlockInfo) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)

	//if watch.IsContractExist(tx.To) || len(tx.Input) > 75 && watch.IsWatchAddressExist("0x"+tx.Input[34:74]) {
	//	//if !strings.HasPrefix(tx.Input, "0xa9059cbb") || len(tx.Input)<138{
	//	//	return nil, errors.New("非合约转账交易")
	//	//}
	//	//tmpTo := "0x" + tx.Input[34:74]
	//	//if !watch.IsWatchAddressExist(tx.From) && !watch.IsWatchAddressExist(tmpTo) {
	//	//	return nil, errors.New("没有监听的地址")
	//	//}
	//TransactionReceipt:
	//	receipt, err := rpc.TransactionReceipt(tx.Hash)
	//	if err != nil {
	//		log.Infof(err.Error())
	//		time.Sleep(time.Second * 10)
	//		goto TransactionReceipt
	//	}
	//
	//	if receipt.Status != 1 {
	//		return nil, errors.New("失败的合约转账交易")
	//	}
	//	if len(receipt.Logs) != 1 || len(receipt.Logs[0].Topics) <= 0 || receipt.Logs[0].Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
	//		return nil, errors.New("交易reciept分析错误,找管理员处理")
	//	}
	//	ContractAddress := receipt.Logs[0].Address
	//	if !watch.IsContractExist(ContractAddress) {
	//		return nil, errors.New("不是监控的合约转账")
	//	}
	//	contract, _ := watch.GetContract(ContractAddress)
	//	from := "0x" + receipt.Logs[0].Topics[1][26:66]
	//	to := "0x" + receipt.Logs[0].Topics[2][26:66]
	//	if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
	//		return nil, errors.New("没有监听的地址")
	//	}
	//	d, err := hex.DecodeString(strings.TrimPrefix(receipt.Logs[0].Data, "0x"))
	//	if err != nil {
	//		panic(tx.Hash + err.Error())
	//	}
	//	amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
	//	blocktx := dao.BlockTx{
	//		CoinName:        contract.Name,
	//		Txid:            tx.Hash,
	//		BlockHeight:     block.Height,
	//		ContractAddress: ContractAddress,
	//		FromAddress:     from,
	//		ToAddress:       to,
	//		Amount:          amount,
	//		Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
	//		BlockHash:       block.Hash,
	//		Status:          "success",
	//		Timestamp:       block.Timestamp,
	//	}
	//	blocktxs = append(blocktxs, &blocktx)
	//	return blocktxs, nil
	//}

	//普通转账处理
	if !watch.IsWatchAddressExist(tx.From) && !watch.IsWatchAddressExist(tx.Content.To) {
		return nil, errors.New("没有监听的地址")
	}
	//TransactionReceipt2:
	//	receipt, err := rpc.TransactionReceipt(tx.Hash)
	//	if err != nil {
	//		log.Infof(err.Error())
	//		time.Sleep(time.Second * 10)
	//		goto TransactionReceipt2
	//	}
	blocktx := dao.BlockTx{
		CoinName:    conf.Cfg.Name,
		Txid:        tx.Hash,
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		FromAddress: tx.From,
		ToAddress:   tx.Content.To,
		Amount:      tx.Content.Amount.Shift(-18),
		Fee: decimal.NewFromBigInt(big.NewInt(0).Mul(big.NewInt(tx.Content.Gas.IntPart()),
			big.NewInt(tx.Content.GasPrice.IntPart())), -18),
		Status:    "success",
		Timestamp: block.Timestamp,
	}
	blocktxs = append(blocktxs, &blocktx)
	return blocktxs, nil
}

// 解析交易
func parseBlockTxInternal(watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.BlockTx) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)

	status, err := rpc.BlockStatus(tx.Block.Hash)
	if err != nil {
		return nil, err
	}
	if status.BlockState.IsStable != 1 {
		return nil, fmt.Errorf("%s: IsStable不为1, 不是稳定的区块, 不处理", tx.Block.Hash)
	}
	if status.BlockState.Type != 2 {
		return nil, fmt.Errorf("%s: 2.不关心的交易类型, 不处理", tx.Block.Hash)
	}
	if status.BlockState.StableContent.Status != 0 {
		return nil, fmt.Errorf("%s: 错误的的区块, 不处理", tx.Block.Hash)
	}
	if tx.Block.Type != 2 {
		return nil, fmt.Errorf("%s: 1.不关心的交易类型, 不处理", tx.Block.Hash)
	}

	//TransactionReceipt:
	//	receipt, err := rpc.TransactionReceipt(tx.Hash)
	//	if err != nil {
	//		log.Infof(err.Error())
	//		time.Sleep(time.Second * 10)
	//		goto TransactionReceipt
	//	}
	//
	//	if receipt.Status != 1 {
	//		return nil, errors.New("失败的交易")
	//	}
	log.Infof(String(tx))
	//if tx.Value.ToInt().Cmp(big.NewInt(0)) != 0 {
	//	log.Infof(tx.To)
	if watch.IsWatchAddressExist(tx.Block.From) || watch.IsWatchAddressExist(tx.Block.Content.To) {
		blocktx := dao.BlockTx{
			CoinName:    conf.Cfg.Name,
			Txid:        tx.Block.Hash,
			BlockHeight: status.BlockState.StableContent.StableIndex,
			BlockHash:   tx.Block.Hash,
			FromAddress: tx.Block.From,
			ToAddress:   tx.Block.Content.To,
			Amount:      tx.Block.Content.Amount.Shift(-18),
			Fee: decimal.NewFromBigInt(big.NewInt(0).Mul(big.NewInt(tx.Block.Content.Gas.IntPart()),
				big.NewInt(tx.Block.Content.GasPrice.IntPart())), -18),
			Status:    "success",
			Timestamp: time.Now(),
		}
		blocktxs = append(blocktxs, &blocktx)
	} else {
		return nil, errors.New("不包含关心的地址")
	}
	//}
	//log.Infof(String(receipt.Logs))
	//for _, tlog := range receipt.Logs {
	//	if !watch.IsContractExist(tlog.Address) {
	//		log.Infof(tlog.Address)
	//		continue
	//	}
	//	if len(tlog.Topics) != 3 || tlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
	//		continue
	//	}
	//	from := "0x" + receipt.Logs[0].Topics[1][26:66]
	//	to := "0x" + receipt.Logs[0].Topics[2][26:66]
	//	//log.Infof(receipt.Logs[0].Topics[1], from, to)
	//	//log.Infof(from, to, tx.To)
	//	if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
	//		log.Infof(from, to)
	//		continue
	//	}
	//	contract, _ := watch.GetContract(tx.To)
	//	d, err := hex.DecodeString(strings.TrimPrefix(tlog.Data, "0x"))
	//	if err != nil {
	//		panic(tx.Hash + err.Error())
	//	}
	//	amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
	//	blocktx := dao.BlockTx{
	//		CoinName:        contract.Name,
	//		Txid:            tx.Hash,
	//		BlockHeight:     int64(tx.BlockNumber),
	//		ContractAddress: tlog.Address,
	//		FromAddress:     from,
	//		ToAddress:       to,
	//		Amount:          amount,
	//		Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
	//		BlockHash:       tx.BlockHash,
	//		Status:          "success",
	//		Timestamp:       time.Now(),
	//	}
	//	blocktxs = append(blocktxs, &blocktx)
	//}
	return blocktxs, nil
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
