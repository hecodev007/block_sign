package sol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"

	//"github.com/portto/solana-go-sdk/client/rpc"
	"github.com/shopspring/decimal"
	"log"
	"solsync/common"
	"solsync/common/conf"
	dao "solsync/models/po/yotta"
	"solsync/services"
	"strings"

	//"solsync/utils"
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
	c := client.NewClient(node.Url)
	return &Scanner{
		client: c,
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
	status, err := s.client.GetSlot(context.Background())
	if err != nil {
		log.Printf("%+v", err.Error())
		return 0, err
	}
	return int64(status), err
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
	//s.client.SendTransaction()
	//s.client.SendTransactionWithConfig()
	//s.client.QuickSendTransaction()
	//s.client.SimulateTransaction()
	//s.client.SimulateTransactionWithConfig()

	starttime := time.Now()
	block, err := s.client.RpcClient.GetBlock(context.Background(), uint64(height))
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	task := &ProcTask{}
	if block.Error != nil {
		log.Printf("get block err: %s, skip block %d ", block.Error, height)
		task = &ProcTask{
			BestHeight: bestHeight,
			Block: &dao.BlockInfo{
				Height:            height,
				Hash:              "nilblock",
				Previousblockhash: "",
				//Timestamp:         time.Unix(*block.Result.BlockTime, 0),
				Transactions:  0,
				Confirmations: bestHeight - height + 1,
				Createtime:    time.Now(),
			},
		}
	} else {
		task = &ProcTask{
			BestHeight: bestHeight,
			Block: &dao.BlockInfo{
				Height:            height,
				Hash:              block.Result.Blockhash,
				Previousblockhash: block.Result.PreviousBlockhash,
				Timestamp:         time.Unix(*block.Result.BlockTime, 0),
				Transactions:      len(block.Result.Transactions),
				Confirmations:     bestHeight - height + 1,
				Createtime:        time.Now(),
			},
		}
	}

	//task = &ProcTask{
	//	BestHeight: bestHeight,
	//	Block: &dao.BlockInfo{
	//		Height:            height,
	//		Hash:              block.Result.Blockhash,
	//		Previousblockhash: block.Result.PreviousBlockhash,
	//		Timestamp:         time.Unix(*block.Result.BlockTime, 0),
	//		Transactions:      len(block.Result.Transactions),
	//		Confirmations:     bestHeight - height + 1,
	//		Createtime:        time.Now(),
	//	},
	//}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for _, tx := range block.Result.Transactions {
		//workpool.Incr()
		//go func(meta *rpc.TransactionMeta, tx *rpc.Transaction, task *ProcTask) {
		//	defer workpool.Dec()
		//s.batchParseTx(*block.BlockTime, block.Blockhash, height, tx.Meta, &tx.Transaction, task)
		s.batchParseTx(*block.Result.BlockTime, block.Result.Blockhash, height, tx, task)
		//}(&tx.Meta, &tx.Transaction, task)

	}
	//workpool.Wait()
	_ = starttime
	log.Printf("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(blockTime int64, blockhash string, height int64, meta rpc.GetBlockTransaction, task *ProcTask) {
	blockTxs, err := s.parseBlockTX(blockTime, blockhash, s.Watch, height, meta)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		//log.Printf(err.Error())
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(blockTime int64, blockhash string, watch *services.WatchControl, height int64, metas rpc.GetBlockTransaction) ([]*dao.BlockTx, error) {
	//if meta == nil || tx == nil {
	meta := metas.Meta
	txJson, err := json.Marshal(metas.Transaction)
	if err != nil {
		return nil, fmt.Errorf("tx json marshal err: ", err.Error())
	}
	tx := &SolTx{}
	//fmt.Println(string(marshal2))
	err = json.Unmarshal(txJson, tx)
	if err != nil {
		return nil, fmt.Errorf("tx json Unmarshal err: ", err.Error())
	}

	//tx := metas.Transaction
	if meta == nil {
		return nil, fmt.Errorf("tx is null")
	}

	var blocktxs = make([]*dao.BlockTx, 0)
	if meta.Err != nil {
		return nil, fmt.Errorf("error tx code1")
	}
	//_, ok := meta.Status["Ok"]
	//if !ok {
	//	return nil, fmt.Errorf("error tx code2")
	//}
	//token交易的from地址转主地址
	from, to, amount, feeAddr, fee, types, conAddr, err := ParseTransaction(watch, metas, tx)
	if err != nil {
		return nil, err
	}

	//if types == 3 || types == 4 {
	//	frominfo, err := s.client.GetAccountInfoWithCfg(context.Background(), from, rpc.GetAccountInfoConfig{Encoding: rpc.GetAccountInfoConfigEncodingJsonParsed})
	//	if err != nil {
	//		return nil, err
	//	}
	//	//fmt.Printf("from: %+v\n", from)
	//	//fmt.Printf("types: %+v\n", types)
	//	//fmt.Printf("frominfo: %+v\n", frominfo)
	//	//fmt.Printf("tx: %+v\n", tx)
	//	if frominfo.Result.Value.Owner != "" {
	//		fromrestlt := frominfo.Result.Value.Data.(map[string]interface{})
	//		fromi, ok := fromrestlt["parsed"].(map[string]interface{})
	//		if !ok {
	//			return nil, errors.New("parsed error code 1")
	//		}
	//		fromr, ok := fromi["info"].(map[string]interface{})
	//		if !ok {
	//			return nil, errors.New("parsed error code 2")
	//		}
	//		fromowner, ok := fromr["owner"]
	//		if !ok {
	//			return nil, errors.New("parsed error code 4")
	//		}
	//		from = fromowner.(string)
	//	}
	//
	//	toinfo, err := s.client.GetAccountInfoWithCfg(context.Background(), to, rpc.GetAccountInfoConfig{Encoding: rpc.GetAccountInfoConfigEncodingJsonParsed})
	//	if err != nil {
	//		return nil, err
	//	}
	//	if toinfo.Result.Value.Owner != "" {
	//		//logs := info.Result.Value.Owner == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	//		torestlt := toinfo.Result.Value.Data.(map[string]interface{})
	//		toi, ok := torestlt["parsed"].(map[string]interface{})
	//		if !ok {
	//			return nil, errors.New("parsed error code 1")
	//		}
	//		tor, ok := toi["info"].(map[string]interface{})
	//		if !ok {
	//			return nil, errors.New("parsed error code 2")
	//		}
	//		toowner, ok := tor["owner"]
	//		if !ok {
	//			return nil, errors.New("parsed error code 4")
	//		}
	//		to = toowner.(string)
	//	}
	//}

	dAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}
	//if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) &&
	//	!watch.IsWatchAddressExistToken(conAddr,from) && !watch.IsWatchAddressExistToken(conAddr,to){
	//	return nil, errors.New("没有监听的地址 code1")
	//}

	baseblocktx := dao.BlockTx{
		Txid:        tx.Signatures[0],
		BlockHeight: height,
		BlockHash:   blockhash,
		Status:      "success",
		Timestamp:   time.Unix(blockTime, 0),
	}
	//types: 1.主链币交易  2.代支付手续费主链币交易  3.代币交易  4.代支付手续费代币交易 5.创建地址交易
	if types == 1 {
		err := hasWatchAddress(watch, from, to)
		if err != nil {
			return nil, err
		}
		blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
			dAmount.Shift(-9), decimal.NewFromInt(fee).Shift(-9))
		blocktxs = append(blocktxs, &blocktx)
	} else if types == 2 {
		err := hasWatchAddress(watch, from, to)
		if err != nil {
			return nil, err
		}
		blocktx1 := buildBlockTx(baseblocktx, conf.Cfg.Name, from, to, "",
			dAmount.Shift(-9), decimal.NewFromInt(0))
		blocktxs = append(blocktxs, &blocktx1)
		blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
			decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
		blocktxs = append(blocktxs, &blocktx2)
	} else if types == 3 {
		contract, err := watch.GetContract(conAddr)
		if err != nil {
			return nil, errors.New("不支持该合约交易")
		}
		//err = hasWatchTokenAddress(watch, conAddr, from, to)
		//if err != nil {
		//	return nil, err
		//}

		err = hasWatchAddress(watch, from, to)
		if err != nil {
			return nil, err
		}
		mainFrom := from
		mainTo := to
		//mainFrom, _ := watch.GetWatchHashAddressToken(conAddr, from)
		//mainTo, _ := watch.GetWatchHashAddressToken(conAddr, to)
		//if mainFrom == "" && mainTo == "" {
		//	return nil, errors.New("没有关心的地址 code1")
		//}
		//if mainFrom == "" {
		//	mainFrom = from
		//}
		//if mainTo == "" {
		//	mainTo = to
		//}
		blocktx := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
			dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(fee).Shift(-9))
		blocktxs = append(blocktxs, &blocktx)
	} else if types == 4 {
		contract, err := watch.GetContract(conAddr)
		if err != nil {
			return nil, errors.New("不支持该合约交易")
		}
		//err = hasWatchTokenAddress(watch, conAddr, from, to)
		//if err != nil {
		//	return nil, err
		//}
		err = hasWatchAddress(watch, from, to)
		if err != nil {
			return nil, err
		}
		mainFrom := from
		mainTo := to

		//mainFrom, _ := watch.GetWatchHashAddressToken(conAddr, from)
		//mainTo, _ := watch.GetWatchHashAddressToken(conAddr, to)
		//if mainFrom == "" && mainTo == "" {
		//	return nil, errors.New("没有关心的地址 code2")
		//}
		//if mainFrom == "" {
		//	mainFrom = from
		//}
		//if mainTo == "" {
		//	mainTo = to
		//}
		blocktx1 := buildBlockTx(baseblocktx, contract.Name, mainFrom, mainTo, conAddr,
			dAmount.Shift(int32(0-contract.Decimal)), decimal.NewFromInt(0))
		blocktxs = append(blocktxs, &blocktx1)
		blocktx2 := buildBlockTx(baseblocktx, conf.Cfg.Name, feeAddr, "fee", "",
			decimal.NewFromInt(fee).Shift(-9), decimal.NewFromInt(0))
		blocktxs = append(blocktxs, &blocktx2)
	} else if types == 5 {
		if !watch.IsWatchAddressExist(from) {
			return nil, errors.New("没有关心的地址 code3")
		}
		_, err := watch.GetContract(to) //这里的to是合约地址
		if err != nil {
			return nil, errors.New("没有关心的地址 code4, err: " + err.Error())
		}
		blocktx := buildBlockTx(baseblocktx, conf.Cfg.Name, from, "create", "",
			dAmount.Shift(-9), decimal.NewFromInt(0))
		blocktxs = append(blocktxs, &blocktx)
	} else {
		return nil, errors.New("error type")
	}
	return blocktxs, nil
}

func hasWatchAddress(watch *services.WatchControl, from, to string) error {
	if !watch.IsWatchAddressExist(from) && !watch.IsWatchAddressExist(to) {
		return errors.New("没有监听的地址 code1")
	}
	return nil
}

//func hasWatchTokenAddress(watch *services.WatchControl, contract, from, to string) error {
//	if !watch.IsWatchAddressExistToken(contract, from) && !watch.IsWatchAddressExistToken(contract, to) {
//		return errors.New("没有监听的地址 code2")
//	}
//	return nil
//}

func buildBlockTx(blocktx dao.BlockTx, coinname, from, to, con string, amount, fee decimal.Decimal) dao.BlockTx {
	blocktx.CoinName = coinname
	blocktx.FromAddress = from
	blocktx.ToAddress = to
	blocktx.Amount = amount
	blocktx.Fee = fee
	blocktx.ContractAddress = con
	return blocktx
}

const (
	MainCoinLog               = "Program 11111111111111111111111111111111 invoke [1]"             //主链币交易
	TokenCoinLog              = "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]"  //代币交易
	CreateContractAddrLog     = "Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL invoke [1]" //创建合约子地址交易
	OwnerValidationProgramId1 = "Program 4MNPdKu9wFMvEeZBMt3Eipfs5ovVWTJb31pEXDJAAxX5 invoke [1]" //创建地址并转移代币
	OwnerValidationProgramId2 = "Program DeJBGdMFa1uynnnKiwrVioatTuHmNLpyFKnmB5kaFdzQ invoke [1]" //创建地址并转移代币

	MainCoin  = "Program 11111111111111111111111111111111"            //主链币交易
	TokenCoin = "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" //代币交易
)

type TokenAccount struct {
	Address  string
	Amount   decimal.Decimal
	Contract string
	Decimal  int64
}

type MainAccount struct {
	Address string
	Amount  int64
}

func ParseTransaction(watch *services.WatchControl, metas rpc.GetBlockTransaction, tx *SolTx) (from, to, amount, feeAddr string, fee int64, types int, conAddr string, err error) {
	meta := metas.Meta

	if len(tx.Signatures) == 0 {
		return "", "", "", "", 0, -1, "", errors.New("不关心的交易,code: error id")
	}
	if len(meta.LogMessages) <= 0 {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不关心的交易,code: 1", tx.Signatures)
	}
	if len(tx.Message.Accountkeys) <= 0 {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不关心的交易,code: 2", tx.Signatures)
	}
	logMsg := meta.LogMessages[0]
	accountKeys := tx.Message.Accountkeys

	isCare := false
	for _, logs := range meta.LogMessages {
		if strings.Contains(logs, MainCoin) || strings.Contains(logs, TokenCoin) {
			isCare = true
		}
	}
	if !isCare {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不关心的交易,code: 3", tx.Signatures)
	}

	for _, log := range meta.LogMessages {
		if strings.Contains(log, "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA") {
			logMsg = TokenCoinLog
		}
	}
	var (
		preAccount       []MainAccount
		postAccount      []MainAccount
		preTokenAccount  []TokenAccount
		postTokenAccount []TokenAccount
	)
	//for i, balance := range meta.PreBalances {
	//	preAccount = append(preAccount, MainAccount{
	//		Address: accountKeys[i],
	//		Amount:  balance,
	//	})
	//}
	//for i, balance := range meta.PostBalances {
	//	postAccount = append(postAccount, MainAccount{
	//		Address: accountKeys[i],
	//		Amount:  balance,
	//	})
	//}
	//for _, tokenBalance := range meta.PreTokenBalances {
	//	preTokenAccount = append(preTokenAccount, TokenAccount{
	//		Address:  tokenBalance.Owner,
	//		Amount:   tokenBalance.UITokenAmount.Amount,
	//		Contract: tokenBalance.Mint,
	//		//Decimal:  tokenBalance.UITokenAmount.Decimals,
	//		Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
	//	})
	//}
	//for _, tokenBalance := range meta.PostTokenBalances {
	//	postTokenAccount = append(postTokenAccount, TokenAccount{
	//		Address:  tokenBalance.Owner,
	//		Amount:   tokenBalance.UITokenAmount.Amount,
	//		Contract: tokenBalance.Mint,
	//		//Decimal:  tokenBalance.UITokenAmount.Decimals,
	//		Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
	//	})
	//}
	//sol正常出账交易
	//logmsg "Program 11111111111111111111111111111111 invoke [1]"
	// len(accountKey) == 3   && accountKey[0] == FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s
	//&& len(postTokenBalances) == 0  && len(preTokenBalances) == 0 && len(postBalances) == 3  && len(preBalances) == 3
	// sol-token 正常出账交易
	// logmsg Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]
	// len(accountKey) == 5  && accountKey[0] == FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s
	// && len(postBalances) == 5  && len(preBalances) == 5   && len(postTokenBalances) == 2  && len(preTokenBalances) == 2
	//sol-token 代支付手续费交易 (代币归集)
	//logmsg "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
	// len(accountKey) == 6   && accountKey[0] == FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s
	//&& len(postTokenBalances) == 2  && len(preTokenBalances) == 2 && len(postBalances) == 6  && len(preBalances) == 6
	//创建地址交易
	//logmsg Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL invoke [1]
	// len(accountKey) == 8  && accountKey[0] == FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s
	// && len(postBalances) == 8  && len(preBalances) == 8   && len(postTokenBalances) == 1  && len(preTokenBalances) == 0
	if logMsg == MainCoinLog {
		for i, balance := range meta.PreBalances {
			preAccount = append(preAccount, MainAccount{
				Address: accountKeys[i],
				Amount:  balance,
			})
		}
		for i, balance := range meta.PostBalances {
			postAccount = append(postAccount, MainAccount{
				Address: accountKeys[i],
				Amount:  balance,
			})
		}
		if (len(preAccount) == 3 || len(preAccount) == 5) && len(preAccount) == len(postAccount) { //正常交易
			if preAccount[0].Amount == postAccount[0].Amount {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 23", tx.Signatures)
			} else if preAccount[0].Amount < postAccount[0].Amount {
				from, to = preAccount[1].Address, preAccount[0].Address
				amount = decimal.NewFromInt(postAccount[0].Amount - preAccount[0].Amount).String()
			} else {
				from, to = preAccount[0].Address, preAccount[1].Address
				amount = decimal.NewFromInt(postAccount[1].Amount - preAccount[1].Amount).String()
			}
			fee = int64(meta.Fee)
			types = 1
		} else if len(preAccount) == 4 && len(preAccount) == len(postAccount) { //手续费代替支付的情况
			feeAddr = accountKeys[0]
			if preAccount[1].Amount == postAccount[1].Amount {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 24", tx.Signatures)
			} else if preAccount[1].Amount < postAccount[1].Amount {
				from, to = preAccount[2].Address, preAccount[1].Address
				amount = decimal.NewFromInt(postAccount[1].Amount - preAccount[1].Amount).String()
			} else {
				from, to = preAccount[1].Address, preAccount[2].Address
				amount = decimal.NewFromInt(postAccount[2].Amount - preAccount[2].Amount).String()
			}
			fee = int64(meta.Fee)
			types = 2
		} else { //其余情况暂不处理
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 13", tx.Signatures)
		}
	} else if logMsg == TokenCoinLog {
		//contractAddr := tx.Message.AccountKeys[len(tx.Message.AccountKeys)-1]
		var contractAddr string
		if len(meta.PostTokenBalances) > 0 {
			contractAddr = meta.PostTokenBalances[0].Mint
		} else {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(meta.PostTokenBalances长度不足),code: 39", tx.Signatures)
		}
		_, err := watch.GetContract(contractAddr)
		if err != nil {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3", tx.Signatures)
		}
		conAddr = contractAddr

		//====
		for _, tokenBalance := range meta.PreTokenBalances {
			dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
			}
			preTokenAccount = append(preTokenAccount, TokenAccount{
				Address:  tokenBalance.Owner,
				Amount:   dAmount,
				Contract: tokenBalance.Mint,
				Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
			})
		}
		for _, tokenBalance := range meta.PostTokenBalances {
			dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
			}
			postTokenAccount = append(postTokenAccount, TokenAccount{
				Address:  tokenBalance.Owner,
				Amount:   dAmount,
				Contract: tokenBalance.Mint,
				Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
			})
		}

		if len(postTokenAccount) > 2 || len(preTokenAccount) > 2 {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不支持合约交易,code: 3", tx.Signatures)
		}

		if len(postTokenAccount) == 1 && len(preTokenAccount) == 0 && len(accountKeys) == 8 {
			if accountKeys[0] == "FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s" ||
				accountKeys[0] == "FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea" ||
				accountKeys[0] == "BW1mQhRXsmEMndgj87A3JmZ1AmW2YSN3C5D36DbEaC8w" {
				if meta != nil && len(meta.PostTokenBalances) > 0 {
					contractAddr := meta.PostTokenBalances[0].Mint
					_, err := watch.GetContract(contractAddr)
					if err != nil {
						return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3ad3", tx.Signatures)
					}
					from = accountKeys[0]
					amount = decimal.NewFromInt(meta.PreBalances[0] - meta.PostBalances[0]).String()
					to = contractAddr
					fee = 0
					types = 5
				}
				log.Printf("捕获到一笔冷地址创建地址交易: %s.", tx.Signatures[0])
				goto END
			}
		}

		//sol-token 代支付手续费交易 (代币归集)
		//logmsg "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [1]",
		// len(accountKey) == 6   && accountKey[0] == FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s
		//&& len(postTokenBalances) == 2  && len(preTokenBalances) == 2 && len(postBalances) == 6  && len(preBalances) == 6

		types = 3
		if len(postTokenAccount) == 2 && len(preTokenAccount) == 2 && len(accountKeys) == 6 {
			if accountKeys[0] == "FgiJTETJS6LJ1mMiuXPxKFCjbGsxA56B9ke2egvPS12s" ||
				accountKeys[0] == "FvegQGzYvoHBHv2wYT1Nw2Xetjn48J7RkqL9RNGmQrea" ||
				accountKeys[0] == "BW1mQhRXsmEMndgj87A3JmZ1AmW2YSN3C5D36DbEaC8w" {
				feeAddr = accountKeys[0]
				types = 4
			}
		}

		//PostTokenBalances 2
		//PreTokenBalances 1或者2
		if len(postTokenAccount) == 2 {
			switch len(preTokenAccount) {
			case 1:
				from = preTokenAccount[0].Address
				var postAmount decimal.Decimal
				for _, post := range postTokenAccount {
					if post.Address != preTokenAccount[0].Address {
						to = post.Address
					} else {
						postAmount = post.Amount
					}
				}
				amount = preTokenAccount[0].Amount.Sub(postAmount).String()
				fee = int64(meta.Fee)
			case 2:
				if preTokenAccount[0].Amount == postTokenAccount[0].Amount {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 23", tx.Signatures)
				} else if preTokenAccount[0].Amount.LessThan(postTokenAccount[0].Amount) {
					from, to = preTokenAccount[1].Address, preTokenAccount[0].Address
					amount = postTokenAccount[0].Amount.Sub(preTokenAccount[0].Amount).String()
				} else {
					from, to = preTokenAccount[0].Address, preTokenAccount[1].Address
					amount = postTokenAccount[1].Amount.Sub(preTokenAccount[1].Amount).String()
				}
				fee = int64(meta.Fee)
			default:
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不支持合约交易,code: 3b3", tx.Signatures)
			}
		} else {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 不支持合约交易,code: 33", tx.Signatures)
		}

		//====

		//
		//
		//
		//if meta.PreTokenBalances[0].AccountIndex == 1 { //正常代币转账
		//	types = 3
		//} else if meta.PreTokenBalances[0].AccountIndex == 2 { //代币支付手续费代币转账, fee转为amount
		//	feeAddr = accountKeys[0]
		//	types = 4
		//} else {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 10", tx.Signatures)
		//}
		//
		//for _, tokenBalance := range meta.PreTokenBalances {
		//	preTokenAccount = append(preTokenAccount, TokenAccount{
		//		Address:  accountKeys[tokenBalance.AccountIndex],
		//		Amount:   tokenBalance.UITokenAmount.Amount,
		//		Contract: tokenBalance.Mint,
		//		Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
		//	})
		//}
		//for _, tokenBalance := range meta.PostTokenBalances {
		//	postTokenAccount = append(postTokenAccount, TokenAccount{
		//		Address:  accountKeys[tokenBalance.AccountIndex],
		//		Amount:   tokenBalance.UITokenAmount.Amount,
		//		Contract: tokenBalance.Mint,
		//		Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
		//	})
		//}
		//if len(preTokenAccount) != 2 || len(preTokenAccount) != len(postTokenAccount) {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 4", tx.Signatures)
		//}
		//for i, _ := range preTokenAccount {
		//	if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
		//		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 5", tx.Signatures)
		//	}
		//}
		//pre0, err := decimal.NewFromString(preTokenAccount[0].Amount)
		//if err != nil {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 5", tx.Signatures)
		//}
		//pre1, err := decimal.NewFromString(preTokenAccount[1].Amount)
		//if err != nil {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 6", tx.Signatures)
		//}
		//post0, err := decimal.NewFromString(postTokenAccount[0].Amount)
		//if err != nil {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 5", tx.Signatures)
		//}
		//post1, err := decimal.NewFromString(postTokenAccount[1].Amount)
		//if err != nil {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 6", tx.Signatures)
		//}
		//if pre0.Cmp(post0) == 0 {
		//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 7", tx.Signatures)
		//}
		//if pre0.Cmp(post0) < 0 {
		//	from, to = preTokenAccount[1].Address, preTokenAccount[0].Address
		//	amount = post0.Sub(pre0).String()
		//} else {
		//	from, to = preTokenAccount[0].Address, preTokenAccount[1].Address
		//	amount = post1.Sub(pre1).String()
		//}
		//fee = int64(meta.Fee)
	}
	//else if logMsg == CreateContractAddrLog {
	//	if len(meta.PreTokenBalances) == 1 && len(meta.PostTokenBalances) == 2 {
	//		conAddr = meta.PostTokenBalances[0].Mint
	//		i := meta.PreTokenBalances[0].AccountIndex
	//		from = accountKeys[i]
	//		for _, data := range meta.PostTokenBalances {
	//			if data.AccountIndex != i {
	//				to = accountKeys[data.AccountIndex]
	//				amount = data.UITokenAmount.Amount
	//			}
	//		}
	//		fee = int64(meta.Fee)
	//		types = 3
	//	} else {
	//		if meta != nil && len(meta.PostTokenBalances) > 0 {
	//			contractAddr := meta.PostTokenBalances[0].Mint
	//			_, err := watch.GetContract(contractAddr)
	//			if err != nil {
	//				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3ad3", tx.Signatures)
	//			}
	//			from = accountKeys[0]
	//			amount = decimal.NewFromInt(meta.PreBalances[0] - meta.PostBalances[0]).String()
	//			to = contractAddr
	//			fee = 0
	//			types = 5
	//		}
	//	}
	//} else if logMsg == OwnerValidationProgramId1 || logMsg == OwnerValidationProgramId2 {
	//	//var contractAddr string
	//	//if len(meta.PostTokenBalances) == 2 {
	//	//	contractAddr = meta.PostTokenBalances[0].Mint
	//	//} else {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(meta.PostTokenBalances长度不足),code: 3439", tx.Signatures)
	//	//}
	//	//
	//	//if len(meta.PreTokenBalances) != 1 {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(meta.PreTokenBalances) != 1,code: 34a39", tx.Signatures)
	//	//}
	//	//
	//	//contract, err := watch.GetContract(contractAddr)
	//	//if err != nil {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3213", tx.Signatures)
	//	//}
	//	//conAddr = contractAddr
	//	//if meta.PreTokenBalances[0].AccountIndex == 2 { //正常代币转账
	//	//	types = 3
	//	//} else {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 1033", tx.Signatures)
	//	//}
	//	//
	//	//for _, tokenBalance := range meta.PreTokenBalances {
	//	//	preTokenAccount = append(preTokenAccount, TokenAccount{
	//	//		Address:  accountKeys[tokenBalance.AccountIndex],
	//	//		Amount:   tokenBalance.UITokenAmount.Amount,
	//	//		Contract: tokenBalance.Mint,
	//	//		Decimal:  tokenBalance.UITokenAmount.Decimals,
	//	//	})
	//	//}
	//	//for _, tokenBalance := range meta.PostTokenBalances {
	//	//	postTokenAccount = append(postTokenAccount, TokenAccount{
	//	//		Address:  accountKeys[tokenBalance.AccountIndex],
	//	//		Amount:   tokenBalance.UITokenAmount.Amount,
	//	//		Contract: tokenBalance.Mint,
	//	//		Decimal:  tokenBalance.UITokenAmount.Decimals,
	//	//	})
	//	//}
	//	//if len(preTokenAccount) != 1 || len(postTokenAccount) != 2 {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 48hd", tx.Signatures)
	//	//}
	//	//for i, _ := range preTokenAccount {
	//	//	if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
	//	//		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t5c", tx.Signatures)
	//	//	}
	//	//}
	//	//pre0, err := decimal.NewFromString(preTokenAccount[0].Amount)
	//	//if err != nil {
	//	//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t75", tx.Signatures)
	//	//}
	//	//
	//	//from = preTokenAccount[0].Address
	//	//var postAmount decimal.Decimal
	//	//for _, post := range postTokenAccount {
	//	//	if post.Address != preTokenAccount[0].Address {
	//	//		to = post.Address
	//	//	} else {
	//	//		postAmount, err = decimal.NewFromString(post.Amount)
	//	//		if err != nil {
	//	//			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 524ea", tx.Signatures)
	//	//		}
	//	//	}
	//	//}
	//	//amount = pre0.Sub(postAmount).String()
	//	//fee = int64(meta.Fee)
	//}
END:
	if from == "" || to == "" || amount == "" {
		return "", "", "", "", 0, -1, "", errors.New("交易解析失败,code: 8")
	}

	dAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败, 金额解析出错. code: 78du", tx.Signatures)
	}
	if dAmount.LessThan(decimal.Zero) {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败, 金额错误. code: 78du", tx.Signatures)
	}
	return
}

func ParseTransactionRepush2(watch *services.WatchControl, metas *rpc.GetTransactionResult) (from, to, amount, feeAddr string, fee int64, types int, conAddr string, err error) {
	if metas == nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx is null")
	}
	meta := metas.Meta
	if meta.Err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("error tx code1")
	}
	txJson, err := json.Marshal(metas.Transaction)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx json marshal err: ", err.Error())
	}
	tx := &SolTx{}
	err = json.Unmarshal(txJson, tx)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx json Unmarshal err: ", err.Error())
	}

	return
}

func ParseTransactionRepush(watch *services.WatchControl, metas *rpc.GetTransactionResult) (from, to, amount, feeAddr string, fee int64, types int, conAddr string, err error) {
	if metas == nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx is null")
	}
	meta := metas.Meta
	if meta.Err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("error tx code1")
	}
	txJson, err := json.Marshal(metas.Transaction)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx json marshal err: ", err.Error())
	}
	tx := &SolTx{}
	err = json.Unmarshal(txJson, tx)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("tx json Unmarshal err: ", err.Error())
	}

	if len(tx.Signatures) < 0 {
		return "", "", "", "", 0, -1, "", errors.New("交易解析失败,code: error id")
	}
	if len(meta.LogMessages) <= 0 {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 1", tx.Signatures)
	}
	if len(tx.Message.Accountkeys) <= 0 {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 2", tx.Signatures)
	}
	logMsg := meta.LogMessages[0]
	accountKeys := tx.Message.Accountkeys

	for _, log := range meta.LogMessages {
		if strings.Contains(log, "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA") {
			logMsg = TokenCoinLog
		}
	}

	var (
		preAccount       []MainAccount
		postAccount      []MainAccount
		preTokenAccount  []TokenAccount
		postTokenAccount []TokenAccount
	)
	if logMsg == MainCoinLog {
		for i, balance := range meta.PreBalances {
			preAccount = append(preAccount, MainAccount{
				Address: accountKeys[i],
				Amount:  balance,
			})
		}
		for i, balance := range meta.PostBalances {
			postAccount = append(postAccount, MainAccount{
				Address: accountKeys[i],
				Amount:  balance,
			})
		}
		if (len(preAccount) == 3) && len(preAccount) == len(postAccount) { //正常交易
			if preAccount[0].Amount == postAccount[0].Amount {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 23", tx.Signatures)
			} else if preAccount[0].Amount < postAccount[0].Amount {
				from, to = preAccount[1].Address, preAccount[0].Address
				amount = decimal.NewFromInt(postAccount[0].Amount - preAccount[0].Amount).String()
			} else {
				from, to = preAccount[0].Address, preAccount[1].Address
				amount = decimal.NewFromInt(postAccount[1].Amount - preAccount[1].Amount).String()
			}
			fee = int64(meta.Fee)
			types = 1
		} else if len(preAccount) == 4 && len(preAccount) == len(postAccount) { //手续费代替支付的情况
			feeAddr = accountKeys[0]
			if preAccount[1].Amount == postAccount[1].Amount {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 24", tx.Signatures)
			} else if preAccount[1].Amount < postAccount[1].Amount {
				from, to = preAccount[2].Address, preAccount[1].Address
				amount = decimal.NewFromInt(postAccount[1].Amount - preAccount[1].Amount).String()
			} else {
				from, to = preAccount[1].Address, preAccount[2].Address
				amount = decimal.NewFromInt(postAccount[2].Amount - preAccount[2].Amount).String()
			}
			fee = int64(meta.Fee)
			types = 2
		} else if (len(preAccount) == 6 || len(preAccount) == 5) && len(preAccount) == len(postAccount) {
			count1 := 0
			count2 := 0
			for i, post := range postAccount {
				if post.Amount > preAccount[i].Amount {
					count1++
					to = post.Address
					amount = decimal.NewFromInt(post.Amount - preAccount[i].Amount).String()
				}
				if post.Amount < preAccount[i].Amount {
					count2++
					from = preAccount[i].Address
				}
			}
			if count1 > 1 || count2 > 1 {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 1223", tx.Signatures)
			}
			fee = int64(meta.Fee)
			types = 1
		} else { //其余情况暂不处理
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该结构交易),code: 13", tx.Signatures)
		}
	} else if logMsg == TokenCoinLog {
		//contractAddr := tx.Message.AccountKeys[len(tx.Message.AccountKeys)-1]
		var contractAddr string
		if len(meta.PostTokenBalances) > 0 {
			contractAddr = meta.PostTokenBalances[0].Mint
		} else {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(meta.PostTokenBalances长度不足),code: 39", tx.Signatures)
		}
		contract, err := watch.GetContract(contractAddr)
		if err != nil {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3", tx.Signatures)
		}
		conAddr = contractAddr
		if meta.PreTokenBalances[0].AccountIndex == 1 || meta.PreTokenBalances[0].AccountIndex == 3 { //正常代币转账
			types = 3
		} else if meta.PreTokenBalances[0].AccountIndex == 2 { //代币支付手续费代币转账, fee转为amount
			feeAddr = accountKeys[0]
			types = 4
		} else {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 10", tx.Signatures)
		}

		for _, tokenBalance := range meta.PreTokenBalances {
			dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
			}
			preTokenAccount = append(preTokenAccount, TokenAccount{
				Address:  tokenBalance.Owner,
				Amount:   dAmount,
				Contract: tokenBalance.Mint,
				Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
			})
		}
		for _, tokenBalance := range meta.PostTokenBalances {
			dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
			}
			postTokenAccount = append(postTokenAccount, TokenAccount{
				Address:  tokenBalance.Owner,
				Amount:   dAmount,
				Contract: tokenBalance.Mint,
				Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
			})
		}

		if len(preTokenAccount) == 1 {
			if len(preTokenAccount) != 1 || len(postTokenAccount) != 2 {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 48hd", tx.Signatures)
			}
			for i, _ := range preTokenAccount {
				if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t5c", tx.Signatures)
				}
			}
			pre0 := preTokenAccount[0].Amount

			from = preTokenAccount[0].Address
			var postAmount decimal.Decimal
			for _, post := range postTokenAccount {
				if post.Address != preTokenAccount[0].Address {
					to = post.Address
				} else {
					postAmount = post.Amount
				}
			}
			amount = pre0.Sub(postAmount).String()
			fee = int64(meta.Fee)
			goto END
		}

		if len(preTokenAccount) != 2 || len(preTokenAccount) != len(postTokenAccount) {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 4", tx.Signatures)
		}
		for i, _ := range preTokenAccount {
			if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 5", tx.Signatures)
			}
		}
		pre0 := preTokenAccount[0].Amount
		pre1 := preTokenAccount[1].Amount
		post0 := postTokenAccount[0].Amount
		post1 := postTokenAccount[1].Amount
		if pre0.Cmp(post0) == 0 {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 7", tx.Signatures)
		}
		if pre0.Cmp(post0) < 0 {
			from, to = preTokenAccount[1].Address, preTokenAccount[0].Address
			amount = post0.Sub(pre0).String()
		} else {
			from, to = preTokenAccount[0].Address, preTokenAccount[1].Address
			amount = post1.Sub(pre1).String()
		}
		fee = int64(meta.Fee)
	} else if logMsg == CreateContractAddrLog {
		if len(meta.PreTokenBalances) == 1 && len(meta.PostTokenBalances) == 2 {
			conAddr = meta.PostTokenBalances[0].Mint
			i := meta.PreTokenBalances[0].AccountIndex
			from = accountKeys[i]
			for _, data := range meta.PostTokenBalances {
				if data.AccountIndex != i {
					to = accountKeys[data.AccountIndex]
					amount = data.UITokenAmount.Amount
				}
			}
			fee = int64(meta.Fee)
			types = 3
		} else {
			if meta != nil && len(meta.PostTokenBalances) > 0 {
				contractAddr := meta.PostTokenBalances[0].Mint
				_, err := watch.GetContract(contractAddr)
				if err != nil {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3ad3", tx.Signatures)
				}
				from = accountKeys[0]
				amount = decimal.NewFromInt(meta.PreBalances[0] - meta.PostBalances[0]).String()
				to = contractAddr
				fee = 0
				types = 5
			}
		}
	} else if logMsg == OwnerValidationProgramId1 || logMsg == OwnerValidationProgramId2 {
		var contractAddr string
		if len(meta.PostTokenBalances) == 2 {
			contractAddr = meta.PostTokenBalances[0].Mint
		} else {
			return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(meta.PostTokenBalances长度不足),code: 3439", tx.Signatures)
		}

		if len(meta.PreTokenBalances) == 1 {
			contract, err := watch.GetContract(contractAddr)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3213", tx.Signatures)
			}
			conAddr = contractAddr
			if meta.PreTokenBalances[0].AccountIndex == 2 || meta.PreTokenBalances[0].AccountIndex == 1 { //正常代币转账
				types = 3
			} else {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 1033", tx.Signatures)
			}

			for _, tokenBalance := range meta.PreTokenBalances {
				dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
				if err != nil {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
				}
				preTokenAccount = append(preTokenAccount, TokenAccount{
					Address:  tokenBalance.Owner,
					Amount:   dAmount,
					Contract: tokenBalance.Mint,
					Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
				})
			}
			for _, tokenBalance := range meta.PostTokenBalances {
				dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
				if err != nil {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
				}
				postTokenAccount = append(postTokenAccount, TokenAccount{
					Address:  tokenBalance.Owner,
					Amount:   dAmount,
					Contract: tokenBalance.Mint,
					Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
				})
			}
			if len(preTokenAccount) != 1 || len(postTokenAccount) != 2 {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 48hd", tx.Signatures)
			}
			for i, _ := range preTokenAccount {
				if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t5c", tx.Signatures)
				}
			}
			pre0 := preTokenAccount[0].Amount
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t75", tx.Signatures)
			}

			from = preTokenAccount[0].Address
			var postAmount decimal.Decimal
			for _, post := range postTokenAccount {
				if post.Address != preTokenAccount[0].Address {
					to = post.Address
				} else {
					postAmount = post.Amount
					if err != nil {
						return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 524ea", tx.Signatures)
					}
				}
			}
			amount = pre0.Sub(postAmount).String()
			fee = int64(meta.Fee)
		} else if len(meta.PreTokenBalances) == 2 {
			contract, err := watch.GetContract(contractAddr)
			if err != nil {
				return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(不支持该合约),code: 3213", tx.Signatures)
			}
			conAddr = contractAddr
			//if meta.PreTokenBalances[0].AccountIndex == 2 { //正常代币转账
			//	types = 3
			//} else {
			//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 10s33", tx.Signatures)
			//}
			types = 3
			for _, tokenBalance := range meta.PreTokenBalances {
				dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
				if err != nil {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
				}
				preTokenAccount = append(preTokenAccount, TokenAccount{
					Address:  tokenBalance.Owner,
					Amount:   dAmount,
					Contract: tokenBalance.Mint,
					Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
				})
			}
			for _, tokenBalance := range meta.PostTokenBalances {
				dAmount, err := decimal.NewFromString(tokenBalance.UITokenAmount.Amount)
				if err != nil {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败(UITokenAmount 类型错误),code: 3", tx.Signatures)
				}
				postTokenAccount = append(postTokenAccount, TokenAccount{
					Address:  tokenBalance.Owner,
					Amount:   dAmount,
					Contract: tokenBalance.Mint,
					Decimal:  int64(tokenBalance.UITokenAmount.Decimals),
				})
			}
			//if len(preTokenAccount) != 1 || len(postTokenAccount) != 2 {
			//	return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 48hd", tx.Signatures)
			//}
			for i, _ := range preTokenAccount {
				if preTokenAccount[i].Contract != contract.ContractAddress || postTokenAccount[i].Contract != contract.ContractAddress {
					return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: t5c", tx.Signatures)
				}
			}
			pre0 := preTokenAccount[0].Amount
			from = preTokenAccount[0].Address
			var postAmount decimal.Decimal
			for _, post := range postTokenAccount {
				if post.Address != preTokenAccount[0].Address {
					to = post.Address
				} else {
					postAmount = post.Amount
					if err != nil {
						return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败,code: 524ea", tx.Signatures)
					}
				}
			}
			amount = pre0.Sub(postAmount).String()
			fee = int64(meta.Fee)
		}
	}

END:
	if from == "" || to == "" || amount == "" {
		return "", "", "", "", 0, -1, "", errors.New("交易解析失败,code: 8")
	}
	dAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败, 金额解析出错. code: 78du", tx.Signatures)
	}
	if dAmount.LessThan(decimal.Zero) {
		return "", "", "", "", 0, -1, "", fmt.Errorf("id: %v. 交易解析失败, 金额错误. code: 78du", tx.Signatures)
	}
	return
}

type SolTx struct {
	Message struct {
		Accountkeys []string `json:"accountKeys"`
		Header      struct {
			Numreadonlysignedaccounts   int `json:"numReadonlySignedAccounts"`
			Numreadonlyunsignedaccounts int `json:"numReadonlyUnsignedAccounts"`
			Numrequiredsignatures       int `json:"numRequiredSignatures"`
		} `json:"header"`
		Instructions []struct {
			Accounts       []int  `json:"accounts"`
			Data           string `json:"data"`
			Programidindex int    `json:"programIdIndex"`
		} `json:"instructions"`
		Recentblockhash string `json:"recentBlockhash"`
	} `json:"message"`
	Signatures []string `json:"signatures"`
}
