package sgb

//
//import (
//	"bytes"
//	crand "crypto/rand"
//	"encoding/hex"
//	"math/big"
//	"testing"
//
//	"github.com/cypherium/cypherBFT/common"
//	"github.com/cypherium/cypherBFT/core/types"
//	"golang.org/x/crypto/ed25519"
//)
//
//func Test_sign2(t *testing.T) {
//	publicKey, privateKey, _ := ed25519.GenerateKey(crand.Reader)
//	tx := types.NewTransaction(100, common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"), big.NewInt(10000000000000000), uint64(21000), big.NewInt(2100000000), []byte{})
//	signer := types.NewEIP155Signer(big.NewInt(200))
//	signedTx, err := types.SignTxWithED25519(tx, signer, privateKey, publicKey)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	var tmpbuf []byte
//	buff := bytes.NewBuffer(tmpbuf)
//	err = signedTx.EncodeRLP(buff)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//	t.Log(hex.EncodeToString(buff.Bytes()))
//}
