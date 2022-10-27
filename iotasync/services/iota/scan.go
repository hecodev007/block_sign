package iota

import (
	"context"
	"errors"
	"fmt"
	iotago "github.com/iotaledger/iota.go/v2"
	"iotasync/common"
	"iotasync/common/conf"
	"iotasync/common/log"
	dao "iotasync/models/po/iota"
	"time"
)

type Scanner struct {
	nodeAPI *iotago.NodeHTTPAPIClient
	conf    conf.SyncConfig
}

//nodeAPI := iotago.NewNodeHTTPAPIClient("https://api.lb-0.h.chrysalis-devnet.iota.cafe:443")
func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		nodeAPI: iotago.NewNodeHTTPAPIClient(node.Url),
		conf:    conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	//dao.TokenTxRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(58044277)

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	info, err := s.nodeAPI.Info(context.Background())
	if err != nil {
		return 0, err
	}
	return int64(info.ConfirmedMilestoneIndex), nil
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
retryGetBlock:
	block, err := s.nodeAPI.MilestoneByIndex(context.Background(), uint32(height))
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlock
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	if has, err := dao.BlockHashExist(block.MessageID); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Index, block.MessageID, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{

			Height: int64(block.Index),
			Hash:   block.MessageID,
			//	Previousblockhash: block.ParentHash,
			Nextblockhash: "",
			//Transactions:      len(rs.ConsumedOutputs),
			Confirmations: bestHeight - height + 1, //block.Confirmations
			Timestamp:     time.Unix(block.Time, 0),
		},
	}

	rs, err := s.nodeAPI.MilestoneUTXOChangesByIndex(context.Background(), uint32(height))
	if err != nil {
		return nil, err
	}

	txids := make(map[string]bool, 0)
	for _, outputIDHex := range rs.CreatedOutputs {
		utxoInput, err := iotago.OutputIDHex(outputIDHex).AsUTXOInput()
		if err != nil {
			log.Error(err)
			continue
		}
		outputRes, err := s.nodeAPI.OutputByID(context.Background(), utxoInput.ID())
		if err != nil {
			log.Error(err)
			continue
		}
		txids[outputRes.MessageID] = true
	}
	//处理区块内的交易
	//if len(rs.ConsumedOutputs) > 0 {
	//	if txInfo, err := s.parseBlockRawTX(rs.CreatedOutputs, rs.ConsumedOutputs, block.MessageID, height, block.Time); err != nil {
	//		log.Info(block.MessageID, "parseBlockRawTX err：", err.Error())
	//	} else if txInfo != nil {
	//		task.TxInfos = append(task.TxInfos, txInfo)
	//	}
	//}
	for txId, _ := range txids {
		if txInfo, err := s.parseBlockRawTXV1(txId, block.MessageID, height, block.Time); err != nil {
			log.Info(block.MessageID, "parseBlockRawTX err：", err.Error())
		} else if txInfo != nil {
			task.TxInfos = append(task.TxInfos, txInfo)
		}
	}
	return task, nil
}
func (s *Scanner) parseBlockRawTXV1(txid string, blockhash string, height, timstap int64) (txInfo *TxInfo, err error) {

	blockTx := &dao.BlockTx{
		Txid:      txid,
		Blockhash: blockhash,
		Height:    height,
		//Vincount:   len(ConsumeOutPuts),
		//Voutcount:  len(CreatedOutputs),
		Createtime: time.Now(),
		Timestamp:  time.Unix(timstap, 0),
	}
	txInfo = &TxInfo{Tx: blockTx}
	messageId, err := iotago.MessageIDFromHexString(txid)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	resp, err := s.nodeAPI.MessageMetadataByMessageID(context.Background(), messageId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	
	if *resp.LedgerInclusionState != "included" {
		return nil, errors.New("conflicting transaction")
	}
	//ms, err := nodeAPI.MessageJSONByMessageID(context.Background(), messageId)
	//fmt.Printf("%+v\n", ms.Payload)

	message, err := s.nodeAPI.MessageByMessageID(context.Background(), messageId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//fmt.Printf("%+v\n", message.Payload)

	tx, ok := message.Payload.(*iotago.Transaction)
	if !ok {
		return nil, errors.New("not transation")
	}
	ess := tx.Essence.(*iotago.TransactionEssence)
	Inputs := ess.Inputs
	Outputs := ess.Outputs
	blockTx.Vincount = len(Inputs)
	blockTx.Voutcount = len(Outputs)
	txInfo = &TxInfo{Tx: blockTx}
	for _, o := range Outputs {
		output, ok := o.(iotago.Output)
		if ok {
			tar, err := output.Target()
			if err != nil {
				log.Error(err)
				continue
			}
			addr := fmt.Sprintf("%v", tar)
			edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)
			value, err := output.Deposit()
			if err != nil {
				log.Error(err)
				continue
			}
			txOut := &dao.BlockTxVout{
				Height:     height,
				Address:    edAddr.Bech32(iotago.PrefixMainnet),
				Value:      fmt.Sprintf("%v", value),
				Txid:       txid,
				MessageId:  txid,
				Createtime: time.Now(),
			}
			txInfo.Vouts = append(txInfo.Vouts, txOut)
		}
	}
	//
	for _, o := range Inputs {
		h, ok := o.(*iotago.UTXOInput)
		if ok {
			//fmt.Println(h.TransactionOutputIndex)
			//fmt.Println(h.ID().ToHex())
			utxoInput, err := iotago.OutputIDHex(h.ID().ToHex()).AsUTXOInput()
			if err != nil {
				log.Error(err)
				continue
			}
			outputRes, err := s.nodeAPI.OutputByID(context.Background(), utxoInput.ID())
			output, err := outputRes.Output()
			tar, err := output.Target()
			if err != nil {
				log.Error(err)
				continue
			}
			addr := fmt.Sprintf("%v", tar)
			edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)
			if err != nil {
				log.Error(err)
				continue
			}
			value, err := output.Deposit()
			if err != nil {
				log.Error(err)
				continue
			}
			txVin := &dao.BlockTxVin{
				Height:      height,
				Address:     edAddr.Bech32(iotago.PrefixMainnet),
				Value:       fmt.Sprintf("%v", value),
				Txid:        outputRes.TransactionID,
				OutputIndex: outputRes.OutputIndex,
				LedgerIndex: outputRes.LedgerIndex,
				MessageId:   outputRes.MessageID,
				Spent:       outputRes.Spent,
				Createtime:  time.Now(),
			}
			txInfo.Vins = append(txInfo.Vins, txVin)
		}
	}

	//fmt.Println("txinfo Vins:", len(txInfo.Vins))
	return txInfo, nil
}

//解析交易
func (s *Scanner) parseBlockRawTX(CreatedOutputs, ConsumeOutPuts []string, blockhash string, height, timstap int64) (txInfo *TxInfo, err error) {

	blockTx := &dao.BlockTx{
		//Txid:       txId,
		Blockhash:  blockhash,
		Height:     height,
		Vincount:   len(ConsumeOutPuts),
		Voutcount:  len(CreatedOutputs),
		Createtime: time.Now(),
		Timestamp:  time.Unix(timstap, 0),
	}
	txInfo = &TxInfo{Tx: blockTx}

	for _, outputIDHex := range CreatedOutputs {
		utxoInput, err := iotago.OutputIDHex(outputIDHex).AsUTXOInput()
		if err != nil {
			log.Error(err)
			continue
		}
		outputRes, err := s.nodeAPI.OutputByID(context.Background(), utxoInput.ID())
		if err != nil {
			log.Error(err)
			continue
		}
		output, err := outputRes.Output()
		if err != nil {
			log.Error(err)
			continue
		}

		tar, err := output.Target()
		if err != nil {
			log.Error(err)
			continue
		}
		addr := fmt.Sprintf("%v", tar)
		edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)

		value, err := output.Deposit()
		if err != nil {
			log.Error(err)
			continue
		}
		txOut := &dao.BlockTxVout{
			Height:      height,
			Address:     edAddr.Bech32(iotago.PrefixMainnet),
			Value:       fmt.Sprintf("%v", value),
			Txid:        outputRes.TransactionID,
			OutputIndex: outputRes.OutputIndex,
			LedgerIndex: outputRes.LedgerIndex,
			MessageId:   outputRes.MessageID,
			Spent:       outputRes.Spent,
			Createtime:  time.Now(),
		}
		txInfo.Vouts = append(txInfo.Vouts, txOut)
	}

	for _, outputIDHex := range ConsumeOutPuts {
		utxoInput, err := iotago.OutputIDHex(outputIDHex).AsUTXOInput()
		if err != nil {
			log.Error(err)
			continue
		}
		outputRes, err := s.nodeAPI.OutputByID(context.Background(), utxoInput.ID())
		if err != nil {
			log.Error(err)
			continue
		}
		output, err := outputRes.Output()
		if err != nil {
			log.Error(err)
			continue
		}

		tar, err := output.Target()
		if err != nil {
			log.Error(err)
			continue
		}
		addr := fmt.Sprintf("%v", tar)
		edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)
		if err != nil {
			log.Error(err)
			continue
		}
		value, err := output.Deposit()
		if err != nil {
			log.Error(err)
			continue
		}
		txVin := &dao.BlockTxVin{
			Height:      height,
			Address:     edAddr.Bech32(iotago.PrefixMainnet),
			Value:       fmt.Sprintf("%v", value),
			Txid:        outputRes.TransactionID,
			OutputIndex: outputRes.OutputIndex,
			LedgerIndex: outputRes.LedgerIndex,
			MessageId:   outputRes.MessageID,
			Spent:       outputRes.Spent,
			Createtime:  time.Now(),
		}
		txInfo.Vins = append(txInfo.Vins, txVin)
	}

	return txInfo, nil
}
