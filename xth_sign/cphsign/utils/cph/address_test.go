package cph

import (
	"cphsign/common/validator"
	"cphsign/utils/sgb"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/ed25519"
)

func Test_addr(t *testing.T) {
	addr, pri, err := GenAccount()

	t.Log(addr, pri, err)
	prikey, err := StringToPrivateKey("1beba0161e1b3b187d235df91aca582f1667fcbe57bb97c499d91230d6d4fd10")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(PubKeyToAddressCypherium(prikey.Public().(ed25519.PublicKey)).String())

}

func Test_pri(t *testing.T) {
	//seed, _ := hex.DecodeString("b1fa2a4469afb2326246b85e66f9e0cfce41882033e6700c24f17ebea63a40fd")
	//pri2 := ed25519.NewKeyFromSeed(seed)
	prikey, err := crypto.HexToECDSA("1beba0161e1b3b187d235df91aca582f1667fcbe57bb97c499d91230d6d4fd10")
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(hex.EncodeToString(crypto.FromECDSAPub(&prikey.PublicKey)))
	//0x7B70Cd3E0Ef59B12aE05f2483b9AdB51ba987234
	t.Log(crypto.PubkeyToAddress(prikey.PublicKey).String())
	//PubKeyToAddressCypherium(prikey.PublicKey.)
}

//068a5405f2ba4d0dfbb8df18702b738ba341fe73ac3d95fa814b9a95e6f63d48
//cc5f03662a50dfdf6f757962ece5c7d98267821289e788974d8db68d0ffb9a972f897090ec7aa47a06a9b13be1157ab2f2e50c0994c50b2c2502d7d2d2286ea7
func Test_secp(t *testing.T) {
	prikey, err := crypto.HexToECDSA("1beba0161e1b3b187d235df91aca582f1667fcbe57bb97c499d91230d6d4fd10")
	if err != nil {
		t.Fatal(err.Error())
	}
	dataBytes := []byte("{\"mch_name\":\"goapi\",\"order_no\":\"btc\",\"coin_name\":\"cph-cph\",\"nonce\":\"0\",\"from_address\":\"CPH7B70Cd3E0Ef59B12aE05f2483b9AdB51ba987234\",\"to_address\":\"CPH7988C24a6a4f46b8886185AC348497E5dFfC97B4\",\"value\":\"100000000000000000\",\"gas_price\":\"120000000000\",\"gas_limit\":\"21000\",\"token\":\"\"}")
	params := new(validator.TelosSignParams)
	err = json.Unmarshal(dataBytes, params)
	if err != nil {
		t.Fatal(err.Error())
	}
	tx := sgb.NewTransaction(uint64(params.Nonce.IntPart()), toCommonAddress(params.ToAddress), params.Value.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), []byte{})
	h := tx.Hash()
	t.Log("0x" + hex.EncodeToString(h[:]))
	sig, err := crypto.Sign(h[:], prikey)
	if err != nil {
		t.Fatal(err.Error())
	}

	signedtx, err := tx.WithSignature(sgb.NewEIP155Signer(nil), sig) //big.NewInt(Chainid)
	if err != nil {
		t.Fatal(err.Error())
	}
	rawtx, err := rlp.EncodeToBytes(signedtx)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("0x" + hex.EncodeToString(rawtx))
}
