package wtc

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func Test_acli(t *testing.T) {
	rpc := NewRpcClient("https://node.waltonchain.pro", "", "")
	t.Log(rpc.GetTransactionCount("0x5a37e535a430a9a5b3da17c7e68c3647035bd7bd", "pending"))
	t.Log(rpc.SendRawTransaction(""))
	return
	t.Log(rpc.GasPrice())
	return
	balanceof, err := rpc.BalanceOf("0x668df218d073f413ed2fcea0d48cfbfd59c030ae", "0x5a37e535a430a9a5b3da17c7e68c3647035bd7bd")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(balanceof.String())
	return
	balance, err := rpc.GetBalance("0x76bb7a894587c67a49165bbac9ef5ff202e6f2b8")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(balance.String())
	return
	h, err := rpc.BlockNumber()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(h)
	block, err := rpc.BlockByNumber(2031462)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(block))
	receipt, err := rpc.TransactionReceipt("0xdade9d7f316fa709ade853f465d320fa58f81fbefdaf94f1da87e6265e2c715c")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(receipt))
	d, err := hex.DecodeString(strings.TrimPrefix(receipt.Logs[0].Data, "0x"))
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(receipt.Logs[0].Data, big.NewInt(0).SetBytes(d).String())
}
func Test_json(t *testing.T) {
	d := "{\"uid\":5,\"txid\":\"2e9ca4ab0738a471160b06420fd1229792b3a8d33ae15f4320b709388a6d3857\",\"coin\":\"trx\"}"

	var jsonObj map[string]interface{}
	json.Unmarshal([]byte(d), &jsonObj)
	t.Log(jsonObj["uid"], jsonObj["txid"], jsonObj["height"])

}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}

func Test_com(t *testing.T) {
	addrstr := "0x50fD218932170D14b8293019e57F71a1Eb6df651"
	addr := common.HexToAddress("0x50fd218932170D14b8293019e57F71a1Eb6df651")

	if addr.String() != addrstr {
		t.Log("faild", addr.String())
	} else {
		t.Log("success")
	}

}
