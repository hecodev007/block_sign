package cph

import (
	"cphsign/common/validator"
	"cphsign/utils/sgb"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"golang.org/x/crypto/ed25519"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	//"github.com/cypherium/cypherBFT/crypto"
)

func Test_rpc(t *testing.T) {
	rpc, _ := NewRpcClient("http://13.231.191.20:8000")

	t.Log(rpc.SendRawTransaction("0xf86e8085174876e80082520894244b8cad21a44bb9cc4ec20c8e7ee56960dc701888016345785d8a000080827e68a05f1e7d8cf323314936515cd96355d808dd7e72da050e29cdf6d83438e40795a2a03d44c5020cb3664d28b03551f3085290b92bd95ac94a7b5d4e95f9428dd96309"))
	return
	t.Log(rpc.GetBlockCount())
	t.Log(rpc.GetBalance("0x2f43cd8d4ee652cb8b5750e1c9c83969b2d6b227"))
	t.Log(rpc.SuggestGasPrice())
}

func Test_client(t *testing.T) {
	prihex := "f922eb6f7cad2359dc485b57e0678bd0186e221988132d2f5a8d9bbe9cd52df2"
	dataBytes := []byte("{\"mch_name\":\"goapi\",\"order_no\":\"btc\",\"coin_name\":\"cph-cph\",\"nonce\":\"0\",\"from_address\":\"CPHeb64B4bC1B7dF4923E3d553A837723c263Fa9022\",\"to_address\":\"CPH244B8CaD21a44BB9Cc4Ec20C8e7ee56960Dc7018\",\"value\":\"100000000000000000\",\"gas_price\":\"120000000000\",\"gas_limit\":\"21000\",\"token\":\"\"}")
	params := new(validator.TelosSignParams)
	err := json.Unmarshal(dataBytes, params)
	if err != nil {
		t.Fatal(err.Error())
	}
	prikey, err := StringToPrivateKey(prihex)
	if err != nil {
		t.Fatal(err.Error())
	}

	tx := sgb.NewTransaction(uint64(params.Nonce.IntPart()), toCommonAddress(params.ToAddress), params.Value.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), []byte{})
	s := sgb.NewEIP155Signer(big.NewInt(Chainid))
	signedtx, err := sgb.SignTxWithED25519(tx, s, prikey, prikey.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err.Error())
	}
	rawtx, err := rlp.EncodeToBytes(signedtx)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("0x" + hex.EncodeToString(rawtx))
	t.Log(tx.Hash().String())
}

func toCommonAddress(addr string) (address common.Address) {
	addr = strings.Replace(strings.ToLower(addr), "cph", "0x", 1)
	return common.HexToAddress(addr)
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
