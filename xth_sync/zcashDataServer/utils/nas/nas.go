package nas

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/nebulasio/go-nebulas/rpc/pb"
//
//	"github.com/shopspring/decimal"
//	"dataserver/utils"
//)
//
//
//func toNodeState(proxy *rpcpb.GetNebStateResponse) *NodeState {
//	state := &NodeState{
//		TailHash:     proxy.Tail,
//		LibHash:      proxy.Lib,
//		Height:       proxy.Height,
//		Synchronized: proxy.Synchronized,
//		Version:      proxy.Version,
//	}
//	return state
//}
//
//func toEvent(proxy *rpcpb.GetNebStateResponse) *Event {
//	evt := &Event{}
//	return evt
//}
//
//func toTransaction(proxy *rpcpb.TransactionResponse) *Transaction {
//	tx := &Transaction{
//		Hash:            proxy.Hash,
//		BlockHeight:     proxy.BlockHeight,
//		From:            proxy.From,
//		To:              proxy.To,
//		Nonce:           proxy.Nonce,
//		Timestamp:       proxy.Timestamp,
//		Type:            proxy.Type,
//		Data:            string(proxy.Data),
//		ContractAddress: proxy.ContractAddress,
//		Status:          int(proxy.Status),
//		ExecuteError:    proxy.ExecuteError,
//		ExecuteResult:   proxy.ExecuteResult,
//	}
//
//	tx.Value, _ = utils.ParseBigInt(proxy.Value)
//	tx.GasPrice, _ = utils.ParseBigInt(proxy.GasPrice)
//	tx.GasUsed, _ = utils.ParseInt64(proxy.GasUsed)
//	return tx
//}
//
//func toBlock(proxy *rpcpb.BlockResponse) *Block {
//	block := &Block{
//		Hash:       proxy.Hash,
//		ParentHash: proxy.ParentHash,
//		Height:     proxy.Height,
//		Nonce:      proxy.Nonce,
//		Coinbase:   proxy.Coinbase,
//		Timestamp:  proxy.Timestamp,
//		IsFinality: proxy.IsFinality,
//	}
//
//	block.Transactions = make([]*Transaction, len(proxy.Transactions))
//	for i := range proxy.Transactions {
//		block.Transactions[i] = toTransaction(proxy.Transactions[i])
//	}
//	return block
//}
//
