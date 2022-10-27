package trx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JFJun/trx-sign-go/genkeys"
	"github.com/JFJun/trx-sign-go/grpcs"
	"github.com/fatih/structs"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/ptypes"
	"math/big"
	"strings"

	"testing"
)

var client, _ = grpcs.NewClient("grpc.trongrid.io:50051")

func TestTrxService_GetBlockByHeight(t *testing.T) {
	lb, err := client.GRPC.GetNowBlock()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(lb.BlockHeader.RawData.Number)
	fmt.Println(string(lb.Blockid))
	fmt.Println(hex.EncodeToString(lb.Blockid))
}

func TestTrxService_GetTxData(t *testing.T) {
	block, err := client.GRPC.GetBlockByNum(1)
	if err != nil || block == nil {
		t.Fatal(err)
	}
	d, _ := json.Marshal(block)
	fmt.Println(string(d))

}

func TestNewScanning(t *testing.T) {
	tx, err := client.GRPC.GetTransactionByID("18761266347b0140b81251be05864b7085e38c0f44bc092bb2d97b99ae9aa764")
	if err != nil {
		t.Fatal(err)
	}
	contracts := tx.GetRawData().GetContract()
	contract := contracts[0]
	var c core.TriggerSmartContract
	if err = ptypes.UnmarshalAny(contract.GetParameter(), &c); err != nil {
		t.Fatal(err)
	}

	tv := structs.Map(c)
	delete(tv, "XXX_NoUnkeyedLiteral")
	delete(tv, "XXX_sizecache")
	delete(tv, "XXX_unrecognized")
	if v, ok := tv["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
		tv["OwnerAddress"] = address.Address(v.([]uint8)).String()
	}
	if v, ok := tv["ToAddress"]; ok && len(v.([]uint8)) > 0 {
		tv["ToAddress"] = address.Address(v.([]uint8)).String()
	}
	if v, ok := tv["ContractAddress"]; ok && len(v.([]uint8)) > 0 {
		tv["ContractAddress"] = address.Address(v.([]uint8)).String()
	}
	if v, ok := tv["Data"]; ok && len(v.([]uint8)) > 0 {
		tv["Data"] = hex.EncodeToString(v.([]uint8))
	}
	if v, ok := tv["AssetName"]; ok && len(v.([]uint8)) > 0 {
		tv["AssetName"] = hex.EncodeToString(v.([]uint8))
	}
	d, _ := json.Marshal(tv)
	fmt.Println(string(d))
	data := tv["Data"].(string)
	fmt.Println(len(data))
	if len(data) != 136 {
		t.Fatal("length is not equal 136")
	}

	if !strings.HasPrefix(data, trc20TransferMethodSignature) {
		t.Fatal("111")
	}
	toAddress := data[len(trc20TransferMethodSignature) : len(trc20TransferMethodSignature)+64]
	amountHex := data[len(data)-64:]
	amount := new(big.Int)
	amount.SetString(amountHex, 16)
	if amount.Sign() < 0 {
		t.Fatal(222)
	}
	fmt.Println(toAddress)
	fmt.Println(genkeys.AddressHexToB58("41" + toAddress[len(amountHex)-40:]))
	fmt.Println(amount.String())
	fmt.Println(genkeys.AddressB58ToHex("TThCjw3z2QYt4G8pgAAcVT1JhvSFBrH4U5"))
}

func TestTrxService_GetTxIsExist(t *testing.T) {
	a := big.NewInt(500)
	d := common.LeftPadBytes(a.Bytes(), 32)
	fmt.Println(d)
	fmt.Println(hex.EncodeToString(d))

}
