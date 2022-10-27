package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"rsksync/utils"
	"unsafe"
)

const BlockLatest = "latest"
const WEI = 18

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
	From     string
	To       string
	Gas      *big.Int
	GasPrice *big.Int
	Value    *big.Int
	Data     string
	Nonce    int
}

// Transaction - transaction object
type Transaction struct {
	Hash        string   `json:"hash"`
	Nonce       int      `json:"nonce"`
	BlockHash   string   `json:"blockHash"`
	BlockNumber int64    `json:"blockNumber"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Value       *big.Int `json:"value"`
	Gas         int64    `json:"gas"`
	GasPrice    *big.Int `json:"gasPrice"`
	Input       string   `json:"input"`
}

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
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  int    `json:"transactionIndex"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       int    `json:"blockNumber"`
	CumulativeGasUsed int    `json:"cumulativeGasUsed"`
	GasUsed           int64  `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Root              string `json:"root"`
	Status            string `json:"status,omitempty"`
	LogIndex          int    `json:"logIndex"`
	Removed           bool   `json:"removed"`
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

// Block - block object
type Block struct {
	Number           int64         `json:"number"`
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parentHash"`
	Nonce            string        `json:"nonce"`
	Sha3Uncles       string        `json:"sha3Uncles"`
	LogsBloom        string        `json:"logsBloom"`
	TransactionsRoot string        `json:"transactionsRoot"`
	StateRoot        string        `json:"stateRoot"`
	Miner            string        `json:"miner"`
	Difficulty       *big.Int      `json:"difficulty"`
	TotalDifficulty  *big.Int      `json:"totalDifficulty"`
	ExtraData        string        `json:"extraData"`
	Size             int           `json:"size"`
	GasLimit         int           `json:"gasLimit"`
	GasUsed          int           `json:"gasUsed"`
	Timestamp        int64         `json:"timestamp"`
	Uncles           []string      `json:"uncles"`
	Transactions     []Transaction `json:"transactions"`
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
	if t.Gas != nil {
		params["gas"] = utils.BigToHex(*t.Gas)
	}
	if t.GasPrice != nil {
		params["gasPrice"] = utils.BigToHex(*t.GasPrice)
	}
	if t.Value != nil {
		params["value"] = utils.BigToHex(*t.Value)
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
func (t *Transaction) UnmarshalJSON(data []byte) error {
	proxy := new(proxyTransaction)
	if err := json.Unmarshal(data, proxy); err != nil {
		return err
	}

	*t = *(*Transaction)(unsafe.Pointer(proxy))

	return nil
}

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

func (rpc *RpcClient) IsContractTx(t *Transaction) bool {
	if t.Input == "0x" || t.Input == "0x00" {
		return false
	}

	if t.Value.Sign() > 0 {
		got, _ := rpc.GetCode(t.To, BlockLatest)
		if got == "0x" {
			return false
		}
	}
	return true
}

// EthGetBalance returns the balance of the account of given address in wei.
func (rpc *RpcClient) GetBalance(address, block string) (*big.Int, error) {
	var response string
	if err := rpc.CallNoAuth("eth_getBalance", &response, address, block); err != nil {
		return nil, err
	}

	return utils.ParseBigInt(response)
}

func (rpc *RpcClient) GetBalanceToken(address, data string) (*big.Int, error) {
	var response string
	cp := ContractParams{
		address, data,
	}
	if err := rpc.CallNoAuth("eth_call", &response, cp, BlockLatest); err != nil {
		return nil, err
	}

	return utils.ParseBigInt(response)
}

// GetTransactionCount returns the number of transactions sent from an address.
func (rpc *RpcClient) GetNonce(address, block string) (int, error) {
	var response string

	if err := rpc.CallNoAuth("eth_getTransactionCount", &response, address, block); err != nil {
		return 0, err
	}

	return utils.ParseInt(response)
}

// GasPrice returns the current price per gas in wei.
func (rpc *RpcClient) GasPrice() (*big.Int, error) {
	var response string
	if err := rpc.CallNoAuth("eth_gasPrice", &response); err != nil {
		return nil, err
	}

	return utils.ParseBigInt(response)
}

// GetCode returns code at a given address.
func (rpc *RpcClient) GetCode(address, block string) (string, error) {
	var code string

	err := rpc.CallNoAuth("eth_getCode", &code, address, block)
	return code, err
}

// BlockNumber returns the number of most recent block.
func (rpc *RpcClient) BlockNumber() (int64, error) {
	var response string
	if err := rpc.CallNoAuth("eth_blockNumber", &response); err != nil {
		return 0, err
	}

	return utils.ParseInt64(response)
}

// SendTransaction creates new message CallNoAuth transaction or a contract creation, if the data field contains code.
func (rpc *RpcClient) SendTransaction(transaction T) (string, error) {
	var hash string

	err := rpc.CallNoAuth("eth_sendTransaction", &hash, transaction)
	return hash, err
}

// SendRawTransaction creates new message CallNoAuth transaction or a contract creation for signed transactions.
func (rpc *RpcClient) SendRawTransaction(data string) (string, error) {

	rpc.mutex.Lock()
	defer rpc.mutex.Unlock()

	var hash string
	err := rpc.CallNoAuth("eth_sendRawTransaction", &hash, data)
	return hash, err
}

// Call executes a new message CallNoAuth immediately without creating a transaction on the block chain.
func (rpc *RpcClient) Call(transaction T, tag string) (string, error) {
	var data string

	err := rpc.CallNoAuth("eth_call", &data, transaction, tag)
	return data, err
}

// EstimateGas makes a call or transaction, which won't be added to the blockchain and returns the used gas, which can be used for estimating the used gas.
func (rpc *RpcClient) EstimateGas(transaction T) (int64, error) {
	var response string

	err := rpc.CallNoAuth("eth_estimateGas", &response, transaction)
	if err != nil {
		return 0, err
	}

	return utils.ParseInt64(response)
}

// Web3ClientVersion returns the current client version.
func (rpc *RpcClient) Web3ClientVersion() (string, error) {
	var clientVersion string

	err := rpc.CallNoAuth("web3_clientVersion", &clientVersion)
	return clientVersion, err
}

// Web3Sha3 returns Keccak-256 (not the standardized SHA3-256) of the given data.
func (rpc *RpcClient) Web3Sha3(data []byte) (string, error) {
	var hash string

	err := rpc.CallNoAuth("web3_sha3", &hash, fmt.Sprintf("0x%x", data))
	return hash, err
}

// NetVersion returns the current network protocol version.
func (rpc *RpcClient) NetVersion() (string, error) {
	var version string

	err := rpc.CallNoAuth("net_version", &version)
	return version, err
}

// NetListening returns true if client is actively listening for network connections.
func (rpc *RpcClient) NetListening() (bool, error) {
	var listening bool

	err := rpc.CallNoAuth("net_listening", &listening)
	return listening, err
}

// NetPeerCount returns number of peers currently connected to the client.
func (rpc *RpcClient) NetPeerCount() (int, error) {
	var response string
	if err := rpc.CallNoAuth("net_peerCount", &response); err != nil {
		return 0, err
	}

	return utils.ParseInt(response)
}

// ProtocolVersion returns the current ethereum protocol version.
func (rpc *RpcClient) ProtocolVersion() (string, error) {
	var protocolVersion string

	err := rpc.CallNoAuth("eth_protocolVersion", &protocolVersion)
	return protocolVersion, err
}

// Syncing returns an object with data about the sync status or false.
func (rpc *RpcClient) Syncing() (*Syncing, error) {
	result, err := rpc.RawCall("eth_syncing", "")
	if err != nil {
		return nil, err
	}
	syncing := new(Syncing)
	if bytes.Equal(result, []byte("false")) {
		return syncing, nil
	}
	err = json.Unmarshal(result, syncing)
	return syncing, err
}

// Coinbase returns the client coinbase address
func (rpc *RpcClient) Coinbase() (string, error) {
	var address string

	err := rpc.CallNoAuth("eth_coinbase", &address)
	return address, err
}

// Mining returns true if client is actively mining new blocks.
func (rpc *RpcClient) Mining() (bool, error) {
	var mining bool

	err := rpc.CallNoAuth("eth_mining", &mining)
	return mining, err
}

// Hashrate returns the number of hashes per second that the node is mining with.
func (rpc *RpcClient) Hashrate() (int, error) {
	var response string

	if err := rpc.CallNoAuth("eth_hashrate", &response); err != nil {
		return 0, err
	}

	return utils.ParseInt(response)
}

// Accounts returns a list of addresses owned by client.
func (rpc *RpcClient) Accounts() ([]string, error) {
	accounts := []string{}

	err := rpc.CallNoAuth("eth_accounts", &accounts)
	return accounts, err
}

// GetStorageAt returns the value from a storage position at a given address.
func (rpc *RpcClient) GetStorageAt(data string, position int, tag string) (string, error) {
	var result string

	err := rpc.CallNoAuth("eth_getStorageAt", &result, data, utils.IntToHex(position), tag)
	return result, err
}

// GetBlockTransactionCountByHash returns the number of transactions in a block from a block matching the given block hash.
func (rpc *RpcClient) GetBlockTransactionCountByHash(hash string) (int, error) {
	var response string

	if err := rpc.CallNoAuth("eth_getBlockTransactionCountByHash", &response, hash); err != nil {
		return 0, err
	}

	return utils.ParseInt(response)
}

// GetBlockTransactionCountByNumber returns the number of transactions in a block from a block matching the given block
func (rpc *RpcClient) GetBlockTransactionCountByNumber(number int) (int, error) {
	var response string

	if err := rpc.CallNoAuth("eth_getBlockTransactionCountByNumber", &response, utils.IntToHex(number)); err != nil {
		return 0, err
	}

	return utils.ParseInt(response)
}

// Sign signs data with a given address.
// Calculates an Ethereum specific signature with: sign(keccak256("\x19Ethereum Signed Message:\n" + len(message) + message)))
func (rpc *RpcClient) Sign(address, data string) (string, error) {
	var signature string

	err := rpc.CallNoAuth("eth_sign", &signature, address, data)
	return signature, err
}

// GetBlockByHash returns information about a block by hash.
func (rpc *RpcClient) GetBlockByHash(hash string, withTransactions bool) (*Block, error) {
	return rpc.getBlock("eth_getBlockByHash", withTransactions, hash, withTransactions)
}

// GetBlockByNumber returns information about a block by block number.
func (rpc *RpcClient) GetBlockByNumber(number int64, withTransactions bool) (*Block, error) {
	return rpc.getBlock("eth_getBlockByNumber", withTransactions, utils.Int64ToHex(number), withTransactions)
}

// GetTransactionByHash returns the information about a transaction requested by transaction hash.
func (rpc *RpcClient) GetTransactionByHash(hash string) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByHash", hash)
}

// GetTransactionByBlockHashAndIndex returns information about a transaction by block hash and transaction index position.
func (rpc *RpcClient) GetTransactionByBlockHashAndIndex(blockHash string, transactionIndex int64) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByBlockHashAndIndex", blockHash, utils.Int64ToHex(transactionIndex))
}

// GetTransactionByBlockNumberAndIndex returns information about a transaction by block number and transaction index position.
func (rpc *RpcClient) GetTransactionByBlockNumberAndIndex(blockNumber, transactionIndex int) (*Transaction, error) {
	return rpc.getTransaction("eth_getTransactionByBlockNumberAndIndex", utils.IntToHex(blockNumber), utils.IntToHex(transactionIndex))
}

// GetTransactionReceipt returns the receipt of a transaction by transaction hash.
// Note That the receipt is not available for pending transactions.
func (rpc *RpcClient) GetTransactionReceipt(hash string) (*TransactionReceipt, error) {
	transactionReceipt := new(TransactionReceipt)

	err := rpc.CallNoAuth("eth_getTransactionReceipt", transactionReceipt, hash)
	if err != nil {
		return nil, err
	}

	return transactionReceipt, nil
}

// GetCompilers returns a list of available compilers in the client.
func (rpc *RpcClient) GetCompilers() ([]string, error) {
	compilers := []string{}

	err := rpc.CallNoAuth("eth_getCompilers", &compilers)
	return compilers, err
}

// NewFilter creates a new filter object.
func (rpc *RpcClient) NewFilter(params FilterParams) (string, error) {
	var filterID string
	err := rpc.CallNoAuth("eth_newFilter", &filterID, params)
	return filterID, err
}

// NewBlockFilter creates a filter in the node, to notify when a new block arrives.
// To check if the state has changed, call GetFilterChanges.
func (rpc *RpcClient) NewBlockFilter() (string, error) {
	var filterID string
	err := rpc.CallNoAuth("eth_newBlockFilter", &filterID)
	return filterID, err
}

// NewPendingTransactionFilter creates a filter in the node, to notify when new pending transactions arrive.
// To check if the state has changed, call GetFilterChanges.
func (rpc *RpcClient) NewPendingTransactionFilter() (string, error) {
	var filterID string
	err := rpc.CallNoAuth("eth_newPendingTransactionFilter", &filterID)
	return filterID, err
}

// UninstallFilter uninstalls a filter with given id.
func (rpc *RpcClient) UninstallFilter(filterID string) (bool, error) {
	var res bool
	err := rpc.CallNoAuth("eth_uninstallFilter", &res, filterID)
	return res, err
}

// GetFilterChanges polling method for a filter, which returns an array of logs which occurred since last poll.
func (rpc *RpcClient) GetFilterChanges(filterID string) ([]Log, error) {
	var logs = []Log{}
	err := rpc.CallNoAuth("eth_getFilterChanges", &logs, filterID)
	return logs, err
}

// GetFilterLogs returns an array of all logs matching filter with given id.
func (rpc *RpcClient) GetFilterLogs(filterID string) ([]Log, error) {
	var logs = []Log{}
	err := rpc.CallNoAuth("eth_getFilterLogs", &logs, filterID)
	return logs, err
}

// GetLogs returns an array of all logs matching a given filter object.
func (rpc *RpcClient) GetLogs(params FilterParams) ([]Log, error) {
	var logs = []Log{}
	err := rpc.CallNoAuth("eth_getLogs", &logs, params)
	return logs, err
}

// Eth1 returns 1 ethereum value (10^18 wei)
func (rpc *RpcClient) Eth1() *big.Int {
	return eth1()
}

type proxyBlock interface {
	toBlock() Block
}

type proxySyncing struct {
	IsSyncing     bool   `json:"-"`
	StartingBlock hexInt `json:"startingBlock"`
	CurrentBlock  hexInt `json:"currentBlock"`
	HighestBlock  hexInt `json:"highestBlock"`
}

type proxyTransaction struct {
	Hash             string `json:"hash"`
	Nonce            string `json:"nonce"`
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Input            string `json:"input"`
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

type proxyBlockWithTransactions struct {
	Number           string             `json:"number"`
	Hash             string             `json:"hash"`
	ParentHash       string             `json:"parentHash"`
	Nonce            string             `json:"nonce"`
	Sha3Uncles       string             `json:"sha3Uncles"`
	LogsBloom        string             `json:"logsBloom"`
	TransactionsRoot string             `json:"transactionsRoot"`
	StateRoot        string             `json:"stateRoot"`
	Miner            string             `json:"miner"`
	Difficulty       string             `json:"difficulty"`
	TotalDifficulty  string             `json:"totalDifficulty"`
	ExtraData        string             `json:"extraData"`
	Size             string             `json:"size"`
	GasLimit         string             `json:"gasLimit"`
	GasUsed          string             `json:"gasUsed"`
	Timestamp        string             `json:"timestamp"`
	Uncles           []string           `json:"uncles"`
	Transactions     []proxyTransaction `json:"transactions"`
}

func (proxy *proxyTransaction) toTransaction() Transaction {
	tx := Transaction{
		Hash:      proxy.Hash,
		BlockHash: proxy.BlockHash,
		From:      proxy.From,
		To:        proxy.To,
		Input:     proxy.Input,
	}
	tx.Value, _ = utils.ParseBigInt(proxy.Value)
	tx.BlockNumber, _ = utils.ParseInt64(proxy.BlockNumber)
	tx.Gas, _ = utils.ParseInt64(proxy.Gas)
	tx.GasPrice, _ = utils.ParseBigInt(proxy.GasPrice)
	tx.Nonce, _ = utils.ParseInt(proxy.Nonce)

	return tx
}

func (proxy *proxyBlockWithTransactions) toBlock() Block {

	block := Block{
		Hash:             proxy.Hash,
		ParentHash:       proxy.ParentHash,
		Nonce:            proxy.Nonce,
		Sha3Uncles:       proxy.Sha3Uncles,
		LogsBloom:        proxy.LogsBloom,
		TransactionsRoot: proxy.TransactionsRoot,
		StateRoot:        proxy.StateRoot,
		Miner:            proxy.Miner,
		ExtraData:        proxy.ExtraData,
		Uncles:           proxy.Uncles,
	}
	block.Number, _ = utils.ParseInt64(proxy.Number)
	block.Difficulty, _ = utils.ParseBigInt(proxy.Difficulty)
	block.TotalDifficulty, _ = utils.ParseBigInt(proxy.TotalDifficulty)
	block.Size, _ = utils.ParseInt(proxy.Size)
	block.GasLimit, _ = utils.ParseInt(proxy.GasLimit)
	block.GasUsed, _ = utils.ParseInt(proxy.GasUsed)
	block.Timestamp, _ = utils.ParseInt64(proxy.Timestamp)

	block.Transactions = make([]Transaction, len(proxy.Transactions))
	for i := range proxy.Transactions {
		block.Transactions[i] = proxy.Transactions[i].toTransaction()

	}

	return block
}

type proxyBlockWithoutTransactions struct {
	Number           string   `json:"number"`
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Nonce            string   `json:"nonce"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	LogsBloom        string   `json:"logsBloom"`
	TransactionsRoot string   `json:"transactionsRoot"`
	StateRoot        string   `json:"stateRoot"`
	Miner            string   `json:"miner"`
	Difficulty       string   `json:"difficulty"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	ExtraData        string   `json:"extraData"`
	Size             string   `json:"size"`
	GasLimit         string   `json:"gasLimit"`
	GasUsed          string   `json:"gasUsed"`
	Timestamp        string   `json:"timestamp"`
	Uncles           []string `json:"uncles"`
	Transactions     []string `json:"transactions"`
}

func (proxy *proxyBlockWithoutTransactions) toBlock() Block {
	block := Block{
		Hash:             proxy.Hash,
		ParentHash:       proxy.ParentHash,
		Nonce:            proxy.Nonce,
		Sha3Uncles:       proxy.Sha3Uncles,
		LogsBloom:        proxy.LogsBloom,
		TransactionsRoot: proxy.TransactionsRoot,
		StateRoot:        proxy.StateRoot,
		Miner:            proxy.Miner,
		ExtraData:        proxy.ExtraData,
		Uncles:           proxy.Uncles,
	}

	block.Number, _ = utils.ParseInt64(proxy.Number)
	block.Difficulty, _ = utils.ParseBigInt(proxy.Difficulty)
	block.TotalDifficulty, _ = utils.ParseBigInt(proxy.TotalDifficulty)
	block.Size, _ = utils.ParseInt(proxy.Size)
	block.GasLimit, _ = utils.ParseInt(proxy.GasLimit)
	block.GasUsed, _ = utils.ParseInt(proxy.GasUsed)
	block.Timestamp, _ = utils.ParseInt64(proxy.Timestamp)

	block.Transactions = make([]Transaction, len(proxy.Transactions))
	for i := range proxy.Transactions {
		block.Transactions[i] = Transaction{
			Hash: proxy.Transactions[i],
		}
	}

	return block
}

// eth1 returns 1 ethereum value (10^18 wei)
func eth1() *big.Int {
	return big.NewInt(1000000000000000000)
}

func (rpc *RpcClient) getBlock(method string, withTransactions bool, params ...interface{}) (*Block, error) {
	result, err := rpc.RawCall(method, "", params...)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(result, []byte("null")) {
		return nil, fmt.Errorf("get block null")
	}

	var response proxyBlock
	if withTransactions {
		response = new(proxyBlockWithTransactions)
	} else {
		response = new(proxyBlockWithoutTransactions)
	}

	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}

	block := response.toBlock()
	return &block, nil
}

func (rpc *RpcClient) getTransaction(method string, params ...interface{}) (*Transaction, error) {
	//transaction := new(Transaction)
	//
	//err := rpc.CallNoAuth(method, transaction, params...)
	//return transaction, err
	result, err := rpc.RawCall(method, "", params...)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(result, []byte("null")) {
		return nil, fmt.Errorf("get block null")
	}

	response := &proxyTransaction{}
	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}

	tx := response.toTransaction()
	return &tx, nil

}
