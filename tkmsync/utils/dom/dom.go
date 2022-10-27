package dom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"rsksync/utils"
	"unsafe"
)

const BlockLatest = "latest"
const WEI = 8

type hexInt int

type hexBig big.Int

type Syncing struct {
	IsSyncing     bool
	StartingBlock int
	CurrentBlock  int
	HighestBlock  int
}

// T - input transaction object
type T struct {
	From   string
	To     string
	Fee    *big.Int
	Amount *big.Int
	Data   string
	Nonce  int
}

type Transaction struct {
	Tx struct {
		Execer  string `json:"execer"` // coins 主币
		Payload struct {
			Transfer struct {
				Cointoken string      `json:"cointoken"`
				Amount    string      `json:"amount"`
				Note      interface{} `json:"note"`
				To        string      `json:"to"`
			} `json:"transfer"`
			Ty int `json:"ty"`
		} `json:"payload"`
		Rawpayload string `json:"rawPayload"`
		Signature  struct {
			Ty        int    `json:"ty"`
			Pubkey    string `json:"pubkey"`
			Signature string `json:"signature"`
		} `json:"signature"`
		Fee       int64  `json:"fee"`    //1000000
		Feefmt    string `json:"feefmt"` //"0.0100"
		Expire    int    `json:"expire"`
		Nonce     int    `json:"nonce"`
		From      string `json:"from"`
		To        string `json:"to"`
		Amount    int64  `json:"amount"`    //460510688
		Amountfmt string `json:"amountfmt"` //"4.6051"
		Hash      string `json:"hash"`
	} `json:"tx"`
	Receipt struct {
		Ty     int    `json:"ty"`
		Tyname string `json:"tyName"`
		Logs   []struct {
			Ty     int    `json:"ty"`
			Tyname string `json:"tyName"`
			Log    struct {
				Prev struct {
					Currency int    `json:"currency"`
					Balance  string `json:"balance"`
					Frozen   string `json:"frozen"`
					Addr     string `json:"addr"`
				} `json:"prev"`
				Current struct {
					Currency int    `json:"currency"`
					Balance  string `json:"balance"`
					Frozen   string `json:"frozen"`
					Addr     string `json:"addr"`
				} `json:"current"`
			} `json:"log"`
			Rawlog string `json:"rawLog"`
		} `json:"logs"`
	} `json:"receipt"`
	Proofs     []string `json:"proofs"`
	Height     int64    `json:"height"`
	Index      int      `json:"index"`
	Blocktime  int      `json:"blockTime"`
	Amount     int64    `json:"amount"`     //转移金额
	Fromaddr   string   `json:"fromAddr"`   //19ZVD1LPwCpceC4A79ZhnXjiJ6Ka1Qrj8Q
	Actionname string   `json:"actionName"` //"transfer"
	Assets     []struct {
		Exec   string `json:"exec"`   //"coins"
		Symbol string `json:"symbol"` //"DOM"
		Amount int    `json:"amount"` //460510688
	} `json:"assets"`
	Txproofs interface{} `json:"txProofs"`
	Fullhash string      `json:"fullHash"`
}

//// Transaction - transaction object
//type Transaction struct {
//	Id                int64             `json:"id"`
//	TransactionResult TransactionResult `json:"result"`
//}

type Log struct {
	Removed          bool     `json:"removed"`
	LogIndex         int      `json:"logIndex"`
	TransactionIndex int      `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
	BlockNumber      int      `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

type TransactionReceipt struct {
	Id                int64             `json:"id"`
	TransactionResult TransactionResult `json:"result"`
}
type TransactionResult struct {
	Tx         TX       `json:"tx"`
	Receipt    Receipt  `json:"receipt"`
	Proofs     []string `json:"proofs"`
	Height     int64    `json:"height"`
	BlockTime  int64    `json:"blockTime"`
	Amount     *big.Int `json:"amount"`
	FromAddr   string   `json:"fromAddr"`
	ActionName string   `json:"actionName"`
}

type TX struct {
	Execer  string `json:"execer"`
	Payload string `json:"payload"`
	Fee     int64  `json:"fee"`
	Expire  int32  `json:"expire"`
	Nonce   int    `json:"nonce"`
	To      string `json:"to"`
}

type Receipt struct {
	Ty int `json:"ty"`
}

type BaseBlock struct {
	Number           int64    `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Nonce            string   `json:"nonce"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	LogsBloom        string   `json:"logsBloom"`
	TransactionsRoot string   `json:"transactionsRoot"`
	StateRoot        string   `json:"stateRoot"`
	Miner            string   `json:"miner"`
	Difficulty       big.Int  `json:"difficulty"`
	TotalDifficulty  big.Int  `json:"totalDifficulty"`
	ExtraData        string   `json:"extraData"`
	Size             int      `json:"size"`
	GasLimit         int      `json:"gasLimit"`
	GasUsed          int      `json:"gasUsed"`
	Timestamp        int64    `json:"timestamp"`
	Uncles           []string `json:"uncles"`
}

type Txs struct {
	Execer  string `json:"execer"`
	Payload struct {
		Transfer struct {
			Cointoken string `json:"cointoken"`
			Amount    string `json:"amount"`
			Note      string `json:"note"`
			To        string `json:"to"`
		} `json:"transfer"`
		Tclose struct {
			Ticketid     []string `json:"ticketId"`
			Mineraddress string   `json:"minerAddress"`
		} `json:"tclose"`
		Miner struct {
			Bits     int    `json:"bits"`
			Reward   string `json:"reward"`
			Ticketid string `json:"ticketId"`
			Modify   string `json:"modify"`
			Privhash string `json:"privHash"`
			Vrfhash  string `json:"vrfHash"`
			Vrfproof string `json:"vrfProof"`
		} `json:"miner"`
		Ty int `json:"ty"`
	} `json:"payload,omitempty"`
	Rawpayload string `json:"rawPayload"`
	Signature  struct {
		Ty        int    `json:"ty"`
		Pubkey    string `json:"pubkey"`
		Signature string `json:"signature"`
	} `json:"signature"`
	Fee    int    `json:"fee"`
	Feefmt string `json:"feefmt"`
	Expire int    `json:"expire"`
	Nonce  int64  `json:"nonce"`
	From   string `json:"from"`
	To     string `json:"to"`
	Hash   string `json:"hash"`
}

type BlockResult struct {
	Items []struct {
		Block struct {
			Version    int    `json:"version"`
			Parenthash string `json:"parentHash"`
			Txhash     string `json:"txHash"`
			Statehash  string `json:"stateHash"`
			Height     int64  `json:"height"`
			Blocktime  int64  `json:"blockTime"`
			Txs        []Txs  `json:"txs"`
		} `json:"block"`
		Recipts []struct {
			Ty     int    `json:"ty"`
			Tyname string `json:"tyName"`
			//Logs   []struct {
			//	Ty     int    `json:"ty"`
			//	Tyname string `json:"tyName"`
			//	Rawlog string `json:"rawLog"`
			//	Log    struct {
			//		Ticketid   string `json:"ticketId"`
			//		Status     int    `json:"status"`
			//		Prevstatus int    `json:"prevStatus"`
			//		Addr       string `json:"addr"`
			//		Execaddr   string `json:"execAddr"`
			//		Prev       struct {
			//			Currency int    `json:"currency"`
			//			Balance  string `json:"balance"`
			//			Frozen   string `json:"frozen"`
			//			Addr     string `json:"addr"`
			//		} `json:"prev"`
			//		Current struct {
			//			Currency int    `json:"currency"`
			//			Balance  string `json:"balance"`
			//			Frozen   string `json:"frozen"`
			//			Addr     string `json:"addr"`
			//		} `json:"current"`
			//	} `json:"log,omitempty"`
			//} `json:"logs"`
		} `json:"recipts"`
	} `json:"items"`
}

type BlockHeader struct {
	Version    int    `json:"version"`
	Parenthash string `json:"parentHash"`
	Txhash     string `json:"txHash"`
	Statehash  string `json:"stateHash"`
	Height     int64  `json:"height"`
	Blocktime  int    `json:"blockTime"`
	Txcount    int    `json:"txCount"`
	Hash       string `json:"hash"`
	Difficulty int    `json:"difficulty"`
}

// FilterParams - Filter parameters object
type FilterParams struct {
	FromBlock string     `json:"fromBlock,omitempty"`
	ToBlock   string     `json:"toBlock,omitempty"`
	Address   []string   `json:"address,omitempty"`
	Topics    [][]string `json:"topics,omitempty"`
}

type ContractParams struct {
	To   string `json:"to,omitempty"`
	Data string `json:"data,omitempty"`
}

///////////////////////////////////////////////////////////////////////////private///////////////////////////////////////////////////////
// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Syncing) UnmarshalJSON(data []byte) error {
	proxy := new(proxySyncing)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	proxy.IsSyncing = true
	*s = *(*Syncing)(unsafe.Pointer(proxy))

	return nil
}

// MarshalJSON implements the json.Unmarshaler interface.
func (t T) MarshalJSON() ([]byte, error) {
	params := map[string]interface{}{
		"from": t.From,
	}
	if t.To != "" {
		params["to"] = t.To
	}
	if t.Fee != nil {
		params["fee"] = utils.BigToHex(*t.Fee)
	}
	if t.Amount != nil {
		params["amount"] = utils.BigToHex(*t.Amount)
	}
	if t.Data != "" {
		params["data"] = t.Data
	}
	if t.Nonce > 0 {
		params["nonce"] = utils.IntToHex(t.Nonce)
	}

	return json.Marshal(params)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//func (t *Transaction) UnmarshalJSON(data []byte) error {
//	proxy := new(Transaction)
//	if err := json.Unmarshal(data, proxy); err != nil {
//		return err
//	}
//
//	*t = *(*Transaction)(unsafe.Pointer(proxy))
//
//	return nil
//}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (log *Log) UnmarshalJSON(data []byte) error {
	proxy := new(proxyLog)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*log = *(*Log)(unsafe.Pointer(proxy))

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *TransactionReceipt) UnmarshalJSON(data []byte) error {
	proxy := new(proxyTransactionReceipt)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*t = *(*TransactionReceipt)(unsafe.Pointer(proxy))

	return nil
}

func (i *hexInt) UnmarshalJSON(data []byte) error {
	result, err := utils.ParseInt(string(bytes.Trim(data, `"`)))
	*i = hexInt(result)

	return err
}

func (i *hexBig) UnmarshalJSON(data []byte) error {
	result, err := utils.ParseBigInt(string(bytes.Trim(data, `"`)))
	i = (*hexBig)(result)

	return err
}

func (rpc *RpcClient) BlockNumber() (int64, error) {
	var response BlockHeader
	if err := rpc.CallNoAuth("DplatformOS.GetLastHeader", &response, nil); err != nil {
		return 0, err
	}
	return response.Height, nil
}

type BlockHashParam struct {
	Hashes []string `json:"hashes"`
}

func (rpc *RpcClient) GetBlockByHash(hash string, withTransactions bool) (*BlockResult, error) {
	p := BlockHashParam{
		Hashes: []string{hash},
	}
	return rpc.getBlock("DplatformOS.GetBlockByHashes", withTransactions, p)
}

func (rpc *RpcClient) GetBlockByHeight(height int64, withTransactions bool) (*BlockResult, error) {
	hash, err := rpc.GetBlockHashByHeight(height)
	if err != nil {
		return nil, err
	}
	p := BlockHashParam{
		Hashes: []string{hash},
	}
	return rpc.getBlock("DplatformOS.GetBlockByHashes", withTransactions, p)
}

type GetBlockHashByHeightRepo struct {
	Hash string `json:"hash"`
}

func (rpc *RpcClient) GetBlockHashByHeight(height int64) (string, error) {
	req := map[string]int64{"height": height}
	hash, err := rpc.RawCall("DplatformOS.GetBlockHash", "", req)
	repo := &GetBlockHashByHeightRepo{}
	err = json.Unmarshal(hash, repo)
	if err != nil {
		return "", err
	}
	return repo.Hash, err
}

type TransactionHashParam struct {
	Hash string `json:"hash"`
}

// GetTransactionByHash returns the information about a transaction requested by transaction hash.
func (rpc *RpcClient) GetTransactionByHash(hash string) (*Transaction, error) {
	p := TransactionHashParam{
		Hash: hash,
	}
	return rpc.getTransaction("DplatformOS.QueryTransaction", p)
}

type proxySyncing struct {
	IsSyncing     bool   `json:"-"`
	StartingBlock hexInt `json:"startingBlock"`
	CurrentBlock  hexInt `json:"currentBlock"`
	HighestBlock  hexInt `json:"highestBlock"`
}

type proxyLog struct {
	Removed          bool     `json:"removed"`
	LogIndex         hexInt   `json:"logIndex"`
	TransactionIndex hexInt   `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
	BlockNumber      hexInt   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

type proxyTransactionReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  hexInt `json:"transactionIndex"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       hexInt `json:"blockNumber"`
	CumulativeGasUsed hexInt `json:"cumulativeGasUsed"`
	GasUsed           hexInt `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Root              string `json:"root"`
	Status            string `json:"status,omitempty"`
}

func (rpc *RpcClient) getBlock(method string, withTransactions bool, params ...interface{}) (*BlockResult, error) {
	result, err := rpc.RawCall(method, "", params...)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(result, []byte("null")) {
		return nil, fmt.Errorf("get block null")
	}

	response := &BlockResult{}
	//if withTransactions {
	//	response = new(proxyBlockWithTransactions)
	//} else {
	//	response = new(proxyBlockWithoutTransactions)
	//}

	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}

	//block := response.toBlock()
	return response, nil
}

func (rpc *RpcClient) getTransaction(method string, params ...interface{}) (*Transaction, error) {
	result, err := rpc.RawCall(method, "", params...)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(result, []byte("null")) {
		return nil, fmt.Errorf("get block null")
	}

	response := &Transaction{}
	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}

	//tx := response.toTransaction()
	return response, nil

}
