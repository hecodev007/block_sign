package eth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"math/big"
	"net/http"
	"rsksync/common"
	"rsksync/conf"
	"rsksync/models/bo"
	dao "rsksync/models/po/eth"
	"rsksync/services"
	"rsksync/utils"
	"rsksync/utils/eth"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	*eth.RpcClient
	watch  *services.WatchControl
	pusher *services.PushServer

	conf conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	return &Processor{
		RpcClient: eth.NewRpcClient(node.Url),
		watch:     watch,
		conf:      conf.Sync,
	}
}

func (s *Processor) SetPusher(p common.Pusher) {
	pusher, ok := p.(*services.PushServer)
	if ok {
		s.pusher = pusher
	}
}

func (s *Processor) RemovePusher() {
	s.pusher = nil
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	log.Printf("RepushTx user: %d , txid : %s \n", userid, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.BlockNumber()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	//if blocktx, err := s.getBlockTxFromDB(txid); err == nil {
	//	return s.processTX([]*dao.BlockTX{blocktx}, bestBlockHeight)
	//} else {
	//	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	//	if err != nil {
	//		return fmt.Errorf("don't get block tx %v", err)
	//	}
	//	return s.processPush(watchs, bestBlockHeight, blocktxs...)
	//}

	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	return s.processPush(watchs, bestBlockHeight, blocktxs...)
	return nil
}

func (s *Processor) RepushTxByIsInternal(userid int64, txid string, isInternal bool) error {
	log.Printf("RepushTxByIsInternal user: %d , txid : %s", userid, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.BlockNumber()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	//if blocktx, err := s.getBlockTxFromDB(txid); err == nil {
	//	return s.processTX([]*dao.BlockTX{blocktx}, bestBlockHeight)
	//} else {
	//	blocktxs, watchs, err := s.getBlockTxFromNode(txid)
	//	if err != nil {
	//		return fmt.Errorf("don't get block tx %v", err)
	//	}
	//	return s.processPush(watchs, bestBlockHeight, blocktxs...)
	//}
	blocktxs := make([]*dao.BlockTX, 0)
	watchs := make(map[string]bool)
	if isInternal {
		blocktxs, watchs, err = s.getBlockTxFromETHAPI(txid)
		////先不推送，打印一下
		//dd0, _ := json.Marshal(blocktxs)
		//dd1, _ := json.Marshal(watchs)
		//log.Println(string(dd0))
		//log.Println(string(dd1))
		//return nil
	} else {
		blocktxs, watchs, err = s.getBlockTxFromNode(txid)
	}
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	return s.processPush(watchs, bestBlockHeight, blocktxs...)
	return nil
}

func (s *Processor) Info() (string, int64, error) {
	dbheight, err := dao.GetMaxBlockIndex()
	return s.conf.Name, dbheight, err
}

func (s *Processor) Init() error {
	return nil
}

func (s *Processor) Clear() {

}

func (s *Processor) CheckIrreverseBlock(hash string) error {
	cnt, err := dao.GetBlockCountByHash(hash)
	if err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	}
	if cnt > 0 {
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, cnt)
	}
	return nil
}

func (s *Processor) ProcIrreverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}
	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		for _, tmp := range tmps {
			go func(w *sync.WaitGroup, tx []*dao.BlockTX) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*dao.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*dao.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
			}
		}
	}
	return nil
}

func (s *Processor) ProcReverseTxs(tmps []interface{}, bestHeight int64) error {
	if len(tmps) <= 0 {
		return fmt.Errorf("ProcIrreverseTxs don't support len is zero")
	}
	if s.conf.EnableGoroutine {
		wg := &sync.WaitGroup{}
		for _, tmp := range tmps {
			go func(w *sync.WaitGroup, tx []*dao.BlockTX) {
				w.Add(1)
				defer w.Done()
				if err := s.processTX(tx, bestHeight); err == nil {
				}
			}(wg, tmp.([]*dao.BlockTX))
		}
		wg.Wait()
	} else {
		for _, tmp := range tmps {
			tx := tmp.([]*dao.BlockTX)
			if err := s.processTX(tx, bestHeight); err == nil {
			}
		}
	}
	return nil
}

func (s *Processor) ProcIrreverseBlock(b interface{}) error {
	block := b.(*dao.BlockInfo)
	if _, err := dao.InsertBlockInfo(block); err != nil {
		return fmt.Errorf("block %d Insert Block err : %v", block.Height, err)
	}
	return nil
}

func (s *Processor) UpdateIrreverseConfirms() {
	//查找所有未确认的区块
	if bs, err := dao.GetUnconfirmBlockInfos(s.conf.Confirmations + 6); err == nil && bs != nil && len(bs) > 0 {
		var ids []int64
		//开始同步更新确认数
		for _, blk := range bs {
			blk.Confirmations++
			s.confirmsPush(blk)
			ids = append(ids, blk.Id)
		}
		//批量更新订单确认数。
		if err := dao.BatchUpdateConfirmations(ids, 1); err != nil {
			log.Printf("batch update confirmations err: %v \n", err)
		}
	}
}

func (s *Processor) UpdateReverseConfirms(b interface{}) {
	s.confirmsPush(b.(*dao.BlockInfo))
}

func (s *Processor) getBlockTxFromDB(txid string) (*dao.BlockTX, error) {
	return dao.SelecBlockTxByHash(txid)
}

func (s *Processor) getBlockTxFromNode(txid string) ([]*dao.BlockTX, map[string]bool, error) {
	tx, err := s.GetTransactionByHash(txid)
	if err != nil || tx == nil {
		return nil, nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	block, err := s.GetBlockByNumber(tx.BlockNumber, false)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return nil, nil, err
	}
	status, _ := utils.ParseInt(txReceipt.Status)
	if s.conf.Name == "etc" {
		status = 1
	}
	if status != 1 {
		return nil, nil, fmt.Errorf("this contract tx status err : %d ", status)
	}
	//if txReceipt.Removed {
	//	return nil, nil, fmt.Errorf("this contract tx Removed err : %t ", txReceipt.Removed)
	//}

	blocktxs := make([]*dao.BlockTX, 0)
	watchLists := make(map[string]bool)
	if !s.IsContractTx(tx) {
		blocktx := &dao.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     eth.WEI,
			Timestamp:   time.Unix(block.Timestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
			GasUsed:     txReceipt.GasUsed,
			CreateTime:  time.Now(),
		}
		blocktxs = append(blocktxs, blocktx)
		if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
			watchLists[blocktx.FromAddress] = true
		}
		if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
			watchLists[blocktx.ToAddress] = true
		}
		log.Printf("dont't have care of watch address ,from: %s, to: %s \n", blocktx.FromAddress, blocktx.ToAddress)
	} else {
		log.Println("合约交易")
		for _, lg := range txReceipt.Logs {
			blocktx := &dao.BlockTX{
				BlockHeight:     tx.BlockNumber,
				BlockHash:       tx.BlockHash,
				Txid:            tx.Hash,
				Nonce:           tx.Nonce,
				GasUsed:         txReceipt.GasUsed,
				GasPrice:        tx.GasPrice.Int64(),
				Input:           tx.Input,
				CoinName:        s.conf.Name,
				Decimal:         eth.WEI,
				Timestamp:       time.Unix(block.Timestamp, 0),
				ContractAddress: lg.Address,
				CreateTime:      time.Now(),
			}
			if lg.Address == "0x1c2349acbb7f83d07577692c75b6d7654899bf10" {
				if len(lg.Topics) > 0 || lg.Topics[0] == "0x3efc190d59645f005a5974aa84aa94401ad997938870e7b2aa74a45138ad679b" {
					//mykey 合约处理
					haTopicsHash := make([]common2.Hash, 0)
					for _, vt := range lg.Topics {
						haTopicsHash = append(haTopicsHash, common2.HexToHash(vt))
					}
					vLog := types.Log{
						Address:     common2.HexToAddress(lg.Address),
						Topics:      haTopicsHash,
						Data:        common2.FromHex(lg.Data),
						BlockNumber: uint64(tx.BlockNumber),
						TxHash:      common2.HexToHash(txReceipt.TransactionHash),
						TxIndex:     uint(txReceipt.TransactionIndex),
						BlockHash:   common2.HexToHash(txReceipt.BlockHash),
						Index:       uint(txReceipt.LogIndex),
						Removed:     txReceipt.Removed,
					}
					sender, receiver, amountFloatStr, txid, err := MyKeyProcessTransferLogic(vLog)
					log.Printf("mykey 解析数据解析：sender：%s,receiver:%s,am:%s \n", sender, receiver, amountFloatStr)
					if err != nil {
						log.Printf("mykey 解析异常 err:%s \n", err.Error())
						continue
					}
					if txid != tx.Hash {
						log.Println("mykey txid 不一致")
						continue
					}

					am, _ := decimal.NewFromString(amountFloatStr)
					if sender == "" || receiver == "" || am.IsZero() {
						log.Printf("mykey txid 解析数据解析不全：sender：%s,receiver:%s,am:%s \n", sender, receiver, am.String())
						continue
					}
					blocktx.Amount = am.Shift(18) //扩大18变成int
					blocktx.FromAddress = sender
					blocktx.ToAddress = receiver
					blocktx.ContractAddress = "" //合约清空
				} else {
					continue
				}
			} else {
				contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
				if err != nil {
					log.Printf("dont't have care of watch contract : %s \n", blocktx.ContractAddress)
					continue
				}
				blocktx.CoinName = contractInfo.Name
				blocktx.Decimal = contractInfo.Decimal
				if lg.Data == "" || len(lg.Data) < 3 {
					continue
				}
				tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
				blocktx.Amount = decimal.NewFromBigInt(tmp, 0)
				if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
					continue
				}
				if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
					blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
					blocktx.ToAddress = "0x" + lg.Topics[2][26:66]
				} else {
					continue
				}
			}
			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				log.Printf("dont't have care of watch address ,from: %s, to: %s \n", blocktx.FromAddress, blocktx.ToAddress)
				continue
			}
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				watchLists[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				watchLists[blocktx.ToAddress] = true
			}
			log.Println("推送结果：", blocktx.ToAddress, "  amm:", blocktx.Amount)
			blocktxs = append(blocktxs, blocktx)
		}
	}
	if len(blocktxs) == 0 {
		return nil, nil, fmt.Errorf("dont't have care of tx")
	}
	return blocktxs, watchLists, nil
}

// 解析交易信息到db
func (s *Processor) processTX(blocktxs []*dao.BlockTX, bestHeight int64) error {

	if len(blocktxs) == 0 {
		return fmt.Errorf("tx is null")
	}
	txReceipt, err := s.GetTransactionReceipt(blocktxs[0].Txid)
	if err != nil {
		return err
	}
	status, _ := utils.ParseInt(txReceipt.Status)
	if s.conf.Name == "etc" {
		status = 1
	}
	//检测是否为关心的地址
	results := make([]*dao.BlockTX, 0)
	tmpWatchList := make(map[string]bool)
	for _, blocktx := range blocktxs {
		if blocktx.ContractAddress == "" {
			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				continue
			}
			blocktx.CreateTime = time.Now()
			blocktx.Decimal = eth.WEI
			blocktx.GasUsed = txReceipt.GasUsed
			blocktx.Status = status
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				tmpWatchList[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				tmpWatchList[blocktx.ToAddress] = true
			}
			if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
				log.Printf("block tx insert err: %v \n", err)
			}
			if blocktx.Status == 1 {
				results = append(results, blocktx)
			}
		} else {
			//合约交易
			contractInfo, err := s.watch.GetContract(blocktx.ContractAddress)
			if err != nil {
				continue
			}
			if !s.watch.IsWatchAddressExist(blocktx.FromAddress) && !s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				continue
			}
			blocktx.CoinName = contractInfo.Name
			blocktx.Decimal = contractInfo.Decimal
			blocktx.GasUsed = txReceipt.GasUsed
			blocktx.Status = status
			if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
				blocktx.Status = 2
			}
			if txReceipt.Logs != nil {
				btys, _ := json.Marshal(txReceipt.Logs)
				blocktx.Logs = string(btys)
			}
			if s.watch.IsWatchAddressExist(blocktx.FromAddress) {
				tmpWatchList[blocktx.FromAddress] = true
			}
			if s.watch.IsWatchAddressExist(blocktx.ToAddress) {
				tmpWatchList[blocktx.ToAddress] = true
			}
			if num, err := dao.InsertBlockTX(blocktx); num <= 0 || err != nil {
				log.Printf("block tx insert err: %v \n", err)
			}
			if blocktx.Status == 1 {
				results = append(results, blocktx)
			}
		}
	}
	if len(tmpWatchList) == 0 || len(results) == 0 {
		return fmt.Errorf("dont't have care of watch address ")
	}
	if status == 1 {
		s.processPush(tmpWatchList, bestHeight, results...)
	} else {
		log.Printf("block tx status : %d is failed \n", status)
	}
	return nil
}

func (s *Processor) processPush(tmpWatchList map[string]bool, bestHeight int64, blocktxs ...*dao.BlockTX) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].BlockHeight,
		Hash:          blocktxs[0].BlockHash,
		CoinName:      s.conf.Name,
		Token:         "",
		Confirmations: bestHeight - blocktxs[0].BlockHeight + 1,
		Time:          blocktxs[0].Timestamp.Unix(),
	}
	if blocktxs[0].CoinName != s.conf.Name {
		pushBlockTx.Token = blocktxs[0].CoinName
	}
	fee := decimal.New(blocktxs[0].GasPrice, 0).
		Mul(decimal.New(blocktxs[0].GasUsed, 0)).
		Shift(int32(0 - eth.WEI)).String()
	for _, blocktx := range blocktxs {
		amount := blocktx.Amount.Shift(int32(0 - blocktx.Decimal)).String()
		pushBlockTx.Txs = append(pushBlockTx.Txs, bo.PushAccountTx{
			Txid:     blocktx.Txid,
			From:     blocktx.FromAddress,
			To:       blocktx.ToAddress,
			Contract: blocktx.ContractAddress,
			Fee:      fee,
			Amount:   amount,
		})
	}
	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}
	if s.pusher != nil {
		log.Printf("推送结构：%s \n", string(pusdata))
		s.pusher.AddPushTask(pushBlockTx.Height, blocktxs[0].Txid, tmpWatchList, pusdata)
	}
	return nil
}

func (s *Processor) confirmsPush(blockInfo *dao.BlockInfo) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountConfir,
		Height:        blockInfo.Height,
		Hash:          blockInfo.Hash,
		CoinName:      s.conf.Name,
		Confirmations: blockInfo.Confirmations,
		Time:          blockInfo.Timestamp.Unix(),
	}
	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && s.pusher != nil {
		s.pusher.AddPushUserTask(blockInfo.Height, pushdata)
	}
	return nil
}

//定制化方法

func (s *Processor) getBlockTxFromETHAPI(txid string) ([]*dao.BlockTX, map[string]bool, error) {
	blocktxs := make([]*dao.BlockTX, 0)
	watchs := make(map[string]bool)

	tx, err := s.GetTransactionByHash(txid)
	if err != nil || tx == nil {
		return nil, nil, fmt.Errorf("GetTransactionByHash err: %v", err)
	}
	block, err := s.GetBlockByNumber(tx.BlockNumber, false)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockByNumber err: %v ", err)
	}
	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return nil, nil, err
	}
	status, _ := utils.ParseInt(txReceipt.Status)
	if status != 1 {
		return nil, nil, fmt.Errorf("this contract tx status err : %d ", status)
	}
	var itData []byte
	itData, err = getInternalEthTx(txid)
	if err != nil {
		return nil, nil, fmt.Errorf("getInternalEthTx err : %s ", err.Error())
	}
	internalData := new(bo.EthAPIInternal)
	err = json.Unmarshal(itData, internalData)
	if err != nil {
		return nil, nil, fmt.Errorf("getInternalEthTx err : %s ", err.Error())
	}
	if internalData.Status != "1" && len(internalData.Result) == 0 && internalData.Message != "OK" {
		return nil, nil, errors.New("getInternalEthTx err DATA")
	}
	for _, v := range internalData.Result {
		log.Printf("%+v", v)
		//是否存在关注交易
		has := false
		if v.IsError != "0" {
			continue
		}
		if v.Type != "call" {
			continue
		}
		if v.ContractAddress != "" {
			//暂时只允许eth合约交易
			continue
		}
		if v.Value.IsZero() {
			continue
		}
		if s.watch.IsWatchAddressExist(strings.ToLower(v.From)) {
			watchs[strings.ToLower(v.From)] = true
			has = true
		}
		if s.watch.IsWatchAddressExist(strings.ToLower(v.To)) {
			watchs[strings.ToLower(v.To)] = true
			has = true
		}

		if !has {
			log.Printf("dont't have care of watch address ,from: %s, to: %s \n", v.From, v.To)
			continue
		}
		log.Println("存在关注交易")
		blocktx := &dao.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: v.From,
			Nonce:       tx.Nonce,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       v.Input,
			CoinName:    s.conf.Name,
			Decimal:     eth.WEI,
			Timestamp:   time.Unix(block.Timestamp, 0),
			Amount:      v.Value, //int64
			ToAddress:   v.To,
			GasUsed:     txReceipt.GasUsed,
			CreateTime:  time.Now(),
		}
		blocktxs = append(blocktxs, blocktx)
	}
	if len(blocktxs) == 0 {
		return nil, nil, fmt.Errorf("dont't have care of tx")
	}
	return blocktxs, watchs, nil

}

func getInternalEthTx(txid string) ([]byte, error) {
	key := "Q5M7EJCBQWPTKGM5A5NS5IN9IXHKIBNJHY"
	url := "https://api.etherscan.io/api?module=account&action=txlistinternal&txhash=%s&apikey=%s"
	url = fmt.Sprintf(url, txid, key)
	// 超时时间：60秒
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	log.Printf("getInternalEthTx:%s \n", string(result.Bytes()))
	return result.Bytes(), nil
}
