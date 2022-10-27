package sgb

import (
	"bytes"
	"cphsign/common/validator"
	"encoding/hex"
	"encoding/json"
	"testing"

	"golang.org/x/crypto/ed25519"

	"github.com/ethereum/go-ethereum/rlp"
)

func Test_sign(t *testing.T) {
	prihex := "7478369195db60d12b732102ea5ffad49fd90ada9b71e05270112ecb482949e0"
	dataBytes := []byte("{\"mch_name\":\"goapi\",\"order_no\":\"btc\",\"coin_name\":\"cph-cph\",\"nonce\":\"0\",\"from_address\":\"CPHD9b8f50930CE173a66dbC5A43e73830bd9519435\",\"to_address\":\"CPHB5DDd6FB2736a0cFb9c2a27Ac1A29251843225AC\",\"value\":\"1000000000000000\",\"gas_price\":\"154000000000\",\"gas_limit\":\"21000\",\"token\":\"\"}")
	params := new(validator.TelosSignParams)
	err := json.Unmarshal(dataBytes, params)
	if err != nil {
		t.Fatal(err.Error())
	}
	prikey, err := StringToPrivateKey(prihex)
	if err != nil {
		t.Fatal(err.Error())
	}

	tx := NewTransaction(uint64(params.Nonce.IntPart()), ToCommonAddress(params.ToAddress), params.Value.BigInt(), uint64(params.GasLimit.IntPart()), params.GasPrice.BigInt(), []byte{}, 290)
	s := NewEIP155Signer(nil)
	signedtx, err := SignTxWithED25519(tx, s, prikey, prikey.Public().(ed25519.PublicKey))
	var tmpbuf []byte
	buff := bytes.NewBuffer(tmpbuf)
	err = signedtx.EncodeRLP(buff)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log("0x" + hex.EncodeToString(buff.Bytes()))

	rawtx, err := rlp.EncodeToBytes(signedtx)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("0x" + hex.EncodeToString(rawtx))
	t.Log(signedtx.Hash().String())
	t.Log(s.Hash(signedtx).String())
	//t.Log(String(signedtx.WithSignature25519()))
}
func Test_rlp(t *testing.T) {
	rawTx := "f88f820122a07cd1cd8e551e98864f8ac7697c4d3658b8526f09343e40074b10bbe67d44bddd27850430e2340082520894eb64b4bc1b7df4923e3d553a837723c263fa9022888ac7230489e80000801ba0a504f64d24f99b1c7c9dfaab514cf43bec57178f0bd8338439b7de3fc98dfcd69f82c50ab6185b7edb1d4f692028a8aa25069883306c303cf1b3b5961571bf04"
	rawtxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		t.Fatal()
	}
	buff := bytes.NewReader(rawtxBytes)
	tx := new(Transaction)
	stream := rlp.NewStream(buff, uint64(len(rawtxBytes)))
	err = tx.DecodeRLP(stream)
	if err != nil {
		t.Fatal()
	}
	t.Log(String(tx))

	var tmpbuf []byte
	buff2 := bytes.NewBuffer(tmpbuf)
	err = tx.EncodeRLP(buff2)
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}

//0xf88f820122a043758429d382a7d2ee77cb966e4fd4c1d90bee9820c3af255b1f85d5e76f8e7d8085051f4d5c0082520894d9b8f50930ce173a66dbc5a43e73830bd951943587038d7ea4c68000801ba0d1f2be7793f2780458e4912b9ae936a4b9037aed9a9cd956e892e592c3d46039a0d493d2e808600f866abfbc0ad6ebe05a4a96c20017e09413dd7773b3de260d0e
//0xf88f820122a043758429d382a7d2ee77cb966e4fd4c1d90bee9820c3af255b1f85d5e76f8e7d8085051f4d5c0082520894d9b8f50930ce173a66dbc5a43e73830bd951943587038d7ea4c68000801ba0882fc5dfea0e277610265dcc2c25363a5132b799168493a21f19befa758350e5a0ea2ac745ea03516244a9acaefa81d4989e1a9d9491ba531cc0b43cc932e68006
