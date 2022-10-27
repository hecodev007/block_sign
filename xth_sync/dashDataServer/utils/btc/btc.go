package btc

import (
	"fmt"
	"dashDataServer/common/log"
)

const (
	send_to_address     = "sendtoaddress"
	send_rawtransaction = "sendrawtransaction"
	get_transaction     = "gettransaction"
	get_raw_transaction = "getrawtransaction"
	get_balance         = "getbalance"
	list_transactions   = "listtransactions"
	list_unspend        = "listunspent"
	help                = "help"
	new_address         = "getnewaddress"
	get_block_count     = "getblockcount"
	validate_address    = "validateaddress"
	get_walletinfo      = "getwalletinfo"
	get_blockchaininfo  = "getblockchaininfo"
	get_block           = "getblock"
	get_block_hash      = "getblockhash"
	get_bestblock_hash  = "getbestblockhash"
)

//区块链详情
type BlockChainInfo struct {
	Chain         string  `json:"chain"`
	Blocks        uint    `json:"blocks"`
	Headers       uint    `json:"headers"`
	Bestblockhash string  `json:"bestblockhash"`
	Difficulty    float64 `json:"difficulty"`
	Mediantime    uint64  `json:"mediantime"`
	Chainwork     string  `json:"chainwork"`
}

//区块链详情
type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	Size              int64    `json:"size"`
	Height            int64    `json:"height"`
	Version           int      `json:"version"`
	Time              int64    `json:"time"`
	Chainwork         string   `json:"chainwork"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     string   `json:"nextblockhash"`
	Txs               []string `json:"tx"`
}

//区块链详情
type BlockWithTx struct {
	Hash              string         `json:"hash"`
	Confirmations     int64          `json:"confirmations"`
	Size              int64          `json:"size"`
	Height            int64          `json:"height"`
	Version           int            `json:"version"`
	Time              int64          `json:"time"`
	Chainwork         string         `json:"chainwork"`
	PreviousBlockHash string         `json:"previousblockhash"`
	NextBlockHash     string         `json:"nextblockhash"`
	Txs               []*Transaction `json:"tx"`
}

type Transaction struct {
	Txid          string       `json:"txid"`
	Hash          string       `json:"hash"`
	Version       int          `json:"version"`
	Size          int          `json:"size"`
	LockTime      int64        `json:"locktime"`
	Vin           []proxyTxIn  `json:"vin"`
	Vout          []proxyTxOut `json:"vout"`
	BlockHash     string       `json:"blockhash"`
	Confirmations int64        `json:"confirmations"`
	Time          int64        `json:"time"`
}

type proxyTxIn struct {
	Txid     string `json:"txid,omitempty"`
	Vout     int    `json:"vout,omitempty"`
	Coinbase string `json:"coinbase,omitempty"`
	Sequence int64  `json:"sequence"`
}

type proxyTxOut struct {
	Value        float64      `json:"value"`
	Index        int          `json:"n"`
	ScriptPubkey scriptPubkey `json:"scriptPubkey"`
}

func (rpc *RpcClient) GetBalance(address string) (bestBlockHash string, err error) {

	err = rpc.CallWithAuth(get_bestblock_hash, rpc.Credentials, &bestBlockHash, address)
	return
}

func (rpc *RpcClient) GetTransaction() {

}

func (rpc *RpcClient) GetBlockChainInfo() (blockChainInfo BlockChainInfo, err error) {

	err = rpc.CallWithAuth(get_blockchaininfo, rpc.Credentials, &blockChainInfo)
	fmt.Printf("respone1 : %+v\n", blockChainInfo)
	return
}

// GetBestBlockHash returns the bestblockhash.
func (rpc *RpcClient) GetBlockCount() (bestBlockCount int64, err error) {
	err = rpc.CallWithAuth(get_block_count, rpc.Credentials, &bestBlockCount)
	return
}

// GetBestBlockHash returns the bestblockhash.
func (rpc *RpcClient) GetBestBlockHash() (bestBlockHash string, err error) {

	err = rpc.CallWithAuth(get_bestblock_hash, rpc.Credentials, &bestBlockHash)
	return
}

// GetBlockByHash returns block infomations by hash.
func (rpc *RpcClient) GetBlockByHash(h string) (block Block, err error) {

	err = rpc.CallWithAuth(get_block, rpc.Credentials, &block, h)
	return
}

// GetBlockByHash returns block infomations by hash.
func (rpc *RpcClient) GetBlockWithTxByHash(h string) (block BlockWithTx, err error) {

	err = rpc.CallWithAuth(get_block, rpc.Credentials, &block, h, 2)
	if err != nil {
		return
	}
	for _, tx := range block.Txs {
		tx.Confirmations = block.Confirmations
		tx.Time = block.Time
	}
	return
}

// GetBlockByHeight returns block infomations by height.
func (rpc *RpcClient) GetBlockByHeight(h int64) (block Block, err error) {
	var blockHash string

	blockHash, err = rpc.GetBlockHash(h)
	if err != nil {
		return
	}
	block, err = rpc.GetBlockByHash(blockHash)
	return
}

// GetBlockByHeight returns block infomations by height.
func (rpc *RpcClient) GetBlockByHeight2(h int64) (block BlockWithTx, err error) {
	var blockHash string

	blockHash, err = rpc.GetBlockHash(h)
	if err != nil {
		log.Info(err.Error())
		return
	}
	block, err = rpc.GetBlockWithTxByHash(blockHash)
	if err != nil {
		log.Info(err.Error())
		return
	}
	return
}

// GetBlockHash returns block hash with block height.
func (rpc *RpcClient) GetBlockHash(height int64) (blockHash string, err error) {

	err = rpc.CallWithAuth(get_block_hash, rpc.Credentials, &blockHash, height)
	return
}

// GetRawTransaction returns raw transaction by transaction hash.
func (rpc *RpcClient) GetRawTransaction(h string) (tx Transaction, err error) {

	err = rpc.CallWithAuth(get_raw_transaction, rpc.Credentials, &tx, h, 1)
	return
}

// SendToAddress sends coin to dest address.
func (rpc *RpcClient) SendToAddress(addr, amount string) (txid interface{}, err error) {

	err = rpc.CallWithAuth(send_to_address, rpc.Credentials, &txid, []interface{}{addr, amount})
	return
}

//UTXO 数据结构
type UnSpend struct {
	Txid          string
	Vout          uint
	Address       string
	RedeemScript  string
	ScriptPubKey  string
	Amount        float64
	Confirmations uint64
	Spendable     bool
	Solvable      bool
	Safe          bool
}

func (rpc *RpcClient) GetUnSpends(addr string) (utxo []UnSpend, err error) {

	err = rpc.CallWithAuth(list_unspend, rpc.Credentials, &utxo, addr)
	return
}

func (rpc *RpcClient) SendRawTransaction(txhash string) (txid string, err error) {

	err = rpc.CallWithAuth(send_rawtransaction, rpc.Credentials, &txid, txhash)
	return
}

// Close closes rpc connection.
/*func (rpc *RpcClient) Close() {
	rpc.client.Close()
}*/
