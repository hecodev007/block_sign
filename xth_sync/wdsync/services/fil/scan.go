package fil

import (
	"wdsync/common"
	"wdsync/common/conf"
	"wdsync/common/log"
	dao "wdsync/models/po/fil"
	rpc "wdsync/utils/fil"
	"github.com/shopspring/decimal"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, "",node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	if _, err := dao.BlockChainRollBack(height); err != nil {
		panic(err.Error())
	}
	if _, err := dao.BlockRollBack(height); err != nil {
		panic(err.Error())
	}
	//if _, err := dao.TxRollBack(height); err != nil {
	//	panic(err.Error())
	//}
}
func (s *Scanner) Init() error {
	if conf.Cfg.Sync.EnableRollback {
		s.Rollback(conf.Cfg.Sync.RollHeight)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(140131)

//获取区块最新高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	return s.BlockHeight()
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}
func (s *Scanner) ChainHeadHeight() (int64, error) {
	return s.RpcClient.HeadHeight()
}
func (s *Scanner) ScanHead(height int64) error {

	CurrentHead, err := s.RpcClient.BlockHead()
	if err != nil {
		log.Info(err.Error())
		return err
	}
	if CurrentHead.Height < height {
		return nil
	}
	bestHeight := CurrentHead.Height
	var parent []map[string]string
	if len(CurrentHead.Blocks) == 0 {
		return nil
	}
	parent = CurrentHead.Blocks[0].Parents
	basefee, _ := decimal.NewFromString(CurrentHead.Blocks[0].ParentBaseFee)
	bc := &dao.BlockChain{
		Id:            CurrentHead.Height,
		Height:        CurrentHead.Height,
		Confirmations: bestHeight - CurrentHead.Height + 1,
		Blocknum:      len(CurrentHead.Blocks),
		Timestamp:     time.Now(),
		Cids:          CurrentHead.Cids,
		Parent:        parent,
		Exed:          0,
		Parentbasefee: basefee.String(),
	}
	dao.InsertBlockChain(bc)

	for {
		if (CurrentHead.Height-1)%10 == 0 {
			log.Info(height, "<-", CurrentHead.Height-1)
		}
		if CurrentHead.Height-1 < height {
			return nil
		}
		//如果数据库已经有了
		if dataBlock, err := dao.GetBlockChain(CurrentHead.Height - 1); err == nil {
			CurrentHead.Height -= 1
			CurrentHead.Cids = dataBlock.Cids
			CurrentHead.Blocks[0].Parents = dataBlock.Parent
			CurrentHead.Blocks[0].ParentBaseFee = dataBlock.Parentbasefee
			continue
		}
		var parent []map[string]string
		if len(CurrentHead.Blocks) > 0 {
			parent = CurrentHead.Blocks[0].Parents
		}
		//链上获取下一个块
		tmpCurrentHead, err := s.RpcClient.GetBlockChain(CurrentHead.Height-1, parent)
		if err != nil {
			//log.Info(height, CurrentHead.Height-1, bestHeight, ",", string(PA))
			//尝试恢复parent里面有些块不存在导致的错误
			var head *rpc.BlockHeader
			for i := 0; i < len(parent); i++ { //查找parent块是否存在
				head, err = s.RpcClient.GetBlockByCid(parent[i]["/"])
				if err == nil {
					break
				}
			}
			//parent块都不存在返回错误
			if err != nil {
				return err
			}
			//过滤掉不存在的块
			if head.Height == CurrentHead.Height-1 {
				for i := 0; i < len(parent); i++ {
					_, err2 := s.RpcClient.GetBlockByCid(parent[i]["/"])
					if err2 != nil {
						parent = append(parent[0:i], parent[i+1:len(parent)]...)
						i -= 1
					}
				}
				if len(parent) == 0 {
					log.Info(CurrentHead.Height, parent, err.Error())
					return err
				}
				tmpCurrentHead, err := s.RpcClient.GetBlockChain(CurrentHead.Height-1, parent)
				if err != nil {
					log.Info(CurrentHead.Height, parent, err.Error())
					return err
				}
				CurrentHead = tmpCurrentHead
			} else { //parent块高度不是上一个高度,跳过这个块
				CurrentHead.Height -= 1
				CurrentHead.Cids = CurrentHead.Cids[0:0]
				CurrentHead.Blocks = CurrentHead.Blocks
			}
		} else {
			CurrentHead = tmpCurrentHead
		}
		if len(CurrentHead.Blocks) > 0 {
			parent = CurrentHead.Blocks[0].Parents
		}
		//basefee, _ := decimal.NewFromString(CurrentHead.Blocks[0].ParentBaseFee)
		bc = &dao.BlockChain{
			Id:            CurrentHead.Height,
			Height:        CurrentHead.Height,
			Confirmations: bestHeight - CurrentHead.Height + 1,
			Blocknum:      len(CurrentHead.Cids),
			Timestamp:     time.Now(),
			Cids:          CurrentHead.Cids,
			Parent:        parent,
			Exed:          0,
			Parentbasefee: CurrentHead.Blocks[0].ParentBaseFee,
		}
		dao.InsertBlockChain(bc)
	}
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight, nil, nil, nil, decimal.Zero)
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	nextblock, err := s.GetBlockChain(height+1, nil)
	if err != nil {
		log.Info(height+1, err.Error())
		return nil, err
	}
	block, err := s.GetBlockChain(height, nil)
	if err != nil {
		log.Info(height, err.Error())
		return nil, err
	}
	var parent []map[string]string
	if len(block.Blocks) > 0 {
		parent = block.Blocks[0].Parents
	}
	return s.scanBlock(height, bestHeight, parent, nil, nextblock.Cids, decimal.Zero)
}
func (s *Scanner) ScanBaseBlock(height, bestHeight int64) (common.ProcTask, error) {

	block, err := dao.GetBlockChain(height)
	if err != nil {
		log.Info(height, err.Error())
		return nil, err
	}
	if len(block.Cids) == 0 {
		task := &ProcTask{
			Irreversible: bestHeight-height+1 >= s.conf.Confirmations,
			BestHeight:   bestHeight,
			BlockChain: &dao.BlockChain{
				Height:        height,
				Cids:          block.Cids,
				Parent:        block.Parent,
				Blocknum:      0,
				Confirmations: bestHeight - height + 1, //block.Confirmations
				Timestamp:     time.Now(),
				Exed:          1,
			},
			TxInfos: make([]*dao.BlockTx, 0),
		}

		return task, nil
	}
	nextheight := height + 1
	nextblock, err := dao.GetBlockChain(nextheight)
	if err != nil {
		log.Info(height+1, err.Error())
		return nil, err
	}

	for len(nextblock.Cids) == 0 {
		nextheight++
		nextblock, err = dao.GetBlockChain(nextheight)
		if err != nil {
			return nil, err
		}
	}

	basefee, _ := decimal.NewFromString(block.Parentbasefee)
	return s.scanBlock(height, bestHeight, block.Parent, block.Cids, nextblock.Cids, basefee)
}
func (s *Scanner) scanBlock(height, bestHeight int64, parent, Cids, nextCids []map[string]string, basefee decimal.Decimal) (tmp common.ProcTask, err error) {
	//starttime := time.Now()
	//getBlockChain:
	//	blockchain, err := s.GetBlockChain(height, cids)
	//	if err != nil {
	//		log.Warnf("%v height:%v", err.Error(), height)
	//		time.Sleep(time.Second * 3)
	//		goto getBlockChain
	//	}
	//	//log.Info(height, blockchain.Cids)
	//
	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		BlockChain: &dao.BlockChain{
			Height:        height,
			Cids:          dao.Cids(Cids),
			Parent:        parent,
			Blocknum:      len(Cids),
			Confirmations: bestHeight - height + 1, //block.Confirmations
			Timestamp:     time.Now(),
			Exed:          1,
		},
		TxInfos: make([]*dao.BlockTx, 0),
	}

	var messages []*rpc.Message
GetParentMessages:
	for _, cid := range nextCids {
		messages, err = s.GetParentMessages(cid["/"])
		if err == nil {
			break
		}else {
			log.Info(cid["/"],err.Error())
		}
	}
	if err != nil {
		log.Info(err.Error())
		log.Info(Cids)
		log.Info(nextCids)
		goto GetParentMessages
	}

	var receipts []*rpc.Receipt
GetParentReceipts:
	for _, cid := range nextCids {
		receipts, err = s.GetParentReceipts(cid["/"])
		if err == nil {
			break
		} else {
			log.Info(cid["/"],err.Error())
		}
	}
	if err != nil {
		log.Info(height,err.Error())
		log.Info(Cids)
		log.Info(nextCids)
		time.Sleep(time.Second)
		goto GetParentReceipts
	}
	//log.Info(txs)
	for txindex, tx := range messages {
		_ = txindex
		if tx.Message.Method != 0 {
			continue
		}
		if receipts[txindex].ExitCode != 0 {
			continue
		}

		tx.Message.Fee = basefee.Add(tx.Message.GasPremium).Mul(decimal.NewFromInt(tx.Message.GasLimit))
		if txInfo, err := s.parseBlockRawTX(s.conf.Name, tx, Cids[0]["/"], height); err != nil {
			log.Info(err.Error())
		} else if txInfo != nil {
			task.TxInfos = append(task.TxInfos, txInfo)
		}
	}

	return task, nil
}

//解析交易
func (s *Scanner) parseBlockRawTX(coinName string, msg *rpc.Message, blockhash string, height int64) (*dao.BlockTx, error) {
	if msg == nil {
		return nil, nil
	}
	tx := msg.Message
	blocktx := &dao.BlockTx{
		Txid:        tx.Cid["/"],
		BlockHeight: height,
		BlockHash:   blockhash,
		Version:     tx.Version,
		FromAddress: tx.From,
		ToAddress:   tx.To,
		Amount:      tx.Value.String(),
		Decimalmnt:  tx.Value.Shift(-18).String(),
		Status:      "success",
		Fee:         tx.Fee.String(),
		Gasfeecap:   tx.GasFeeCap.IntPart(),
		Gaslimit:    tx.GasLimit,
		Gaspremium:  tx.GasPremium.IntPart(),
		Nonce:       tx.Nonce,
		Method:      tx.Method,
		Timestamp:   time.Now(),
		Createtime:  time.Now(),
	}
	return blocktx, nil
}
