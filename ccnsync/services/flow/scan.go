package flow

import (
	"context"
	"errors"
	"fmt"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"solsync/common"
	"solsync/common/conf"
	"solsync/common/log"
	dao "solsync/models/po/yotta"
	"solsync/services"
	"solsync/utils"
	"sync"
	"time"
)

type Scanner struct {
	client *client.Client
	lock   *sync.Mutex
	conf   conf.SyncConfig
	Watch  *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	flowClient, err := client.New(node.Url, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return &Scanner{
		client: flowClient,
		lock:   &sync.Mutex{},
		conf:   conf.Sync,
		Watch:  watch,
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
	status, err := s.client.GetLatestBlockHeader(context.Background(), true)
	if err != nil {
		log.Infof("%+v", err.Error())
		return 0, err
	}
	return int64(status.Height), err
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
	//log.Infof("scanBlock %d ", height)
	block, err := s.client.GetBlockByHeight(context.Background(), uint64(height))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height: int64(block.Height),
			Hash:   block.ID.String(),
			//Hash:              "0x" + block.ID.String(),
			//Previousblockhash: "0x" + block.ParentID.String(),
			Previousblockhash: block.ParentID.String(),
			Timestamp:         block.Timestamp,
			Transactions:      len(block.CollectionGuarantees),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for _, tx := range block.CollectionGuarantees {
		workpool.Incr()
		go func(tx *flow.CollectionGuarantee, task *ProcTask) {
			defer workpool.Dec()
			collectionS, err := s.client.GetCollection(context.Background(), flow.HexToID(tx.CollectionID.String()))
			if err != nil {
				log.Infof("获取[%s]Collection失败%s", tx.CollectionID.String(), err.Error())
				return
			}
			for _, txId := range collectionS.TransactionIDs {
				s.batchParseTx(txId.String(), height, block.Timestamp.Unix(), block.ID.String(), task)
			}
		}(tx, task)

	}
	workpool.Wait()
	_ = starttime
	log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx string, height, blockTimestamp int64, blockHash string, task *ProcTask) {
	blockTxs, err := s.parseBlockTX(s.Watch, tx, height, blockTimestamp, blockHash)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		//log.Infof(err.Error())
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(watch *services.WatchControl, tx string, height, blockTimestamp int64, blockHash string) ([]*dao.BlockTx, error) {
	if tx == "" {
		return nil, fmt.Errorf("tx is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)
	result, err := s.client.GetTransactionResult(context.Background(), flow.HexToID(tx))
	if err != nil {
		return nil, err
	}
	if result.Status.String() != "SEALED" {
		return nil, errors.New("Status is not SEALED")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	from, to, amount, fee, err := ParseTransaction(result)
	if err != nil {
		return nil, err
	}
	dAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}

	if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
		return nil, errors.New("没有监听的地址")
	}
	blocktx := dao.BlockTx{
		CoinName:    conf.Cfg.Name,
		Txid:        tx,
		BlockHeight: height,
		BlockHash:   blockHash,
		FromAddress: from,
		ToAddress:   to,
		Amount:      dAmount.Shift(-8),
		Fee:         decimal.NewFromInt(fee).Shift(-8),
		Status:      "success",
		Timestamp:   time.Unix(blockTimestamp, 0),
	}
	blocktxs = append(blocktxs, &blocktx)
	return blocktxs, nil
}

func ParseTransaction(transaction *flow.TransactionResult) (from, to, amount string, fee int64, err error) {
	if len(transaction.Events) != 2 {
		return "", "", "", 0, fmt.Errorf("交易解析失败,code: 1")
	}
	for _, event := range transaction.Events {
		if event.Type == "A.1654653399040a61.FlowToken.TokensWithdrawn" {
			address := cadence.NewAddress(event.Value.Fields[1].ToGoValue().([8]byte))
			from = "0x" + address.Hex()
		}
		if event.Type == "A.1654653399040a61.FlowToken.TokensDeposited" {
			address := cadence.NewAddress(event.Value.Fields[1].ToGoValue().([8]byte))
			to = "0x" + address.Hex()
		}
		if from != "" && to != "" {
			decimalAmount, err := decimal.NewFromString(event.Value.Fields[0].String())
			if err != nil {
				return "", "", "", 0, errors.New("交易解析失败,code: 2")
			}
			amount = decimalAmount.Shift(8).String()
		}
	}
	if from == "" || to == "" || amount == "" {
		return "", "", "", 0, errors.New("交易解析失败,code: 3")
	}
	return
}
