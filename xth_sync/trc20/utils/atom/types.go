package atom

import (
	"encoding/hex"

	"lunasync/utils"
	"time"

	"github.com/terra-money/core/app"

	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	btypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"
	"github.com/tendermint/tendermint/types"
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("kava", "kava"+sdk.PrefixPublic)
}

const (
	TypeMsgSend                     = "send"
	TypeMsgDelegate                 = "delegate"
	TypeMultiSend                   = "multisend"
	TypeMsgDeposit                  = "deposit"
	TypeMsgWithdrawDelegationReward = "withdraw_delegator_reward"
)

type BlockReponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  Block       `json:"result"`
	Error   struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
}
type Block struct {
	BlockId BlockID `json:"block_id"`
	Block   struct {
		Header ResponseBlockHeader `json:"header"`
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
	//Hash         string    `json:"hash"`
	//ParentHash   string    `json:"parent_hash"`
	//Height       int64     `json:"height"`
	//ChainID      string    `json:"chain_id"`
	//Timestamp    time.Time `json:"timestamp"`
	//Transactions []string  `json:"transactions"`
}

// BlockID defines the unique ID of a block as its Hash and its PartSetHeader
type BlockID struct {
	Hash        string `json:"hash"`
	PartsHeader struct {
		Total int64  `json:"total"`
		Hash  string `json:"hash"`
	} `json:"parts"`
}

type ResponseBlockHeader struct {
	// basic block info
	ChainID string          `json:"chain_id"`
	Height  decimal.Decimal `json:"height"`
	Time    time.Time       `json:"time"`
	NumTxs  string          `json:"num_txs"`
	//TotalTxs string    `json:"total_txs"`

	// prev block info
	LastBlockID BlockID `json:"last_block_id"`

	// hashes of block data
	//LastCommitHash string `json:"last_commit_hash"` // commit from validators from the last block
	//DataHash       string `json:"data_hash"`        // transactions
	//
	//// hashes from the app output from the prev block
	//ValidatorsHash     string `json:"validators_hash"`      // validators for the current block
	//NextValidatorsHash string `json:"next_validators_hash"` // validators for the next block
	//ConsensusHash      string `json:"consensus_hash"`       // consensus params for current block
	//AppHash            string `json:"app_hash"`             // state after txs from the previous block
	//LastResultsHash    string `json:"last_results_hash"`    // root hash of all results from the txs from the previous block
	//
	//// consensus info
	//EvidenceHash    string `json:"evidence_hash"`    // evidence included in the block
	//ProposerAddress string `json:"proposer_address"` // original proposer of the block
}

// Single block (with meta)
type BlockResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  BlockResult `json:"result"`
}
type BlockResult struct {
	BlockID BlockID `json:"block_id"` // the block hash and partsethash
	Block   struct {
		Header ResponseBlockHeader `json:"header"`
		Data   struct {
			Txs types.Txs `json:"txs"`
		} `json:"data"`
	} `json:"block"`
}

func (proxy *BlockResult) toBlock() *Block {

	block := Block{
		BlockId: proxy.BlockID,
		Block:   proxy.Block,
	}
	return &block
}

type TxReponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  Result      `json:"result"`
	Error   struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
}
type Result struct {
	Height   string   `json:"height"`
	TxHash   string   `json:"hash"`
	TxResult TxResult `json:"tx_result"`
	//RawLog    string        `json:"raw_log,omitempty"`
	Tx types.Tx `json:"tx"`
}
type TxResult struct {
	Code      int64         `json:"code"`
	Logs      string        `json:"log,omitempty"`
	Events    []interface{} `json:"events"`
	GasWanted string        `json:"gas_wanted,omitempty"`
	GasUsed   string        `json:"gas_used,omitempty"`
}

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      sdk.Coins `json:"amount"`
	UnlockTime  string    `json:"unlock_time"`
}

type MsgDelegate struct {
	DelegatorAddress string    `json:"delegator_address"`
	ValidatorAddress string    `json:"validator_address"`
	Amount           sdk.Coin  `json:"amount"`
	UnlockTime       time.Time `json:"unlock_time"`
}

//Msgs []struct {
//FromAddress string    `json:"from_address"`
//ToAddress   string    `json:"to_address"`
//Amount      sdk.Coins `json:"amount"`
//}
type stdTx struct {
	Msgs interface{} `json:"msg"`
	Fee  struct {
		Amount sdk.Coins `json:"amount"`
		Gas    int64     `json:"gas"`
	} `json:"fee"`
	Memo string `json:"emo"`
}

func (proxy *Result) ToTx() (txs []*Transaction, err error) {
	if proxy.TxResult.Code != 0 || len(proxy.TxResult.Events) == 0 {
		return nil, errors.New("失败的交易")
	}

	encodingConfig := app.MakeEncodingConfig()
	decodertx, err := encodingConfig.TxConfig.TxDecoder()(proxy.Tx)
	if err != nil {
		return nil, err
	}

	feeTx, err := encodingConfig.TxConfig.WrapTxBuilder(decodertx)
	//_, _ = feeTx, msgs
	signTx := feeTx.GetTx()
	feeStr := signTx.GetFee()
	memo := signTx.GetMemo()
	msgs := signTx.GetMsgs()
	_, _, _ = feeStr, memo, msgs

	//log.Println(xutils.String(msgs))
	//log.Println(xutils.String(feeStr))
	//msg, ok := msgs[0].(*btypes.MsgSend)
	sendIndex := -1
	for k, _ := range msgs {
		if _, ok := msgs[0].(*btypes.MsgSend); ok {
			sendIndex = k
			break
		}
	}
	if sendIndex < 0 {
		return nil, errors.New("非转账交易")
	}
	tmptx := Transaction{
		Hash:    hex.EncodeToString(proxy.Tx.Hash()),
		Memo:    memo,
		Success: true,
		Type:    "send",
	}
	tmptx.Timestamp = time.Now()
	tmptx.GasWanted, _ = utils.ParseInt64(proxy.TxResult.GasWanted)
	tmptx.GasUsed, _ = utils.ParseInt64(proxy.TxResult.GasUsed)
	tmptx.BlockHeight, _ = utils.ParseInt64(proxy.Height)
	tmptx.Fee, _ = AtomToInt(feeStr.String())
	for k, _ := range msgs {
		if _, ok := msgs[k].(*btypes.MsgSend); !ok {
			break
		}
		tx := new(Transaction)
		*tx = tmptx
		tx.From = msgs[k].(*btypes.MsgSend).FromAddress
		tx.To = msgs[k].(*btypes.MsgSend).ToAddress
		tx.Value = msgs[sendIndex].(*btypes.MsgSend).Amount[0].Amount.Int64()
		tx.Token = msgs[sendIndex].(*btypes.MsgSend).Amount[0].Denom
		txs = append(txs, tx)
	}

	return txs, nil
	//t.Log(xutils.String(msgs))
	//codec := app.MakeCodec()
	//tmptx, err := auth.DefaultTxDecoder(codec)(proxy.Tx)
	//if err != nil {
	//	return nil, err
	//}
	//msgs := tmptx.GetMsgs()
	//sendIndex := -1
	//for k, _ := range msgs {
	//	if msgs[k].Type() == "send" {
	//		sendIndex = k
	//		break
	//	}
	//}
	//if sendIndex < 0 {
	//	return nil, errors.New("非转账交易")
	//}
	//
	//txBytes, _ := json.Marshal(tmptx)
	////fmt.Println(string(txBytes))
	//stdtx := new(stdTx)
	//err = json.Unmarshal(txBytes, stdtx)
	//if err != nil {
	//	return nil, errors.New("交易格式错误,需要升级")
	//}
	//
	//tx = &Transaction{
	//	Hash:    hex.EncodeToString(proxy.Tx.Hash()),
	//	Memo:    stdtx.Memo,
	//	Success: true,
	//	Type:    "send",
	//}
	//tx.Timestamp = time.Now()
	//tx.GasWanted, _ = utils.ParseInt64(proxy.TxResult.GasWanted)
	//tx.GasUsed, _ = utils.ParseInt64(proxy.TxResult.GasUsed)
	//tx.BlockHeight, _ = utils.ParseInt64(proxy.Height)
	////fmt.Println(stdtx.Fee.Amount[0].Amount, stdtx.Fee.Amount.String())
	//Fee, err := AtomToInt(stdtx.Fee.Amount.String())
	//if err != nil {
	//	panic(tx.Hash + err.Error() + "  " + stdtx.Fee.Amount.String())
	//}
	//tx.Fee = Fee
	//msgbytes, _ := json.Marshal(msgs[sendIndex])
	//var msg MsgSend
	//err = json.Unmarshal(msgbytes, &msg)
	//if err != nil {
	//	return nil, errors.New("msg交易格式错误,需要升级" + err.Error())
	//}
	//tx.From = msg.FromAddress
	//tx.To = msg.ToAddress
	//tx.Value, _ = AtomToInt(msg.Amount.String())

	//return tx, nil
	return nil, err
}

type Transaction struct {
	Hash        string `json:"hash"`
	BlockHeight int64  `json:"block_height"`
	//BlockHash   string    `json:"block_hash"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Value      int64     `json:"value"`
	Fee        int64     `json:"fee,omitempty"`
	Memo       string    `json:"memo,omitempty"`
	Success    bool      `json:"success"`
	Token      string    `json:"token"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	UnlockTime string    `json:"unlock_time"`

	GasUsed   int64   `json:"gas_used,omitempty" `
	GasWanted int64   `json:"gas_wanted,omitempty"`
	GasPrice  string  `json:"gas_price,omitempty"`
	RawLogs   string  `json:"raw_logs,omitempty"`
	TxMsgs    []TxMsg `json:"tx_msgs"`
}

type TxMsg struct {
	Index      int    `json:"index"`
	Type       string `json:"type"`
	Success    bool   `json:"success"`
	Log        string `json:"log"`
	From       string `json:"from"`
	To         string `json:"to"`
	Amount     string `json:"amount"`
	UnlockTime string `json:"unlock_time"`
}

type MessageLog struct {
	MsgIndex int64    `json:"msg_index"`
	Success  bool     `json:"success"`
	Log      string   `json:"log"`
	Events   []*Event `json:"events"`
}
type Event struct {
	Type       string       `json:"type"`
	Attributes []*Attribute `json:"attributes"`
}
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type ProxyTx struct {
	Type string `json:"type"`
	//Value auth.StdTx `json:"value"`
}

type ProxyTransaction struct {
	Height    string        `json:"height"`
	TxHash    string        `json:"txhash"`
	RawLog    string        `json:"raw_log,omitempty"`
	Logs      []*MessageLog `json:"logs,omitempty"`
	GasWanted string        `json:"gas_wanted,omitempty"`
	GasUsed   string        `json:"gas_used,omitempty"`
	Tx        ProxyTx       `json:"tx"`
	Timestamp string        `json:"timestamp"`
}

func GetCoinNum(coin, Denom string) decimal.Decimal {
	//feecoins, err := sdk.ParseCoins(coin)
	//if err != nil {
	return decimal.Zero
	//}
	//return decimal.NewFromBigInt(feecoins.AmountOf(Denom).BigInt(), 0)
}
