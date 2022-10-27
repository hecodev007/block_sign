package test

import (
	"fmt"
	"github.com/rjman-ljm/go-substrate-crypto/crypto"
	"github.com/rjman-ljm/substrate-go/client"
	"github.com/rjman-ljm/substrate-go/expand"
	"github.com/rjman-ljm/substrate-go/tx"
	"testing"
)

func Test_Tx2(t *testing.T) {
	// 1. 初始化rpc客户端
	c, err := client.New("ws://127.0.0.1:8087")
	if err != nil {
		t.Fatal(err)
	}
	//2. 如果某些链（例如：chainX)的地址的字节前面需要0xff,则下面这个值设置为false
	expand.SetSerDeOptions(false)
	from := "5FNTYUQwxjrVE5zRRH1hKh6fZ72AosHB7ThVnNnq9Bv9BFjm"
	to := "5FnCTkAtgLinh6apZJwTX7n72H1A37MHE6xAXChZDbtUWMSg"
	amount := uint64(10000000000)
	//3. 获取from地址的nonce
	acc, err := c.GetAccountInfo(from)
	if err != nil {
		t.Fatal(err)
	}
	nonce := uint64(acc.Nonce)
	//4. 创建一个substrate交易，这个方法满足所有遵循substrate 的交易结构的链
	transaction := tx.NewSubstrateTransaction(from, nonce)
	//5. 初始化metadata的扩张结构
	ed, err := expand.NewMetadataExpand(c.Meta)
	if err != nil {
		t.Fatal(err)
	}
	//6. 初始化Balances.transfer的call方法
	call, err := ed.BalanceTransferCall(to, amount)
	if err != nil {
		t.Fatal(err)
	}
	/*
		//Balances.transfer_keep_alive  call方法
		btkac,err:=ed.BalanceTransferKeepAliveCall(to,amount)
	*/

	/*
		toAmount:=make(map[string]uint64)
		toAmount[to] = amount
		//...
		//true: user Balances.transfer_keep_alive  false: Balances.transfer
		ubtc,err:=ed.UtilityBatchTxCall(toAmount,false)
	*/

	//7. 设置交易的必要参数
	transaction.SetGenesisHashAndBlockHash(c.GetGenesisHash(), c.GetGenesisHash()).
		SetSpecAndTxVersion(uint32(c.SpecVersion), uint32(c.TransactionVersion)).
		SetCall(call) //设置call
	//8. 签名交易
	sig, err := transaction.SignTransaction("0x3c0c4fc26010d0512cd36a0f467375b3dbe2f207bbfda0c551b5e41ee495e909", crypto.Sr25519Type)
	if err != nil {
		t.Fatal(err)
	}
	//9. 提交交易
	var result interface{}
	err = c.Api.Client.Call(&result, "author_submitExtrinsic", sig)
	if err != nil {
		t.Fatal(err)
	}
	//10. txid
	txid := result.(string)
	fmt.Println(txid)
}
