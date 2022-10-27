package ghost

import (
	"encoding/json"
	"fmt"
	"ghostDataServerNew/common"
	"ghostDataServerNew/common/conf"
	"ghostDataServerNew/common/log"
	dao "ghostDataServerNew/models/po/ghost"
	"ghostDataServerNew/utils"
	util "ghostDataServerNew/utils/btc"
	"runtime"
	"sync"
	"time"
)

type Scanner struct {
	*util.RpcClient

	lock            *sync.Mutex
	taskJobs        []*ScanTask
	jobsNum         int
	EnableGoroutine bool
	conf            conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: util.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		lock:      &sync.Mutex{},
		jobsNum:   runtime.NumCPU(),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
}

func (s *Scanner) Init() error {
	for i := 0; i < s.jobsNum; i++ {
		s.taskJobs = append(s.taskJobs, &ScanTask{
			Txids: make(chan string, 50000),
			Done:  make(chan int, 50000),
		})
	}
	return nil
}

func (s *Scanner) Clear() {
	for i := 0; i < s.jobsNum; i++ {
		close(s.taskJobs[i].Txids)
		close(s.taskJobs[i].Done)
	}
}
var i = int64(58776)
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i,nil
	//这里得到的是块的数量
	h, err := s.GetBlockCount()
	if err != nil {
		return h, err
	}
	h = h - 1 //

	h -= s.conf.Delaynum
	if h < 0 {
		h = 0
	}
	return h, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	return nil
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
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	if has, err := dao.BlockHashExist(block.Hash); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.Hash,
			Previousblockhash: block.PreviousBlockHash,
			Nextblockhash:     "",
			Transactions:      len(block.Txs),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Unix(block.Time, 0),
			Createtime:        time.Now(),
		},
		TxInfos: make([]*TxInfo,0),
	}
	//并发处理区块内的交易

	lock := new(sync.Mutex)
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for index, txid := range block.Txs {
		workpool.Add(1)
		go func(txid string, index int) {
			defer workpool.Dec()
			//过滤非转账交易

		getTransaction:
			tx, err := s.GetRawTransaction(txid)
			if err == nil {
				if txInfo, err := parseBlockRawTX(s.RpcClient,&tx, block.Hash, height); err != nil {
					log.Info("err:",err.Error())
				} else if txInfo != nil {
					lock.Lock() //append并发不安全
					defer lock.Unlock()
					task.TxInfos = append(task.TxInfos, txInfo)
				}
			} else {
				log.Warn(err.Error(), txid)
				goto getTransaction
			}
		}(txid, index)
	}
	workpool.Wait()


	//if txjson,err := json.Marshal(task);err ==nil {
	//	log.Infof("block:%v",string(txjson))
	//} else {
	//	log.Warn(err.Error())
	//}
	return task, nil
}

//解析交易
func parseBlockRawTX(rpc *util.RpcClient,tx *util.Transaction, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTxVout
	var vins []*dao.BlockTxVout

	if tx == nil {
		return nil, fmt.Errorf("txid is null")
	}

	//log.Infof("parse tx : %s",tx.Txid)
	blocktx := &dao.BlockTx{
		Txid:        tx.Txid,
		Height: height,
		Blockhash:   blockhash,
		Version:     tx.Version,
		Size:        tx.Size,
		Voutcount:   len(tx.Vout),
		Vincount:    len(tx.Vin),
		Timestamp:   time.Unix(tx.Time, 0),
		Createtime:  time.Now(),
		Fee: "0",
	}

	for _, vout := range tx.Vout {
		blocktxvout := &dao.BlockTxVout{
			Txid:       blocktx.Txid,
			VoutN:      vout.Index,
			Blockhash:  blocktx.Blockhash,
			Value:      vout.Value.String(),
			Status:     0,
			Timestamp:  blocktx.Timestamp,
			Createtime: time.Now(),
		}
		if address, err := vout.ScriptPubkey.GetAddress(); err == nil {
			blocktxvout.Address = address[0]
		}
		data, _ := json.Marshal(vout.ScriptPubkey)
		blocktxvout.ScriptPubkey = string(data)

		vouts = append(vouts, blocktxvout)
	}

	for _, vin := range tx.Vin {

		if vin.Txid == "" {
			vin.Txid = "coinbase"
			blocktx.Coinbase = 1
			continue
		}
		getRawTransaction:
		txin,err :=rpc.GetRawTransaction(vin.Txid)
		if err != nil {
			log.Warn(err.Error(),vin.Txid)
			goto getRawTransaction
		}
		blocktxvin := &dao.BlockTxVout{
			Txid:      vin.Txid,
			VoutN:     vin.Vout,
			SpendTxid: blocktx.Txid,
			Value: txin.Vout[vin.Vout].Value.String(),

			Status:    1,
		}
		if address, err := txin.Vout[vin.Vout].ScriptPubkey.GetAddress(); err == nil {
			blocktxvin.Address = address[0]
		}
		vins = append(vins, blocktxvin)
	}

	if blocktx.Coinbase == 1 {
		for _, vout := range vouts {
			vout.Coinbase = 1
		}
	}

	return &TxInfo{
		Tx:    blocktx,
		Vouts: vouts,
		Vins:  vins,
	}, nil
}
