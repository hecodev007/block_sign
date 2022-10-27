package cfx

import (
	"cfxDataServer/common"
	"cfxDataServer/common/conf"
	"cfxDataServer/common/log"
	dao "cfxDataServer/models/po/cfx"
	"cfxDataServer/services"
	rpc "cfxDataServer/utils/mob"
	"fmt"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}

func (s *Scanner) Init() error {
	if s.conf.EnableRollback{
		s.Rollback(s.conf.EpochCount)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(781001)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
	return count-1, err
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
	//if height != 2071702{
	//	time.Sleep(time.Second*10000)
	//	return nil, errors.New("123")
	//}
	//starttime := time.Now()
//	log.Info(height, bestHeight)
	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            height,
			Hash:              fmt.Sprintf("%v",height),
			Previousblockhash: "",
			Nextblockhash:     "",
			Transactions:      0,
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}
	for _,monitorid := range monitorids {
	retryGetBlockByHeight:
		block, err := s.ProcessBlock(monitorid, height)
		if err != nil {
			log.Infof("%v height:%v", err.Error(), height)
			time.Sleep(time.Second * 3)
			goto retryGetBlockByHeight
		}
		for k, v := range block.Txs {
		getaddress:
			block.Txs[k].Address, err = s.GetAddress(monitorid, v.Subaddress_index)
			if err != nil {
				log.Infof("%v monitor:%v,addr:%v", err.Error(), monitorid, v.Subaddress_index)
				time.Sleep(time.Second * 3)
				goto getaddress
			}
		}
		//并发处理区块内的交易
		if txInfo, err := parseBlockRawTX( block.Txs, string(task.Block.Hash), task.Block.Height); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			task.TxInfos = append(task.TxInfos, txInfo...)
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", block.Height.ToInt().Int64(), string(ts), "task")
	return task, nil
}

//解析交易
func parseBlockRawTX( tx []*rpc.Txo, blockhash string, blockheight int64) (blocktxs []*dao.BlockTx, err error) {

	for _,v := range tx {
		blocktx := &dao.BlockTx{
			Txid:            v.Public_key,
			CoinName:        conf.Cfg.Sync.Name,
			ContractAddress: "",
			FromAddress:     "",
			ToAddress:       "",
			BlockHeight:     blockheight,
			BlockHash:       blockhash,
			Amount:          "",
			Status:          0,
			Fee:             "",
			Timestamp:       time.Now(),
			CreateTime:      time.Now(),
		}
		blocktx.Amount = v.Value.Shift(-12).String()
		if v.Direction =="received"{
			blocktx.ToAddress = v.Address
		} else if v.Direction == "spent" {
			blocktx.FromAddress = v.Address
		} else {
			log.Info("txo.Direction == "+v.Direction)
			panic("txo.Direction == "+v.Direction)
		}
		blocktxs = append(blocktxs,blocktx)
	}
	return blocktxs, nil

}
