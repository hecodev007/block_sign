package zen

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"sync"
	"time"
	"zenDataServer/common"
	"zenDataServer/common/conf"
	"zenDataServer/common/log"
	dao "zenDataServer/models/po/zen"
	"zenDataServer/services"
	"zenDataServer/utils"
	rpc "zenDataServer/utils/zen"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig,watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.TxVoutRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(3395436)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
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

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	//starttime := time.Now()
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight2(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
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
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	workpool := utils.NewWorkPool(5) //一次性发太多请求会让节点窒息
	for index, tx := range block.Txs {
		//log.Info(index,tx.Txid)
		workpool.Add(1)
		go func(tx *rpc.Transaction, index int) {
			defer workpool.Dec()
			//defer log.Info("end:",tx.Txid)
			//过滤非转账交易
			if txInfo, err := parseBlockRawTX(s.RpcClient, tx, block.Hash, height); err != nil {
				log.Info(err.Error())
			} else if txInfo != nil {
				lock.Lock() //append并发不安全
				defer lock.Unlock()
				task.TxInfos = append(task.TxInfos, txInfo)
			}
		}(tx, index)
	}
	workpool.Wait()
	//if txjson, err := json.Marshal(task); err == nil {
	//	log.Infof("block:%v", string(txjson))
	//} else {
	//	log.Warn(err.Error())
	//}
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, tx *rpc.Transaction, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTxVout
	var vins []*dao.BlockTxVout
	 l := &sync.Mutex{}
	txcache := make(map[string]*rpc.Transaction, 0)
	if tx == nil {
		return nil, nil
	}

	blocktx := &dao.BlockTx{
		Txid:       tx.Txid,
		Height:     height,
		Blockhash:  blockhash,
		Version:    tx.Version,
		Voutcount:  len(tx.Vout),
		Vincount:   len(tx.Vin),
		Timestamp:  time.Unix(tx.Time, 0),
		Createtime: time.Now(),
		Fee:        "0",
	}
	outAmount := decimal.NewFromInt(0)
	inAmount := decimal.NewFromInt(0)
	for _, vout := range tx.Vout {
		outAmount = outAmount.Add(vout.Value)
		blocktxvout := &dao.BlockTxVout{
			Height:     height,
			Txid:       blocktx.Txid,
			VoutN:      vout.Index,
			Blockhash:  blocktx.Blockhash,
			Value:      vout.Value.String(),
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
	wkpool := utils.NewWorkPool(5)
	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			blocktx.Coinbase = 1
			continue
		}
		if vin.Txid == "" { //跳过挖矿交易
			log.Warn(tx.Txid, "empty vin.txid")
			continue
		}
		wkpool.Incr()
		go func(vin rpc.ProxyTxIn) {
			defer wkpool.Dec()

			l.Lock()
			vintx,ok :=txcache[vin.Txid]
			l.Unlock()
			if !ok {
				GetRawTransaction:
					//defer log.Info(tx.Txid, vin.Txid)
					tmptx, err := RpcClient.GetRawTransaction(vin.Txid)
					if err != nil {
						log.Warn(err.Error(), tx.Txid, vin.Txid)
						time.Sleep(time.Second*3)
						goto GetRawTransaction
					}
				l.Lock()
					txcache[vin.Txid] = &tmptx
				l.Unlock()
					vintx = &tmptx
			} else {
				//log.Info("cached",tx.Txid, vin.Txid)
			}

			inAmount = inAmount.Add(vintx.Vout[vin.Vout].Value)
			//获得vin对应的vout
			vout := vintx.Vout[vin.Vout]
			blocktxvin := &dao.BlockTxVout{
				Blockhash:  vintx.BlockHash,
				Value:      vout.Value.String(),
				Timestamp:  time.Unix(vintx.Time, 0),
				Createtime: time.Now(),
				Txid:       vin.Txid,
				VoutN:      vin.Vout,
				SpendTxid:  blocktx.Txid,
			}
			if address, err := vout.ScriptPubkey.GetAddress(); err == nil {
				blocktxvin.Address = address[0]
			}
			//data, _ := json.Marshal(vout.ScriptPubkey)
			//blocktxvin.Scriptpubkey = string(data)
			l.Lock()
			vins = append(vins, blocktxvin)
			l.Unlock()
		}(vin)
	}
	wkpool.Wait()
	fee := inAmount.Sub(outAmount)
	if fee.GreaterThan(decimal.NewFromInt(0)) {
		blocktx.Fee = fee.String()
	}
	return &TxInfo{
		Tx:    blocktx,
		Vouts: vouts,
		Vins:  vins,
	}, nil
}
