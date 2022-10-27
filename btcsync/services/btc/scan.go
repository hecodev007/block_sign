package btc

import (
	"btcsync/common"
	"btcsync/common/conf"
	"btcsync/common/log"
	dao "btcsync/models/po/btc"
	"btcsync/services"
	"btcsync/utils"
	rpc "btcsync/utils/btcrpc"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	Usdt           *rpc.RpcClient
	conf           conf.SyncConfig
	Watch          *services.WatchControl
	IrreverseBlock map[int64]common.ProcTask
	txlock         sync.Mutex
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	log.Info(node.Url)
	return &Scanner{
		RpcClient:      rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		Usdt:           rpc.NewRpcClient(node.Usdt, node.RPCKey, node.RPCSecret),
		conf:           conf.Sync,
		Watch:          watch,
		IrreverseBlock: make(map[int64]common.ProcTask),
		//txlock:    sync.Mutex{},
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.TxVoutRollBack(height)
}

func (s *Scanner) Init() error {
	if conf.Cfg.Sync.EnableRollback {
		s.Rollback(conf.Cfg.Sync.RollHeight)
	}

	return nil
}

func (s *Scanner) Clear() {
}

var i = int64(730530)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	countInfo, err := s.RpcClient.BtcGetBlockCount() //获取到的是区块个数
	if err != nil {
		return 0, err
	}
	return countInfo.Result, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	task, ok := s.IrreverseBlock[height]
	if ok {
		BlockHashRet, err := s.BtcGetBlockHash(height)
		if err != nil {
			return nil, err
		}
		if BlockHashRet.Result == task.GetBlockHash() {
			task.SetBestHeight(bestHeight, false)
			return task, nil
		}
	}
	task, err := s.scanBlock(height, bestHeight)
	if err != nil {
		return task, err
	}
	s.IrreverseBlock[height] = task
	for i := height - 150; i < height-100; i++ {
		delete(s.IrreverseBlock, i)
	}
	return task, nil
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	task, ok := s.IrreverseBlock[height]
	if ok {
		BlockHashRet, err := s.BtcGetBlockHash(height)
		if err != nil {
			return nil, err
		}
		if BlockHashRet.Result == task.GetBlockHash() {
			task.SetBestHeight(bestHeight, true)
			return task, nil
		}
	}
	task, err := s.scanBlock(height, bestHeight)
	if err != nil {
		return task, err
	}
	s.IrreverseBlock[height] = task
	for i := height - 150; i < height-100; i++ {
		delete(s.IrreverseBlock, i)
	}
	return task, nil

}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	//starttime := time.Now()
	//_, err := new(Scan).BlockByHeight(height)
	//if err != nil {
	//	log.Info(err.Error())
	//	return nil, err
	//}
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}
	//log.Info(block.Time)
	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Tx))
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
			Previousblockhash: block.Previousblockhash,
			Nextblockhash:     "",
			Transactions:      len(block.Tx),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Unix(block.Time, 0),
			Createtime:        time.Now(),
		},
	}

	scantxs, err := new(Scan).BlockByHeight(block.Height)
	//log.Info("scantxs", len(scantxs))
	if err == nil {
		txinfos, err := parseBlockRawTxScanTx(s.Watch, s.Usdt, block, scantxs)
		if err != nil {
			log.Info(err.Error())
		} else {
			task.TxInfos = txinfos
			return task, nil
		}
	} else {
		log.Info(err.Error())
	}
	//并发处理区块内的交易
	txPool := utils.NewWorkPool(4)
	for index, btcTxInfo := range block.Tx {
		txPool.Incr()
		go func(tx *rpc.BtcTxInfo) {
			defer txPool.Dec()
			_ = index
			//log.Info(tx.Hash)
			//if tx.Hash != "269bf08c145c218f4639cf7041d7c14109889d69ad31b003a828f398d2050c0b" {
			//	return
			//}
			//log.Info(xutils.String(tx))
			//defer log.Info("getrawtransaction success:", btcTxInfo.Hash, index, len(block.Tx))
			tx.Time = block.Time
			if txInfo, err := parseBlockRawTX(s.Watch, s.Usdt, s.RpcClient, tx, block.Hash, height); err != nil {
				log.Info(err.Error())
			} else if txInfo != nil {
				s.txlock.Lock()
				defer s.txlock.Unlock()
				task.TxInfos = append(task.TxInfos, txInfo)
			}
		}(btcTxInfo)
	}
	txPool.Wait()
	return task, nil
}

//解析交易
func parseBlockRawTX(watch *services.WatchControl, Usdt, RpcClient *rpc.RpcClient, tx *rpc.BtcTxInfo, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTxVout
	var vins []*dao.BlockTxVin
	var ContractTxs []*dao.BtcUsdtTx

	if tx.Locktime != 0 {
		//return nil, nil
	}
	txcache := make(map[string]*rpc.BtcTxInfo, 0)
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
		//过滤掉代币交易
		//if len(vout.Assets) != 0 {
		//	continue
		//}

		outAmount = outAmount.Add(vout.Value)
		blocktxvout := &dao.BlockTxVout{
			Height:     height,
			Txid:       blocktx.Txid,
			VoutN:      vout.N,
			Blockhash:  blocktx.Blockhash,
			Value:      vout.Value.String(),
			Timestamp:  blocktx.Timestamp,
			Createtime: time.Now(),
		}
		if len(vout.ScriptPubKey.Addresses) == 1 {
			blocktxvout.Address = vout.ScriptPubKey.Addresses[0]
		}
		data, _ := json.Marshal(vout.ScriptPubKey)
		blocktxvout.ScriptPubkey = string(data)
		if vout.ScriptPubKey.Type != "nulldata" {
			vouts = append(vouts, blocktxvout)
		}

		var isHasAssets bool
		if conf.Cfg.Sync.EnableUsdtScan && vout.ScriptPubKey.Type == "nulldata" && strings.HasPrefix(vout.ScriptPubKey.Hex, "6a146f6d6e69000000000000001f") || strings.HasPrefix(vout.ScriptPubKey.Hex, "6a4c146f6d6e69000000000000001f") {
			//log.Info(fmt.Sprintf("存在USDT数据，txid:%v %v", tx.Txid), tx.Hash)
			//判断存在USDT
			isHasAssets = true
		}
		if isHasAssets {
			omniTxInfo, err := Usdt.OmniGetrawtransaction(tx.Txid)
			if err != nil {
				log.Info(err.Error())
				panic(err.Error())
			}
			//log.Info(xutils.String(omniTxInfo))
			if omniTxInfo.Result.Propertyid == 31 {
				//usdt
				ContractTxs = append(ContractTxs, &dao.BtcUsdtTx{
					Txid:             tx.Txid,
					Sendingaddress:   omniTxInfo.Result.Sendingaddress,
					Referenceaddress: omniTxInfo.Result.Referenceaddress,
					Amount:           omniTxInfo.Result.Amount,
					Valid:            omniTxInfo.Result.Valid,
					//Blockhash        string          `gorm:"column:blockhash"`
					//Blocktime        int64           `gorm:"column:blocktime"`
					//Block            int64           `gorm:"column:block"`
				})
				//pushContractTxs = append(pushContractTxs, bo.PushContractTx{
				//	Coin:    "usdt",
				//	From:    omniTxInfo.Result.Sendingaddress,
				//	To:      omniTxInfo.Result.Referenceaddress,
				//	Amount:  omniTxInfo.Result.Amount,
				//	Fee:     omniTxInfo.Result.Fee,
				//	Valid:   omniTxInfo.Result.Valid,
				//	FeeCoin: "btc",
				//})
			}
		}
	}

	for _, vin := range tx.Vin {
		if vin.Coinbase != "" {
			blocktx.Iscoinbase = true
			continue
		}
		if vin.Txid == "" { //跳过挖矿交易
			log.Warn(tx.Txid, "empty vin.txid")
			continue
		}
		vintx, ok := txcache[vin.Txid]
		//watch.IsWatchAddressExist(vin.Address)
		if !ok {
			//log.Info(tx.Txid, vin.Txid)
		GetRawTransaction:
			tmptx, err := RpcClient.BtcGetrawtransaction(vin.Txid)
			if err != nil {
				log.Warn(err.Error(), tx.Txid, vin.Txid)
				time.Sleep(time.Second)
				goto GetRawTransaction
			}
			//过滤掉其他代币交易
			//if len(tmptx.Vout[vin.Vout].Assets) != 0 {
			//	continue
			//}
			txcache[vin.Txid] = tmptx
			vintx = tmptx
			//log.Info("success", vin.Txid)
		}

		inAmount = inAmount.Add(vintx.Vout[vin.Vout].Value)
		//获得vin对应的vout
		vout := vintx.Vout[vin.Vout]
		blocktxvin := &dao.BlockTxVin{
			Blockhash:  vintx.Hash,
			Value:      vout.Value.String(),
			Timestamp:  time.Unix(vintx.Time, 0),
			Createtime: time.Now(),
			Txid:       vin.Txid,
			VoutN:      vin.Vout,
			SpendTxid:  blocktx.Txid,
		}
		if len(vout.ScriptPubKey.Addresses) == 1 {
			blocktxvin.Address = vout.ScriptPubKey.Addresses[0]
		}
		data, _ := json.Marshal(vout.ScriptPubKey)
		blocktxvin.Scriptpubkey = string(data)

		vins = append(vins, blocktxvin)
	}
	fee := inAmount.Sub(outAmount)
	if fee.GreaterThan(decimal.NewFromInt(0)) {
		blocktx.Fee = fee.String()
	}
	return &TxInfo{
		Tx:         blocktx,
		Vouts:      vouts,
		Vins:       vins,
		Contractxs: ContractTxs,
	}, nil
}
func parseBlockRawTxScanTx(watch *services.WatchControl, Usdt *rpc.RpcClient, block *rpc.BtcBlockInfo, scanTxs []*Tx) (txs []*TxInfo, err error) {
	blockhash := block.Hash
	height := block.Height
	//log.Info(len(block.Tx))
	for _, nodetx := range block.Tx {
		if nodetx.Locktime != 0 {
			//continue
		}

		//log.Info(nodetx.Txid, nodetx.Hash)
		txid := nodetx.Txid
		var scantx *Tx
		for k, tmpscantx := range scanTxs {
			if strings.ToLower(tmpscantx.Txid) == strings.ToLower(txid) {
				scantx = scanTxs[k]
				break
			}
		}
		if scantx == nil {
			return nil, errors.New("txid:" + txid + " 在scanapi没找到")
		}
		//log.Info(index, nodetx.Txid)
		//if nodetx.Txid != "269bf08c145c218f4639cf7041d7c14109889d69ad31b003a828f398d2050c0b" {
		//	continue
		//}
		//log.Info("start")
		hasWatchAddr := false
		for _, v := range scantx.Inputs {
			if watch.IsWatchAddressExist(v.Address) {
				hasWatchAddr = true
				break
			}
		}
		for _, v := range scantx.Outputs {
			if watch.IsWatchAddressExist(v.Address) {
				hasWatchAddr = true
				break
			}
		}
		if !hasWatchAddr {
			continue
		}

		var vouts []*dao.BlockTxVout
		var vins []*dao.BlockTxVin
		var ContractTxs []*dao.BtcUsdtTx
		blocktx := &dao.BlockTx{
			Txid:       nodetx.Txid,
			Height:     height,
			Blockhash:  blockhash,
			Version:    nodetx.Version,
			Voutcount:  len(nodetx.Vout),
			Vincount:   len(nodetx.Vin),
			Timestamp:  time.Unix(block.Time, 0),
			Createtime: time.Now(),
			Fee:        scantx.Fee.String(),
		}

		for _, vout := range nodetx.Vout {
			//outAmount = outAmount.Add(vout.Value)
			blocktxvout := &dao.BlockTxVout{
				Height:     height,
				Txid:       blocktx.Txid,
				VoutN:      vout.N,
				Blockhash:  blocktx.Blockhash,
				Value:      vout.Value.String(),
				Timestamp:  blocktx.Timestamp,
				Createtime: time.Now(),
			}
			if len(vout.ScriptPubKey.Addresses) == 1 {
				blocktxvout.Address = vout.ScriptPubKey.Addresses[0]
			}
			data, _ := json.Marshal(vout.ScriptPubKey)
			blocktxvout.ScriptPubkey = string(data)
			if vout.ScriptPubKey.Type != "nulldata" {
				vouts = append(vouts, blocktxvout)
			}

			var isHasAssets bool
			if conf.Cfg.Sync.EnableUsdtScan && vout.ScriptPubKey.Type == "nulldata" && strings.HasPrefix(vout.ScriptPubKey.Hex, "6a146f6d6e69000000000000001f") || strings.HasPrefix(vout.ScriptPubKey.Hex, "6a4c146f6d6e69000000000000001f") {
				//log.Info(fmt.Sprintf("存在USDT数据，txid:%v %v", tx.Txid), tx.Hash)
				//判断存在USDT
				isHasAssets = true
			}
			if isHasAssets {
				omniTxInfo, err := Usdt.OmniGetrawtransaction(nodetx.Txid)
				if err != nil {
					log.Info(err.Error())
					panic(err.Error())
				}
				//log.Info(xutils.String(omniTxInfo))
				if omniTxInfo.Result.Propertyid == 31 {
					//usdt
					ContractTxs = append(ContractTxs, &dao.BtcUsdtTx{
						Txid:             nodetx.Txid,
						Sendingaddress:   omniTxInfo.Result.Sendingaddress,
						Referenceaddress: omniTxInfo.Result.Referenceaddress,
						Amount:           omniTxInfo.Result.Amount,
						Valid:            omniTxInfo.Result.Valid,
						//Blockhash        string          `gorm:"column:blockhash"`
						//Blocktime        int64           `gorm:"column:blocktime"`
						//Block            int64           `gorm:"column:block"`
					})
				}
			}
		}

		for no, vin := range nodetx.Vin {
			if vin.Coinbase != "" {
				blocktx.Iscoinbase = true
				continue
			}
			if vin.Txid == "" { //跳过挖矿交易
				//log.Warn(tx.Txid, "empty vin.txid")
				continue
			}

			//获得vin对应的vout
			scantxVin := scantx.Inputs[no]
			blocktxvin := &dao.BlockTxVin{
				//Blockhash:  scantx.,
				Address:    scantxVin.Address,
				Value:      scantxVin.Value.String(),
				Timestamp:  time.Unix(scantx.Time, 0),
				Createtime: time.Now(),
				Txid:       vin.Txid,
				VoutN:      vin.Vout,
				SpendTxid:  blocktx.Txid,
			}
			vins = append(vins, blocktxvin)
		}
		tmpBlockTxInfo := &TxInfo{
			Tx:         blocktx,
			Vins:       vins,
			Vouts:      vouts,
			Contractxs: ContractTxs,
		}
		txs = append(txs, tmpBlockTxInfo)
	}
	return
}
