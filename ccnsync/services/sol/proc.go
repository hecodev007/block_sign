package sol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/portto/solana-go-sdk/client"
	//"github.com/portto/solana-go-sdk/client/rpc"
	"github.com/shopspring/decimal"
	"log"
	"solsync/common"
	"solsync/common/conf"
	"solsync/models/bo"
	dao "solsync/models/po/yotta"
	"solsync/services"
	"time"
)

type Processor struct {
	client *client.Client
	watch  *services.WatchControl
	pusher *services.PushServer
	conf   conf.SyncConfig
}

func NewProcessor(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Processor {
	c := client.NewClient(node.RepushUrl)
	return &Processor{
		client: c,
		watch:  watch,
		conf:   conf.Sync,
	}
}
func (p *Processor) Init() error {

	for _, v := range p.watch.WatchAddrs {
		go p.UpdateAmount(v[0].Address)

	}
	return nil
}
func (p *Processor) Clear() {
}
func (p *Processor) SetPusher(push common.Pusher) {
	pusher, ok := push.(*services.PushServer)
	if ok {
		p.pusher = pusher
	}
}
func (p *Processor) RemovePusher() {
	p.pusher = nil
}

func (p *Processor) Info() (string, int64, error) {
	dbheight, err := dao.MaxBlockHeight()
	return p.conf.Name, dbheight, err
}

func (p *Processor) CheckIrreverseBlock(hash string) error {
	if has, err := dao.BlockHashExist(hash); err != nil {
		return fmt.Errorf("get BlockCount ByHash err: %v", err)
	} else if has {
		return fmt.Errorf("already have block  hash: %s , count: %d", hash, 1)
	}
	return nil
}

//以上全世界都一样

////暂没用到 ，查询数据是否已有这个区块
//CheckIrreverseBlock(hash string) error
////处理不可逆区块交易(推送交易，保存到数据库)
//ProcIrreverseTxs(ProcTask) error
////推送不可逆交易确认数（） 待定
//UpdateIrreverseConfirms(ProcTask)
////处理不可逆区块（保存到数据库）
//ProcIrreverseBlock(ProcTask) error
////处理可逆区块交易(是否需要推送)；有需要推送？，error
//ProcReverseTxs(ProcTask) (bool, error)
////推送可逆区块确认数(不需要保存数据库，直接推送数据)
//UpdateReverseConfirms(ProcTask)

//处理不可逆区块交易(推送交易，保存到数据库)
func (p *Processor) ProcIrreverseTask(procTask common.ProcTask) error {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	//tj, _ := json.Marshal(task)
	//log.Printf(string(tj))

	for _, txInfo := range task.TxInfos {

		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)

		if p.conf.FullBackup {
			dao.InsertTx(txInfo)

		} else if len(watchAddrs) > 0 {
			dao.InsertTx(txInfo)
		}
		if len(watchAddrs) > 0 {
			p.processPush(task.BestHeight, watchAddrs, []*dao.BlockTx{txInfo})
		}

	}
	dao.InsertBlock(task.Block)
	return nil
}
func (s *Processor) processPush(bestHeight int64, tmpWatchList map[string]bool, blocktxs []*dao.BlockTx) error {
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeAccountTX,
		Height:        blocktxs[0].BlockHeight,
		Hash:          blocktxs[0].BlockHash,
		CoinName:      s.conf.Name,
		Token:         "", //blocktx.CoinName
		Confirmations: bestHeight - blocktxs[0].BlockHeight + 1,
		Time:          blocktxs[0].Timestamp.Unix(),
	}
	for _, blocktx := range blocktxs {
		pushBlockTx.Txs = append(pushBlockTx.Txs, &bo.PushAccountTx{
			Name: blocktx.CoinName,
			//Txid:     "0x" + blocktx.Txid,
			Txid:     blocktx.Txid,
			From:     blocktx.FromAddress,
			To:       blocktx.ToAddress,
			Contract: blocktx.ContractAddress,
			Fee:      blocktx.Fee.String(),
			Amount:   blocktx.Amount.String(),
			Memo:     blocktx.Memo,
		})
	}
	// pushs.pusher
	pusdata, err := json.Marshal(&pushBlockTx)
	if err != nil {
		return fmt.Errorf("marshal err %v", err)
	}

	if s.pusher != nil {
		s.pusher.AddPushTask(pushBlockTx.Height, blocktxs[0].Txid, tmpWatchList, pusdata)
	}
	log.Printf(string(pusdata))
	return nil
}
func (p *Processor) parseContractTX(tx *dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Printf(string(txj))
	watchaddrs = make(map[string]bool)
	if tx.ContractAddress != "" && !p.watch.IsContractExist(tx.ContractAddress) {
		return
	}

	if p.watch.IsWatchAddressExist(tx.FromAddress) {
		watchaddrs[tx.FromAddress] = true
	}
	if p.watch.IsWatchAddressExist(tx.ToAddress) {
		watchaddrs[tx.ToAddress] = true
	}
	return
}
func (p *Processor) parseContractTxs(txs []*dao.BlockTx) (watchaddrs map[string]bool) {
	//txj, _ := json.Marshal(txs)
	//log.Printf(string(txj))
	watchaddrs = make(map[string]bool)
	for _, tx := range txs {

		if p.watch.IsWatchAddressExist(tx.FromAddress) {
			watchaddrs[tx.FromAddress] = true
		}
		if p.watch.IsWatchAddressExist(tx.ToAddress) {
			watchaddrs[tx.ToAddress] = true
		}
	}
	return
}
func (p *Processor) UpdateAmount(addr string) error {

	return nil
}

func (s *Processor) GetBestBlockHeight() (int64, error) {
	status, err := s.client.GetSlot(context.Background())
	if err != nil {
		log.Printf("%+v", err.Error())
		return 0, err
	}
	return int64(status), err
}

func (s *Processor) RepushTxWithHeight(userId int64, txid string, height int64) error {
	log.Printf("RepushTxWithHeight user: %d , txid : %s \n", userId, txid)
	if txid == "" {
		return fmt.Errorf("don't allow txid is empty")
	}
	bestBlockHeight, err := s.GetBestBlockHeight()
	if err != nil {
		return fmt.Errorf("BlockNumber err : %v", err)
	}
	blocktxs, watchs, err := s.getBlockTxFromNode(txid, height)
	if err != nil {
		return fmt.Errorf("don't get block tx %v", err)
	}
	return s.processPush(bestBlockHeight, watchs, blocktxs)
}

func (s *Processor) RepushTx(userid int64, txid string) error {
	return nil
	//var (
	//	err     error
	//	txinfos []*dao.BlockTx
	//)
	//log.Printf("RepushTx user: %d , txid : %s", userid, txid)
	//
	//if txid == "" {
	//	return fmt.Errorf("don't allow txid is empty")
	//}
	//
	//if txinfos, err = s.getBlockTxFromNode(txid); err != nil {
	//	return fmt.Errorf("%v", err)
	//}
	//if len(txinfos) == 0 {
	//	return fmt.Errorf("txid:%v 不存在监控地址", txid)
	//}
	//bestBlockHeight, err := s.BlockNumber()
	//if err != nil {
	//	return err
	//}
	//watchaddrs := s.parseContractTxs(txinfos)
	//if len(watchaddrs) == 0 {
	//	return fmt.Errorf("txid %s don't have care of address", txid)
	//}
	//
	//return s.processPush(txinfos, watchaddrs, bestBlockHeight)
}

//func (s *Processor) processTX(txInfo *TxInfo) (map[string]bool, []*dao.BlockTxVout, []*dao.BlockTxVout, error) {
//
//	if txInfo == nil {
//		return nil, nil, nil, fmt.Errorf("tx info don't allow nil")
//	}
//
//	var updateVins, insertVins []*dao.BlockTxVout
//	tmpWatchList := make(map[string]bool)
//
//	//starttime := time.Now()
//	amtout := decimal.Zero
//	for _, txvout := range txInfo.vouts {
//		amtout = amtout.Add(txvout.Value)
//
//		if s.watch.IsWatchAddressExist(txvout.Address) {
//			tmpWatchList[txvout.Address] = true
//		}
//	}
//
//	amtin := decimal.Zero
//	for _, txvin := range txInfo.vins {
//
//		if txvin.Txid == "coinbase" {
//			continue
//		}
//
//		if vout, err := dao.SelectBlockTXVout(txvin.Txid, txvin.Voutn); err == nil {
//			txvin.Id = vout.Id
//			txvin.Value = vout.Value
//			txvin.Address = vout.Address
//
//			if s.conf.FullBackup {
//				updateVins = append(updateVins, txvin)
//			} else {
//				if s.watch.IsWatchAddressExist(txvin.Address) {
//					updateVins = append(updateVins, txvin)
//				}
//			}
//		} else {
//			if err != gorm.ErrRecordNotFound {
//				log.Printf("processTX SelectBlockTXVout txid: %s, n: %d, err: %v", txvin.Txid, txvin.Voutn, err)
//				continue
//			}
//			tx, err := s.GetRawTransaction(txvin.Txid)
//			if err != nil {
//				log.Printf(err.Error())
//				return nil, nil, nil, fmt.Errorf("processTX : GetRawTransaction %s", txvin.Txid)
//			}
//			txvin.BlockHash = tx.BlockHash
//			txvin.Value = decimal.NewFromFloat(tx.Vout[txvin.Voutn].Value)
//			txvin.Timestamp = time.Unix(tx.Time, 0)
//			txvin.CreateTime = time.Now()
//			address, err := tx.Vout[txvin.Voutn].ScriptPubkey.GetAddress()
//			if err == nil {
//				txvin.Address = address[0]
//			}
//			data, _ := json.Marshal(tx.Vout[txvin.Voutn].ScriptPubkey)
//			txvin.ScriptPubKey = string(data)
//
//			if s.conf.FullBackup {
//				insertVins = append(insertVins, txvin)
//			} else {
//				if s.watch.IsWatchAddressExist(txvin.Address) {
//					insertVins = append(insertVins, txvin)
//				}
//			}
//		}
//
//		amtin = amtin.Add(txvin.Value)
//		if s.watch.IsWatchAddressExist(txvin.Address) {
//			tmpWatchList[txvin.Address] = true
//		}
//	}
//
//	txInfo.tx.Fee = amtin.Sub(amtout)
//	if txInfo.tx.Fee.IsNegative() && s.conf.Name != "doge" {
//		return nil, nil, nil, fmt.Errorf("tx fee don't allow negative,vin:%v, vout:%v", amtin, amtout)
//	}
//
//	//log.Printf("processTX : %s ,vins : %d , vouts : %d , used time : %f 's", blocktx.Txid, len(txvins), len(insertVouts), time.Since(starttime).Seconds())
//	return tmpWatchList, updateVins, insertVins, nil
//}
func (s *Processor) getBlockTxFromNode(txid string, height int64) ([]*dao.BlockTx, map[string]bool, error) {
	if height == 0 {
		return nil, nil, errors.New("repush flow need height")
	}
	block, err := s.client.RpcClient.GetBlock(context.Background(), uint64(height))
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	for _, t := range block.Result.Transactions {
		txJson, err := json.Marshal(t.Transaction)
		if err != nil {
			return nil, nil, fmt.Errorf("tx json marshal err: ", err.Error())
		}
		tx := &SolTx{}
		//fmt.Println(string(marshal2))
		err = json.Unmarshal(txJson, tx)
		if err != nil {
			return nil, nil, fmt.Errorf("tx json Unmarshal err: ", err.Error())
		}
		if tx.Signatures[0] == txid {
			if t.Meta.Err != nil {
				return nil, nil, fmt.Errorf("error tx : ", t.Meta.Err)
			}
			from, to, amount, feeAddr, fee, types, conAddr, err := ParseTransaction(s.watch, t, tx)
			if err != nil {
				return nil, nil, err
			}
			dAmount, err := decimal.NewFromString(amount)
			if err != nil {
				return nil, nil, err
			}
			//if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) &&
			//	!watch.IsWatchAddressExistToken(conAddr,from) && !watch.IsWatchAddressExistToken(conAddr,to){
			//	return nil, errors.New("没有监听的地址 code1")
			//}
			blocktxs := make([]*dao.BlockTx, 0)
			watchLists := make(map[string]bool)
			baseblocktx := dao.BlockTx{
				//Txid:        result.Transaction.Signatures[0],
				Txid:        txid,
				BlockHeight: height,
				BlockHash:   block.Result.Blockhash,
				Status:      "success",
				Timestamp:   time.Unix(*block.Result.BlockTime, 0),
			}
			//types: 1.主链币交易  2.代支付手续费主链币交易  3.代币交易  4.代支付手续费代币交易 5.创建地址交易
			if types == 1 {
				err := hasWatchAddress(s.watch, from, to)
				if err != nil {
					return nil, nil, err
				}
				blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
					dAmount.Shift(-9), decimal.NewFromInt(fee).Shift(-9))
				blocktxs = append(blocktxs, &blocktx)

				if s.watch.IsWatchAddressExist(from) {
					watchLists[from] = true
				}
				if s.watch.IsWatchAddressExist(to) {
					watchLists[to] = true
				}

			} else if types == 2 {
				err := hasWatchAddress(s.watch, from, to)
				if err != nil {
					return nil, nil, err
				}
				blocktx1 := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
					dAmount.Shift(-9), decimal.NewFromInt(0))
				blocktxs = append(blocktxs, &blocktx1)
				blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
					decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
				blocktxs = append(blocktxs, &blocktx2)

				if s.watch.IsWatchAddressExist(from) {
					watchLists[from] = true
				}
				if s.watch.IsWatchAddressExist(to) {
					watchLists[to] = true
				}
				if s.watch.IsWatchAddressExist(feeAddr) {
					watchLists[feeAddr] = true
				}

			} else if types == 3 {
				contract, err := s.watch.GetContract(conAddr)
				if err != nil {
					return nil, nil, errors.New("不支持该合约交易")
				}
				err = hasWatchAddress(s.watch, from, to)
				if err != nil {
					return nil, nil, err
				}
				//err = hasWatchTokenAddress(s.watch, conAddr, from, to)
				//if err != nil {
				//	return nil, nil, err
				//}
				mainFrom := from
				mainTo := to
				//mainFrom, _ := s.watch.GetWatchHashAddressToken(conAddr, from)
				//mainTo, _ := s.watch.GetWatchHashAddressToken(conAddr, to)
				//if mainFrom == "" && mainTo == "" {
				//	return nil, nil, errors.New("没有关心的地址 code1")
				//}
				//if mainFrom == "" {
				//	mainFrom = from
				//} else {
				//	watchLists[mainFrom] = true
				//}
				//if mainTo == "" {
				//	mainTo = to
				//} else {
				//	watchLists[mainTo] = true
				//}
				blocktx := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
					dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(fee).Shift(-9))
				blocktxs = append(blocktxs, &blocktx)

				if s.watch.IsWatchAddressExist(mainFrom) {
					watchLists[mainFrom] = true
				}
				if s.watch.IsWatchAddressExist(mainTo) {
					watchLists[mainTo] = true
				}

			} else if types == 4 {
				contract, err := s.watch.GetContract(conAddr)
				if err != nil {
					return nil, nil, errors.New("不支持该合约交易")
				}
				err = hasWatchAddress(s.watch, from, to)
				if err != nil {
					return nil, nil, err
				}
				//err = hasWatchTokenAddress(s.watch, conAddr, from, to)
				//if err != nil {
				//	return nil, nil, err
				//}
				//mainFrom, _ := s.watch.GetWatchHashAddressToken(conAddr, from)
				//mainTo, _ := s.watch.GetWatchHashAddressToken(conAddr, to)
				//if mainFrom == "" && mainTo == "" {
				//	return nil, nil, errors.New("没有关心的地址 code2")
				//}
				//if mainFrom == "" {
				//	mainFrom = from
				//} else {
				//	watchLists[mainFrom] = true
				//}
				//if mainTo == "" {
				//	mainTo = to
				//} else {
				//	watchLists[mainTo] = true
				//}
				mainFrom := from
				mainTo := to
				blocktx1 := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
					dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(0))
				blocktxs = append(blocktxs, &blocktx1)
				blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
					decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
				blocktxs = append(blocktxs, &blocktx2)
				if s.watch.IsWatchAddressExist(mainFrom) {
					watchLists[mainFrom] = true
				}
				if s.watch.IsWatchAddressExist(mainTo) {
					watchLists[mainTo] = true
				}
				if s.watch.IsWatchAddressExist(feeAddr) {
					watchLists[feeAddr] = true
				}

			} else if types == 5 {
				if !s.watch.IsWatchAddressExist(from) {
					return nil, nil, errors.New("没有关心的地址 code3")
				} else {
					watchLists[from] = true
				}
				contract, err := s.watch.GetContract(to) //这里的to是合约地址
				if err != nil {
					return nil, nil, errors.New("没有关心的地址 code4, err: " + err.Error())
				}
				blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, "create", "",
					dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(0))
				blocktxs = append(blocktxs, &blocktx)
			} else {
				return nil, nil, errors.New("error type")
			}
			if len(blocktxs) == 0 {
				return nil, nil, fmt.Errorf("dont't have care of tx")
			}
			return blocktxs, watchLists, nil
		}
	}
	return nil, nil, errors.New("高度不包含该笔交易")
	////result, err := s.client.GetConfirmedTransaction(context.Background(), txid)
	//result, err := s.client.RpcClient.GetTransaction(context.Background(), txid)
	////result, err := s.client.GetTransactionWithConfig(context.Background(), txid, rpc.GetTransactionConfig{
	////	Commitment: rpc.CommitmentFinalized,
	////}
	//if err != nil {
	//	log.Printf("%+v", err.Error())
	//	return nil, nil, err
	//}
	////if result.Meta.Err != nil || result.Meta.Status["Err"] != nil {
	////	return nil, nil, fmt.Errorf("error tx")
	////}
	//if result.Result.Meta.Err != nil  {
	//	return nil, nil, fmt.Errorf("error tx")
	//}
	//from, to, amount, feeAddr, fee, types, conAddr, err := ParseTransaction(s.watch, result.Result)
	//
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	////if types == 3 || types == 4 {
	////	//s.client.RpcClient.GetAccountInfoWithConfig()
	////	frominfo, err := s.client.RpcClient.GetAccountInfoWithConfig(context.Background(), from, rpc.GetAccountInfoConfig{Encoding: rpc.GetAccountInfoConfigEncodingJsonParsed})
	////	if err != nil {
	////		return nil, nil, err
	////	}
	////	//logs := info.Result.Value.Owner == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	////	fromrestlt := frominfo.Result.Value.Data.(map[string]interface{})
	////	fromi, ok := fromrestlt["parsed"].(map[string]interface{})
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 1")
	////	}
	////	fromr, ok := fromi["info"].(map[string]interface{})
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 2")
	////	}
	////	fromowner, ok := fromr["owner"]
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 4")
	////	}
	////	ofrom := fromowner.(string)
	////	if ofrom == from {
	////		return nil, nil, errors.New("parsed error code 5")
	////	} else {
	////		from = ofrom
	////	}
	////
	////	toinfo, err := s.client.RpcClient.GetAccountInfoWithConfig(context.Background(), to, rpc.GetAccountInfoConfig{Encoding: rpc.GetAccountInfoConfigEncodingJsonParsed})
	////	if err != nil {
	////		return nil, nil, err
	////	}
	////	//logs := info.Result.Value.Owner == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	////	torestlt := toinfo.Result.Value.Data.(map[string]interface{})
	////	toi, ok := torestlt["parsed"].(map[string]interface{})
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 11")
	////	}
	////	tor, ok := toi["info"].(map[string]interface{})
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 21")
	////	}
	////	toowner, ok := tor["owner"]
	////	if !ok {
	////		return nil, nil, errors.New("parsed error code 41")
	////	}
	////	oto := toowner.(string)
	////	if oto == to {
	////		return nil, nil, errors.New("parsed error code 51")
	////	} else {
	////		to = oto
	////	}
	////}
	//
	//dAmount, err := decimal.NewFromString(amount)
	//if err != nil {
	//	return nil, nil, err
	//}
	////if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) &&
	////	!watch.IsWatchAddressExistToken(conAddr,from) && !watch.IsWatchAddressExistToken(conAddr,to){
	////	return nil, errors.New("没有监听的地址 code1")
	////}
	//blocktxs := make([]*dao.BlockTx, 0)
	//watchLists := make(map[string]bool)
	//baseblocktx := dao.BlockTx{
	//	//Txid:        result.Transaction.Signatures[0],
	//	Txid:        txid,
	//	BlockHeight: height,
	//	BlockHash:   block.Result.Blockhash,
	//	Status:      "success",
	//	Timestamp:   time.Unix(*block.Result.BlockTime, 0),
	//}
	////types: 1.主链币交易  2.代支付手续费主链币交易  3.代币交易  4.代支付手续费代币交易 5.创建地址交易
	//if types == 1 {
	//	err := hasWatchAddress(s.watch, from, to)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
	//		dAmount.Shift(-9), decimal.NewFromInt(fee).Shift(-9))
	//	blocktxs = append(blocktxs, &blocktx)
	//
	//	if s.watch.IsWatchAddressExist(from) {
	//		watchLists[from] = true
	//	}
	//	if s.watch.IsWatchAddressExist(to) {
	//		watchLists[to] = true
	//	}
	//
	//} else if types == 2 {
	//	err := hasWatchAddress(s.watch, from, to)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	blocktx1 := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
	//		dAmount.Shift(-9), decimal.NewFromInt(0))
	//	blocktxs = append(blocktxs, &blocktx1)
	//	blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
	//		decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
	//	blocktxs = append(blocktxs, &blocktx2)
	//
	//	if s.watch.IsWatchAddressExist(from) {
	//		watchLists[from] = true
	//	}
	//	if s.watch.IsWatchAddressExist(to) {
	//		watchLists[to] = true
	//	}
	//	if s.watch.IsWatchAddressExist(feeAddr) {
	//		watchLists[feeAddr] = true
	//	}
	//
	//} else if types == 3 {
	//	contract, err := s.watch.GetContract(conAddr)
	//	if err != nil {
	//		return nil, nil, errors.New("不支持该合约交易")
	//	}
	//	err = hasWatchAddress(s.watch, from, to)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	//err = hasWatchTokenAddress(s.watch, conAddr, from, to)
	//	//if err != nil {
	//	//	return nil, nil, err
	//	//}
	//	mainFrom := from
	//	mainTo := to
	//	//mainFrom, _ := s.watch.GetWatchHashAddressToken(conAddr, from)
	//	//mainTo, _ := s.watch.GetWatchHashAddressToken(conAddr, to)
	//	//if mainFrom == "" && mainTo == "" {
	//	//	return nil, nil, errors.New("没有关心的地址 code1")
	//	//}
	//	//if mainFrom == "" {
	//	//	mainFrom = from
	//	//} else {
	//	//	watchLists[mainFrom] = true
	//	//}
	//	//if mainTo == "" {
	//	//	mainTo = to
	//	//} else {
	//	//	watchLists[mainTo] = true
	//	//}
	//	blocktx := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
	//		dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(fee).Shift(-9))
	//	blocktxs = append(blocktxs, &blocktx)
	//
	//	if s.watch.IsWatchAddressExist(mainFrom) {
	//		watchLists[mainFrom] = true
	//	}
	//	if s.watch.IsWatchAddressExist(mainTo) {
	//		watchLists[mainTo] = true
	//	}
	//
	//} else if types == 4 {
	//	contract, err := s.watch.GetContract(conAddr)
	//	if err != nil {
	//		return nil, nil, errors.New("不支持该合约交易")
	//	}
	//	err = hasWatchAddress(s.watch, from, to)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	//err = hasWatchTokenAddress(s.watch, conAddr, from, to)
	//	//if err != nil {
	//	//	return nil, nil, err
	//	//}
	//	//mainFrom, _ := s.watch.GetWatchHashAddressToken(conAddr, from)
	//	//mainTo, _ := s.watch.GetWatchHashAddressToken(conAddr, to)
	//	//if mainFrom == "" && mainTo == "" {
	//	//	return nil, nil, errors.New("没有关心的地址 code2")
	//	//}
	//	//if mainFrom == "" {
	//	//	mainFrom = from
	//	//} else {
	//	//	watchLists[mainFrom] = true
	//	//}
	//	//if mainTo == "" {
	//	//	mainTo = to
	//	//} else {
	//	//	watchLists[mainTo] = true
	//	//}
	//	mainFrom := from
	//	mainTo := to
	//	blocktx1 := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
	//		dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(0))
	//	blocktxs = append(blocktxs, &blocktx1)
	//	blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
	//		decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
	//	blocktxs = append(blocktxs, &blocktx2)
	//	if s.watch.IsWatchAddressExist(mainFrom) {
	//		watchLists[mainFrom] = true
	//	}
	//	if s.watch.IsWatchAddressExist(mainTo) {
	//		watchLists[mainTo] = true
	//	}
	//	if s.watch.IsWatchAddressExist(feeAddr) {
	//		watchLists[feeAddr] = true
	//	}
	//
	//} else if types == 5 {
	//	if !s.watch.IsWatchAddressExist(from) {
	//		return nil, nil, errors.New("没有关心的地址 code3")
	//	} else {
	//		watchLists[from] = true
	//	}
	//	contract, err := s.watch.GetContract(to) //这里的to是合约地址
	//	if err != nil {
	//		return nil, nil, errors.New("没有关心的地址 code4, err: " + err.Error())
	//	}
	//	blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, "create", "",
	//		dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(0))
	//	blocktxs = append(blocktxs, &blocktx)
	//} else {
	//	return nil, nil, errors.New("error type")
	//}
	//if len(blocktxs) == 0 {
	//	return nil, nil, fmt.Errorf("dont't have care of tx")
	//}
	//return blocktxs, watchLists, nil
}

func (p *Processor) ProcReverseTxs(procTask common.ProcTask) (ret bool, err error) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error task type")
	}
	for k, txInfo := range task.TxInfos {
		//过滤出监控地址的vin,vout
		watchAddrs := p.parseContractTX(txInfo)
		if len(watchAddrs) > 0 {
			ret = true
			p.processPush(task.BestHeight, watchAddrs, []*dao.BlockTx{task.TxInfos[k]})
		}
	}
	return ret, nil
}

func (p *Processor) PushReverseConfirms(procTask common.ProcTask) {
	task, ok := procTask.(*ProcTask)
	if !ok {
		panic("error ProcTask type")
	}
	pushBlockTx := &bo.PushAccountBlockInfo{
		Type:          bo.PushTypeConfir,
		Height:        task.Block.Height,
		Hash:          task.Block.Hash,
		CoinName:      p.conf.Name,
		Confirmations: task.BestHeight - task.Block.Height + 1,
		Time:          task.Block.Timestamp.Unix(),
	}

	// pushs.pusher
	pushdata, err := json.Marshal(&pushBlockTx)
	if err == nil && p.pusher != nil {
		p.pusher.AddPushUserTask(task.Block.Height, pushdata)
	}

}
