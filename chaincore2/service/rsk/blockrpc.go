package rsk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	"math/big"
	"strings"
)

type RskClient struct {
	Nodeurl string
}

func NewRskBlock(node string) *RskClient {
	client := new(RskClient)
	client.Nodeurl = node
	return client
}

type BlockNumber struct {
	ID      interface{}  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// 获取节点区块高度
func (b *RskClient) GetBlockCount() (int64, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return 0, err
	}
	var block BlockNumber
	err = json.Unmarshal(bytes, &block)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	currentheight := common.StrBaseToInt(block.Result, 16)
	return int64(currentheight), nil
}

//////////////////////////////////////////////////////
type Transaction struct {
	BlockHash        string      `json:"blockHash"`
	BlockNumber      string      `json:"blockNumber"`
	From             string      `json:"from"`
	Gas              string      `json:"gas"`
	GasPrice         string      `json:"gasPrice"`
	Hash             string      `json:"hash"`
	Input            string      `json:"input"`
	Nonce            string      `json:"nonce"`
	R                interface{} `json:"r"`
	S                interface{} `json:"s"`
	To               string      `json:"to"`
	TransactionIndex string      `json:"transactionIndex"`
	V                interface{} `json:"v"`
	Value            string      `json:"value"`
}

type BlockDataResult  struct {
	BitcoinMergedMiningCoinbaseTransaction string `json:"bitcoinMergedMiningCoinbaseTransaction"`
	BitcoinMergedMiningHeader              string `json:"bitcoinMergedMiningHeader"`
	BitcoinMergedMiningMerkleProof         string `json:"bitcoinMergedMiningMerkleProof"`
	CumulativeDifficulty                   string `json:"cumulativeDifficulty"`
	Difficulty                             string `json:"difficulty"`
	ExtraData                              string `json:"extraData"`
	GasLimit                               string `json:"gasLimit"`
	GasUsed                                string `json:"gasUsed"`
	Hash                                   string `json:"hash"`
	HashForMergedMining                    string `json:"hashForMergedMining"`
	LogsBloom                              string `json:"logsBloom"`
	Miner                                  string `json:"miner"`
	MinimumGasPrice                        string `json:"minimumGasPrice"`
	Number                                 string `json:"number"`
	PaidFees                               string `json:"paidFees"`
	ParentHash                             string `json:"parentHash"`
	ReceiptsRoot                           string `json:"receiptsRoot"`
	Sha3Uncles                             string `json:"sha3Uncles"`
	Size                                   string `json:"size"`
	StateRoot                              string `json:"stateRoot"`
	Timestamp                              string `json:"timestamp"`
	TotalDifficulty                        string `json:"totalDifficulty"`
	Transactions                           []Transaction `json:"transactions"`

	TransactionsRoot string        `json:"transactionsRoot"`
	Uncles           []interface{} `json:"uncles"`
}

type BlockDataStruct struct {
	ID      interface{}  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result BlockDataResult `json:"result"`
}

// 获取区块数据
func (b *RskClient) GetBlockDataByNumber(number int64) (*BlockDataResult, error) {
	var blockData BlockDataStruct

	numberStr := common.Int64ToString(number, 16) //先将number转为16进制数

	numberOf16 := fmt.Sprintf("0x%s",numberStr)

	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{numberOf16, true},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return &blockData.Result, err
	}
	err = json.Unmarshal(bytes, &blockData)
	if err != nil {
		fmt.Println(err.Error())
		return &blockData.Result, err
	}

	return &blockData.Result, nil
}

func (b *RskClient) ParseTransferData(input string) (to string, amount *big.Int, err error) {
	if strings.Index(input, "0xa9059cbb") != 0 {
		return to, amount, errors.New("input is not transfer data")
	}
	if len(input) < 138 {
		return to, amount, fmt.Errorf("input data isn't 138 , size %d ", 138)
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:138], 16)
	if amount.Sign() < 0 {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}

type TransactionReceiptResult  struct {
	BlockHash         string      `json:"blockHash"`
	BlockNumber       string      `json:"blockNumber"`
	ContractAddress   interface{} `json:"contractAddress"`
	CumulativeGasUsed string      `json:"cumulativeGasUsed"`
	From              string      `json:"from"`
	GasUsed           string      `json:"gasUsed"`
	Logs              []struct {
		Address          string   `json:"address"`
		BlockHash        string   `json:"blockHash"`
		BlockNumber      string   `json:"blockNumber"`
		Data             string   `json:"data"`
		LogIndex         string   `json:"logIndex"`
		Topics           []string `json:"topics"`
		TransactionHash  string   `json:"transactionHash"`
		TransactionIndex string   `json:"transactionIndex"`
	} `json:"logs"`
	LogsBloom        string `json:"logsBloom"`
	Root             string `json:"root"`
	Status           string `json:"status"`
	To               string `json:"to"`
	TransactionHash  string `json:"transactionHash"`
	TransactionIndex string `json:"transactionIndex"`
}

func (b *RskClient) GetTransactionReceipt(txId string) (*TransactionReceiptResult, error) {
	var blockData struct {
		ID      interface{}  `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Result TransactionReceiptResult `json:"result"`
	}

	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionReceipt",
		"params":  []interface{}{txId},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return &blockData.Result, err
	}
	err = json.Unmarshal(bytes, &blockData)
	if err != nil {
		fmt.Println(err.Error())
		return &blockData.Result, err
	}

	return &blockData.Result, nil
}

//根据输入的addr判断是否是合约地址，如果合约地址返回true，否则返回false
func (b *RskClient) IsContact(addr string) (bool, error) {
	var blockData struct {
		ID      interface{}  `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Result string `json:"result"`
	}

	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "eth_getCode",
		"params":  []interface{}{addr, "latest"},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(bytes, &blockData)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	if blockData.Result == "0x00" {
		return false, nil
	} else {
		return true, nil
	}
}


///////////////
// 获取区块数据
func getblock_data(val interface{}) (string, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "getblock",
		"params":  []interface{}{val, 1},
	}
	req.JSONBody(reqbody)
	respdata, err := req.String()
	return respdata, err
}
