package xtz

import (
	"errors"
	"fmt"
	"github.com/goat-systems/go-tezos/v4/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/shopspring/decimal"
	"os"
	"strings"
	"sync"
)

/*
模版
*/

const (
	MainNetChainId = "NetXdQprcVkpaWU"
)

type XtzService struct {
	cfg          conf.Config
	nodeCfg      conf.NodeConfig
	latestHeight int64
	url          string
	lock         sync.RWMutex
	// add custom filed
	client *rpc.Client
}

func (ts *XtzService) GetLatestBlockHeight() (int64, error) {
	var err error
	if ts.client == nil {
		ts.client, err = rpc.New(ts.url)
		if err != nil {
			log.Errorf("init xtz client error: %v", err)
			os.Exit(1)
		}
	}
	resp, head, err := ts.client.Header(&rpc.BlockIDHead{})
	if err != nil {
		return 0, err
	}
	if resp.IsSuccess() {
		ts.lock.Lock()
		ts.latestHeight = int64(head.Level)
		defer ts.lock.Unlock()
		return int64(head.Level), nil
	}
	return 0, fmt.Errorf("xtz rpc get head error: %v,status=%s", resp.Error(), resp.Status())
}

func (ts *XtzService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	blockId := rpc.BlockIDLevel(int(height))
	resp, block, err := ts.client.Block(&blockId)
	if err != nil {
		return nil, fmt.Errorf("xtz rpc get block error: %v", err)
	}
	if resp.IsError() || block == nil {
		return nil, fmt.Errorf("xtz rpc get block error: %v,status=%s,height: %d",
			resp.Error(), resp.Status(), height)
	}
	bd := new(common.BlockData)
	bd.Height = int64(block.Header.Level)
	bd.Hash = block.Hash
	bd.PrevHash = block.Header.Predecessor
	bd.Timestamp = block.Header.Timestamp.Unix()
	bd.Confirmation = ts.latestHeight - height
	txDatas, err := ts.parseBlock(block)
	if err != nil {
		return nil, fmt.Errorf("%d parse transaction error: %v", height, err)
	}
	bd.TxDatas = txDatas
	bd.TxNums = len(txDatas)
	return bd, nil
}

func (ts *XtzService) parseBlock(block *rpc.Block) ([]*common.TxData, error) {
	var txDatas []*common.TxData
	for _, txs := range block.Operations {
		if len(txs) > 0 {
			for _, tx := range txs {
				//1. 判断是否是主网id
				if tx.ChainID != MainNetChainId {
					log.Errorf("这笔交易不是主网的chainId,请检查主网id是否已经更换："+
						"localChainId：%s ,failChainId: %s", MainNetChainId, tx.ChainID)
					continue
				}
				txid := tx.Hash
				for _, content := range tx.Contents {
					if content.Kind == rpc.TRANSACTION {
						from := content.Source
						to := content.Destination
						amount := content.Amount
						fee, _ := decimal.NewFromString(content.Fee)
						//todo parse metadata
						change, isFakeTx, err := ts.parseMetadata(content.Metadata, from, amount)
						if err != nil {
							log.Errorf("parse tx has error: %v", err)
							continue //避免后续的交易没有被解析
						}
						// 在手续费中加上分配费
						if change != "" {
							cd, _ := decimal.NewFromString(change)
							if cd.GreaterThanOrEqual(decimal.NewFromFloat(0.1).Shift(common.XTZDecimal)) {
								// 目前观察分配费用都是 0.06425 这个值一般来说不太会改动，所以这里做一个判断，如果分配费用大于0.1,这里就要检查一下了
								log.Errorf("==============>: 分配费大于等于0.1,请检查这笔交易：TXID=[%s]", txid)
								continue
							}
							fee = fee.Add(cd)
						}
						if isFakeTx {
							// 如果是假充值，需要把amount设置为 0 ，避免推送上去
							amount = "0"
						}

						//构建 txData
						td := new(common.TxData)
						td.MainDecimal = common.XTZDecimal
						td.Height = int64(block.Header.Level)
						td.Txid = txid
						td.IsFakeTx = isFakeTx
						td.FromAddr = from
						td.ToAddr = to
						td.Amount = amount
						td.Fee = fee.String()
						txDatas = append(txDatas, td)
					}
				}
			}
		}
	}
	return txDatas, nil
}

/*
交易的metadata一般只有balance_updates和operation_result这两个，所以目前只处理这两个，如果后续有需求，在解析其他的
解析metadata最主要是为了获取分配费，这个分配费我也不太清楚，有可能是往新地址转账需要分配费。这里我把分配费合并在手续费中
因为这个钱其实都是从from地址出的，而且都是给矿工的
*/
func (ts *XtzService) parseMetadata(metadata *rpc.ContentsMetadata, from, amount string) (string, bool, error) {
	if metadata == nil {
		// todo 返回错误，一般交易都会有metadata
		return "", false, errors.New("metadata is nil")
	}
	//这里判断是否是假充值
	if metadata.OperationResults.Status == "applied" {
		/*
			这里为什么等于3了？
			当amount=0时，BalanceUpdates这里其实是等于空的
			当amount>0时，这里必定会发生余额更新的操作，所以这里会有两笔
			当如果有一笔分配的时候，这里就会从from地址在发生一笔余额变化
		*/
		if len(metadata.OperationResults.BalanceUpdates) == 3 {
			for _, bu := range metadata.OperationResults.BalanceUpdates {
				if bu.Kind == "contract" && bu.Contract == from {
					/*
						这个3币BalanceUpdate中有两笔是from地址的，一笔是出账，一笔是分配
					*/
					change := strings.TrimPrefix(bu.Change, "-")
					if change != amount { //当两个金额不想等的时候，表示这笔是分配产生的费用
						return change, false, nil
					}
				}
			}
			return "", false, nil
		}
	}
	// true 表示是假充值，false表示不是假充值
	return "", true, nil
}

func (ts *XtzService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	return nil, errors.New("unsupport it")
}

func (ts *XtzService) GetHeightByTxid(txid string) (int64, error) {
	return 0, errors.New("补推需要添加对应高度")
}

func (ts *XtzService) GetTxIsExist(height int64, txid string) bool {
	blockId := rpc.BlockIDLevel(int(height))
	resp, block, err := ts.client.Block(&blockId)
	if err != nil {
		return false
	}
	if resp.IsError() || block == nil {
		return false
	}
	if len(block.Operations) == 0 {
		return false
	}
	for _, txs := range block.Operations {
		if len(txs) > 0 {
			for _, tx := range txs {
				if tx.Hash == txid {
					if tx.ChainID == MainNetChainId {
						return true
					}
				}
			}
		}
	}
	return false
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ts := new(XtzService)
	ts.cfg = cfg
	ts.nodeCfg = nodeCfg
	ts.lock = sync.RWMutex{}
	ts.url = nodeCfg.Url
	// init other filed
	ts.client, _ = rpc.New(ts.url)
	return ts
}
