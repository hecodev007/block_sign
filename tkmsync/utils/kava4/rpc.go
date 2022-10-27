package kava4

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/shopspring/decimal"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// custom tx codec
func makeCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	cdc.Seal()
	return cdc
}

type Client struct {
	client rpcclient.Client
	cdc    *codec.Codec
}

func NewClient(url string) *Client {
	c := &Client{
		client: rpcclient.NewHTTP(url, "/websocket"),
		cdc:    makeCodec(),
	}
	return c
}
func (c *Client) GetLastBlockHeight() (int64, error) {
	status, err := c.client.Status()
	if err != nil {
		return -1, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}
func (c *Client) GetBlockByHeight(height int64) (*Block, error) {
	proxy, err := c.client.Block(&height)
	if err != nil {
		return nil, err
	}
	block := Block{
		Hash:       proxy.Block.Header.Hash().String(),
		Height:     proxy.Block.Height,
		ParentHash: proxy.Block.LastBlockID.Hash.String(),
		ChainID:    proxy.Block.ChainID,
		Timestamp:  proxy.Block.Time,
	}
	for _, tx := range proxy.Block.Txs {
		block.Transactions = append(block.Transactions, fmt.Sprintf("%X", tx.Hash()))
	}
	return &block, nil
}

//func (c *Client) GetTransactionByHash(txid string, blockTime time.Time) (*Transaction, error) {
//	hash, err := hex.DecodeString(txid)
//	if err != nil {
//		return nil, fmt.Errorf("DecodeString %v", err)
//	}
//	resTx, err := c.client.Tx(hash, true)
//	if err != nil {
//		log.Printf("get Tx : %v ", resTx)
//		return nil, fmt.Errorf("tx %v", err)
//	}
//	var tx auth.StdTx
//	err = c.cdc.UnmarshalBinaryLengthPrefixed(resTx.Tx, &tx)
//	if err != nil {
//		return nil, err
//	}
//	logs, err := sdk.ParseABCILogs(resTx.TxResult.Log)
//	if err != nil {
//		return nil, err
//	}
//	res := &Transaction{
//		Hash:        resTx.Hash.String(),
//		RawLogs:     resTx.TxResult.Log,
//		GasWanted:   resTx.TxResult.GasWanted,
//		GasUsed:     resTx.TxResult.GasUsed,
//		BlockHeight: resTx.Height,
//		Timestamp:   blockTime,
//		Type:        "auth/StdTx",
//		Memo:        tx.GetMemo(),
//		Fee:         tx.Fee.Amount.String(),
//		GasPrice:    tx.Fee.Gas,
//	}
//	msgs := tx.GetMsgs()
//	for i, m := range msgs {
//		switch m.Type() {
//		case TypeMsgSend:
//			msg := TxMsg{
//				Index: i,
//				Type:  m.Type(),
//			}
//			if len(logs) > i {
//				msg.Log = logs[i].Log
//				//msg.Success = logs[i].Success
//			}
//			if m1, ok := m.(bank.MsgSend); ok {
//				msg.From = m1.FromAddress.String()
//				msg.To = m1.ToAddress.String()
//				msg.Amount = m1.Amount.String()
//			}
//			res.TxMsgs = append(res.TxMsgs, msg)
//		case TypeMsgDelegate:
//			msg := TxMsg{
//				Index: i,
//				Type:  m.Type(),
//			}
//			if len(logs) > i {
//				msg.Log = logs[i].Log
//				msg.Success = logs[i].Success
//			}
//			if m1, ok := m.(staking.MsgDelegate); ok {
//				msg.From = m1.DelegatorAddress.String()
//				msg.To = m1.ValidatorAddress.String()
//				msg.Amount = m1.Amount.String()
//			}
//			res.TxMsgs = append(res.TxMsgs, msg)
//		case TypeMultiSend:
//		case TypeMsgDeposit:
//		default:
//			log.Printf("don't supported type %s", m.Type())
//		}
//	}
//	return res, nil
//}

func GetKavaNum(coin string) decimal.Decimal {
	feecoins, err := sdk.ParseCoins(coin)
	if err != nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(feecoins.AmountOf(MainnetDenom).BigInt(), 0).Div(decimal.New(1, 6))
}
