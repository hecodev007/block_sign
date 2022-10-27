package mw

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func Test_addr(t *testing.T) {
	pub := "e9152ec5141124859dd8d5efcbb52b646c051c474ce99f628cb144e90c0c6b31"
	addr, err := PrivateToPub(pub)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)
	addr, err = PrivateToAddr("f3a582c88b5f2c4f2a5cc206344f5de17b69801c056502ee6374ad5fa246200a")
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(addr)
	accountid, err := AddrToAccoutid("CDW-SRDW-PWUB-CZDJ-9LFPT")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(accountid)
}
func Test_getPri(t *testing.T) {
	pri := getPrivate("begin someone house everyone darkness worse hollow guitar sanctuary bubble beam sword")
	t.Log(pri)
	pub, err := PrivateToPub(pri)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(pub)
	signtx, err := Sign([]byte("hello,world"), pri)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(signtx)
}
func Test_sign(t *testing.T) {
	message := "hello,world"
	pri := "e8152ec5141124859dd8d5efcbb52b646c051c474ce99f628cb144e90c0c6b71"
	sm, err := Sign([]byte(message), pri)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(sm)
}
func Test_tx(t *testing.T) {
	Tx := NewTransaction()
	rawtx := "0030390dd107a0052dd12b457c73f1ded3bf0565034c1843ec8048a3d9964eea29de558af03208593ff253cb5ef3993440ef5a070000000000e1f5050000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000001ae58358bdbcee5e011e050e54f1c5516958d577380a223ea317865ff21523083ac2068e34b8237811"
	raw, _ := hex.DecodeString(rawtx)
	Tx.Unseriallize(raw)
	t.Log(len(rawtx), String(Tx))

	return
	Tx.Deadline = 10
	Tx.Timestamp = uint32(time.Now().Unix())
	Tx.FeeNQT = 1000000
	Tx.AmountNQT = 1000000000

	accountid, err := AddrToAccoutid("CDW-SRDW-PWUB-CZDJ-9LFPT")
	if err != nil {
		t.Fatal(err.Error())
	}

	err = Tx.SetRecipient(accountid)
	if err != nil {
		t.Fatal(err.Error())
	}

	pri := "e9152ec5141124859dd8d5efcbb52b646c051c474ce99f628cb144e90c0c6b31"
	pub, err := PrivateToPub(pri)
	if err != nil {
		t.Fatal(err.Error())
	}
	pubytes, _ := hex.DecodeString(pub)
	copy(Tx.PublickKey[:], pubytes[0:32])
	t.Log(String(Tx))
	Tx.Unseriallize(Tx.Seriallize())
	t.Log(String(Tx))

}
func String(v interface{}) string {
	str, _ := json.Marshal(v)
	return string(str)
}
func Test_sha(t *testing.T) {
	addr, err := PubkeyToAddr("f36729029c5acf4b7779ccedee1ab27ec492c28496dad5670fa4e73b95809f30")
	t.Log(addr)
	addr, pri, err := GenAccount()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr, pri)
}

func TestPubkeyToAddr(t *testing.T) {
	pub := "6222d33d433754c0cddf85312fae5502c7d4e79cddb1ebccfb3f3c649ad8da3b"
	pubhash := Sha256(pub)
	t.Log(hex.EncodeToString(pubhash))
	bi := big.NewInt(0)
	//t.Log(Converse(pubhash[0:8]))
	bi.SetBytes(Converse(pubhash[0:8]))

	t.Log(bi.String())
}
