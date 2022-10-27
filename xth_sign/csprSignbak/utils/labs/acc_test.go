package labs

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"golang.org/x/crypto/blake2b"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
)
func Test_genbtc(t *testing.T){
	privatekey,err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err.Error())
	}
	pri_hex := hex.EncodeToString(privatekey.Serialize())

	pub_hex := hex.EncodeToString((*btcec.PublicKey)(&privatekey.PublicKey).SerializeCompressed())
	t.Log(pri_hex,pub_hex)
}
func Test_accbtc(t *testing.T){
	pri_str := "80fb83c0b703308497180a4b60703e18e1a7db3b86871535bdb0e79a8627d005"
	privateKeyByte, err := hex.DecodeString(pri_str)
	if err != nil {
		panic(err.Error())
	}
	pri ,pub :=btcec.PrivKeyFromBytes(btcec.S256(),privateKeyByte)
	pri_hex := hex.EncodeToString(pri.Serialize())

	//pub_hex := hex.EncodeToString(pub.SerializeCompressed())
	prefix := "secp256k1"
	hash := append([]byte(prefix),0x0)
	hash = append(hash,pub.SerializeCompressed()...)
	sum256 := blake2b.Sum256(hash)

	account_hex := hex.EncodeToString(sum256[:])
	t.Log(pri_hex,pub_hex)
	t.Log(account_hex)
	//0ce7e671d0cd3a20302b03a1d8f12298c2017720542dc41f494e5a2ff62372db
	//93f8e0fcf0b95e16c14dfa1e3482fedef568691c7323c8cd150168f51fe94d76
}
func Test_gen(t *testing.T){
	crypto.S256()
	privatekey,err :=ecdsa.GenerateKey(crypto.S256(),rand.Reader)
	if err != nil {
		panic(err.Error())
	}
	pri_hex := hex.EncodeToString(crypto.FromECDSA(privatekey))
	pub_hex := hex.EncodeToString(crypto.FromECDSAPub(&privatekey.PublicKey))
	t.Log(pri_hex,pub_hex)
}

func Test_acc(t *testing.T){
	pri_str := "80fb83c0b703308497180a4b60703e18e1a7db3b86871535bdb0e79a8627d005"
	privateKeyByte, err := hex.DecodeString(pri_str)
	if err != nil {
		panic(err.Error())
	}
	privateKey, err := crypto.ToECDSA(privateKeyByte)
	if err != nil {
		panic(err.Error())
	}
	pub_hex := hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	t.Log(pub_hex)

}
//f6dee03277e8f99c754bb24bc40ac6b5bf8131890cf1a259953f4ca8332eaa9c 0422add3577c641925f42c9c9b228e4dca1403b9ea4dff66f8720b7910156ccd62113f7cc3e470372774b04be90a71442170bbf80f4e921f746f91de3d79cd6064
//80fb83c0b703308497180a4b60703e18e1a7db3b86871535bdb0e79a8627d005 0466ae6b432e52e285518cc5d98e3605c88c9b5ee147f28a560f3fbb0c7c683cc7b196c271834368479034fc58c8f3ba84101ae3f884dbe6c63e4b5e842bda2045
//