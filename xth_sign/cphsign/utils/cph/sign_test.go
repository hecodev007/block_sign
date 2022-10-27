package cph

//
//import (
//	"cphsign/common/validator"
//	"encoding/json"
//	"math/big"
//	"testing"
//
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/rlp"
//)
//
//func Test_client2(t *testing.T) {
//	prihex := "f922eb6f7cad2359dc485b57e0678bd0186e221988132d2f5a8d9bbe9cd52df2"
//	dataBytes := []byte("{\"mch_name\":\"goapi\",\"order_no\":\"btc\",\"coin_name\":\"cph-cph\",\"nonce\":\"0\",\"from_address\":\"CPHeb64B4bC1B7dF4923E3d553A837723c263Fa9022\",\"to_address\":\"CPH244B8CaD21a44BB9Cc4Ec20C8e7ee56960Dc7018\",\"value\":\"100000000000000000\",\"gas_price\":\"100000000000\",\"gas_limit\":\"21000\",\"token\":\"\"}")
//	params := new(validator.TelosSignParams)
//	err := json.Unmarshal(dataBytes, params)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	prikey, err := StringToPrivateKey(prihex)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	tx := types.NewTransaction(uint64(params.Nonce.IntPart()), ToCommonAddress(params.ToAddress), params.Value.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), []byte{})
//	s := types.NewEIP155Signer(big.NewInt(Chainid))
//	signedtx, err := types.SignTx(tx, s, prikey)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	rawtx, err := rlp.EncodeToBytes(signedtx)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	t.Log(rawtx)
//	t.Log(tx.Hash().String())
//}
