package atom

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/shopspring/decimal"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"strings"
	"telosDataServer/common/log"
	"telosDataServer/utils"
	"time"
)

const (
	BroadcastBlock       = "block"
	BroadcastSync        = "sync"
	BroadcastAsync       = "async"
	MainnetDenom         = "uatom"
	MainChainID          = "cosmoshub-2"
	DefaultGasAdjustment = 1.5
	DefaultGasLimit      = 300000
)

type Client struct {
	client rpcclient.Client
	Codec  *codec.Codec
	//AccDecoder    auth.AccountDecoder
	AccountStore  string
	BroadcastMode string
	Height        int64
	TrustNode     bool
}

func NewClient(url string) *Client {
	c := &Client{
		client:        rpcclient.NewHTTP(url, "/websocket"),
		Codec:         makeDefaultCodec(),
		AccountStore:  auth.StoreKey,
		BroadcastMode: BroadcastSync,
	}

	//c.WithAccountDecoder(c.Codec)
	return c
}

func GetAtomNum(coin string) decimal.Decimal {
	feecoins, err := sdk.ParseCoins(coin)
	if err != nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(feecoins.AmountOf(MainnetDenom).BigInt(), 0).Div(decimal.New(1, 6))
}

func (c *Client) GetChainID() (string, error) {
	var height int64 = 1000
	block, err := c.client.Block(&height)
	if err != nil {
		return "", err
	}
	return block.Block.ChainID, nil
}

// EnsureAccountExists ensures that an account exists for a given context. An
// error is returned if it does not.
func (c *Client) EnsureAccountExists(address string) bool {
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}
	_, err = c.queryAccount(accAddr)
	if err != nil {
		return false
	}
	return true
}

func (c *Client) GetLatestBlockHeight() (int64, error) {
	status, err := c.client.Status()
	if err != nil {
		return -1, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// custom tx codec
func makeDefaultCodec() *codec.Codec {
	var cdc = codec.New()

	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	crisis.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc.Seal()
}

func (c *Client) GetNodeStatus() (*ctypes.ResultStatus, error) {
	// get the node
	return c.client.Status()
}

// queryAccount queries an account using custom query endpoint of auth module
// returns an error if result is `null` otherwise account data
func (c *Client) queryAccount(addr sdk.AccAddress) ([]byte, error) {
	bz, err := c.Codec.MarshalJSON(auth.NewQueryAccountParams(addr))
	if err != nil {
		log.Infof("MarshalJSON err : %v", err)
		return nil, err
	}
	route := fmt.Sprintf("custom/%s/%s", c.AccountStore, auth.QueryAccount)
	res, err := c.query(route, bz)
	if err != nil {
		log.Infof("QueryWithData err : %v", err)
		return nil, err
	}
	return res, nil
}

func (c *Client) GetBlockByHeight(height int64) (*Block, error) {
	res, err := c.client.Block(&height)
	if err != nil {
		return nil, err
	}
	//log.Infof("block %v", res)
	return toBlock(res), nil
}

func toBlock(proxy *ctypes.ResultBlock) *Block {
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
	return &block
}

func (c *Client) ParseTxResponse(txres sdk.TxResponse) (*Transaction, error) {
	proxy := &ProxyTransaction{}
	output, err := c.Codec.MarshalJSON(txres)
	if err != nil {
		return nil, fmt.Errorf("MarshalJSON %v", err)
	}

	log.Infof("output ----- %s", string(output))
	err = c.Codec.UnmarshalJSON(output, proxy)
	if err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	log.Infof("proxy ----- %v", proxy)
	tx := &Transaction{
		Hash:    proxy.TxHash,
		RawLogs: proxy.RawLog,
		Type:    proxy.Tx.Type,
	}
	tx.Timestamp, _ = time.Parse(time.RFC3339, proxy.Timestamp)
	tx.GasWanted, _ = utils.ParseInt64(proxy.GasWanted)
	tx.GasUsed, _ = utils.ParseInt64(proxy.GasUsed)
	tx.BlockHeight, _ = utils.ParseInt64(proxy.Height)

	if tx.Type == "auth/StdTx" {
		tx.Fee = proxy.Tx.Value.Fee.Amount.String()
		tx.Memo = proxy.Tx.Value.Memo

		msgs := proxy.Tx.Value.GetMsgs()
		log.Infof("get msg num : %d ", len(msgs))
		if len(msgs) > 0 {
			for i, tmp := range msgs {
				log.Infof("get msg type : %s ", tmp.Type())
				switch tmp.Type() {
				case TypeMsgDelegate:
					txmsg := TxMsg{
						Index: i,
						Type:  tmp.Type(),
					}
					if len(proxy.Logs) > i {
						txmsg.Log = proxy.Logs[i].Log
						txmsg.Success = proxy.Logs[i].Success
					}

					if msg, ok := tmp.(staking.MsgDelegate); ok {
						txmsg.From = msg.DelegatorAddress.String()
						txmsg.To = msg.ValidatorAddress.String()
						txmsg.Amount = msg.Amount.String()
					}
					tx.TxMsgs = append(tx.TxMsgs, txmsg)
					break
				case TypeMsgSend:
					txmsg := TxMsg{
						Index: i,
						Type:  tmp.Type(),
					}
					if len(proxy.Logs) > i {
						txmsg.Log = proxy.Logs[i].Log
						txmsg.Success = proxy.Logs[i].Success
					}
					if msg, ok := tmp.(bank.MsgSend); ok {
						txmsg.From = msg.FromAddress.String()
						txmsg.To = msg.ToAddress.String()
						txmsg.Amount = msg.Amount.String()
					}
					tx.TxMsgs = append(tx.TxMsgs, txmsg)
					break
				case TypeMultiSend:
					break
				case TypeMsgDeposit:
					break
				case TypeMsgWithdrawDelegationReward:
					break
				}
			}
		}
	} else {
		return nil, fmt.Errorf("don't support tx type : %s", tx.Type)
	}

	return tx, nil
}

func (c *Client) GetTx(txid string, blocktime time.Time) (*Transaction, error) {
	hash, err := hex.DecodeString(txid)
	if err != nil {
		return nil, fmt.Errorf("DecodeString %v", err)
	}
	res, err := c.client.Tx(hash, true)
	if err != nil {
		return nil, fmt.Errorf("tx %v", err)
	}
	log.Infof("get Tx : %v ", res)
	txres, err := c.formatTxResult(res, blocktime)
	if err != nil {
		return nil, fmt.Errorf("formatTxResult %v", err)
	}

	return c.ParseTxResponse(txres)
}

func (c *Client) SearchTxs(tags []string, page, limit int) ([]sdk.TxResponse, error) {
	if len(tags) == 0 {
		return nil, errors.New("must declare at least one tag to search")
	}

	if page <= 0 {
		return nil, errors.New("page must greater than 0")
	}

	if limit <= 0 {
		return nil, errors.New("limit must greater than 0")
	}

	query := strings.Join(tags, " AND ")
	resTxs, err := c.client.TxSearch(query, false, page, limit)
	if err != nil {
		return nil, err
	}

	resBlocks, err := c.getBlocksForTxResults(resTxs.Txs)
	if err != nil {
		return nil, err
	}

	txs, err := c.formatTxResults(resTxs.Txs, resBlocks)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (c *Client) BroadcastTxCommit(txBytes []byte) (sdk.TxResponse, error) {
	res, err := c.client.BroadcastTxCommit(txBytes)
	if err != nil {
		return sdk.NewResponseFormatBroadcastTxCommit(res), err
	}
	if !res.CheckTx.IsOK() {
		return sdk.NewResponseFormatBroadcastTxCommit(res), fmt.Errorf(res.CheckTx.Log)
	}
	if !res.DeliverTx.IsOK() {
		return sdk.NewResponseFormatBroadcastTxCommit(res), fmt.Errorf(res.DeliverTx.Log)
	}
	return sdk.NewResponseFormatBroadcastTxCommit(res), nil
}

func (c *Client) BroadcastTxSync(txBytes []byte) (sdk.TxResponse, error) {
	res, err := c.client.BroadcastTxSync(txBytes)
	return sdk.NewResponseFormatBroadcastTx(res), err
}

func (c *Client) BroadcastTxAsync(txBytes []byte) (sdk.TxResponse, error) {
	res, err := c.client.BroadcastTxAsync(txBytes)
	return sdk.NewResponseFormatBroadcastTx(res), err
}

func (c *Client) BroadcastTx(txBytes []byte) (res sdk.TxResponse, err error) {
	switch c.BroadcastMode {
	case BroadcastSync:
		res, err = c.BroadcastTxSync(txBytes)
	case BroadcastAsync:
		res, err = c.BroadcastTxAsync(txBytes)
	case BroadcastBlock:
		res, err = c.BroadcastTxCommit(txBytes)
	default:
		return sdk.TxResponse{}, fmt.Errorf("unsupported return type %s; supported types: sync, async, block", c.BroadcastMode)
	}
	return res, err
}

//func (c *Client) GetAccount(address string) (auth.Account, error) {
//	accAddr, err := sdk.AccAddressFromBech32(address)
//	if err != nil {
//		return nil, err
//	}
//
//	res, err := c.queryAccount(accAddr)
//	if err != nil {
//		return nil, fmt.Errorf("query account err : %v", err)
//	}
//
//	var account auth.Account
//	if err := c.Codec.UnmarshalJSON(res, &account); err != nil {
//		return nil, fmt.Errorf("unmarshl Json err :%v, res : %s", err, string(res))
//	}
//
//	return account, nil
//}

func (c *Client) EstimatedFee(bz []byte) (uint64, error) {
	route := "/app/simulate"
	rawRes, err := c.query(route, bz)
	if err != nil {
		log.Infof("QueryWithData err : %v", err)
		return 0, err
	}

	var simulationResult sdk.Result
	if err := c.Codec.UnmarshalBinaryLengthPrefixed(rawRes, &simulationResult); err != nil {
		return 0, err
	}

	return uint64(DefaultGasAdjustment * float64(simulationResult.GasUsed)), nil
}

// QueryWithData performs a query to a Tendermint node with the provided path
// and a data payload. It returns the result and height of the query upon success
// or an error if the query fails.
//func (c *Client) QueryWithData(path string, data []byte) ([]byte, error) {
//	return c.query(path, data)
//}
func (c *Client) query(path string, key []byte) (res []byte, err error) {

	opts := rpcclient.ABCIQueryOptions{
		Height: c.Height,
		Prove:  !c.TrustNode,
	}
	result, err := c.client.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, fmt.Errorf("ABCIQueryWithOptions err : %v,result : %v", err, result)
	}
	resp := result.Response
	if !resp.IsOK() {
		return res, errors.New(resp.Log)
	}
	// data from trusted node or subspace query doesn't need verification
	if c.TrustNode {
		return resp.Value, nil
	}
	return resp.Value, nil
}

func (c *Client) parseTx(txBytes []byte) (sdk.Tx, error) {
	var tx auth.StdTx
	err := c.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *Client) formatTxResult(resTx *ctypes.ResultTx, blockTime time.Time) (sdk.TxResponse, error) {
	tx, err := c.parseTx(resTx.Tx)
	if err != nil {
		return sdk.TxResponse{}, fmt.Errorf("parseTx %v", err)
	}
	return sdk.NewResponseResultTx(resTx, tx, blockTime.Format(time.RFC3339)), nil
}

func (c *Client) getBlocksForTxResults(resTxs []*ctypes.ResultTx) (map[int64]*ctypes.ResultBlock, error) {
	resBlocks := make(map[int64]*ctypes.ResultBlock)
	for _, resTx := range resTxs {
		if _, ok := resBlocks[resTx.Height]; !ok {
			resBlock, err := c.client.Block(&resTx.Height)
			if err != nil {
				return nil, err
			}
			resBlocks[resTx.Height] = resBlock
		}
	}
	return resBlocks, nil
}

// formatTxResults parses the indexed txs into a slice of TxResponse objects.
func (c *Client) formatTxResults(resTxs []*ctypes.ResultTx, resBlocks map[int64]*ctypes.ResultBlock) ([]sdk.TxResponse, error) {
	var err error
	out := make([]sdk.TxResponse, len(resTxs))
	for i := range resTxs {
		out[i], err = c.formatTxResult(resTxs[i], resBlocks[resTxs[i].Height].Block.Time)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
