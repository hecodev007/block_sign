package crust

import (
	"encoding/hex"
	"fmt"
	sr25519 "github.com/ChainSafe/go-schnorrkel"
	sr255192 "github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/JFJun/substrate-go/ss58"
	"testing"
)

func Test_address(t *testing.T) {
	addr, pri, err := CreateAddress([]byte("1"))
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(addr, pri)

	p := []byte{42}
	addr, err = PrivateToAddress("e0996543e9a7ad0892aaa712fb3e3b0531f55d12f6ae70e4df5c075d43be714c", p)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(addr)
}
func Test_pri(t *testing.T) {
	priBytes, err := hex.DecodeString("73f4aea1869349a7911086d146d40d314e1b56bbb4bc87b40f55decdfcab1157")
	if err != nil {
		panic(err.Error())
	}
	var pri32 [32]byte
	copy(pri32[:], priBytes)

	mnp := new(sr25519.MiniSecretKey)
	err = mnp.Decode(pri32)
	if err != nil {
		panic(err.Error())
	}
	seckey := mnp.ExpandEd25519()
	pub, err := seckey.Public()
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%x\n", pub.Encode())
}
func Test_sign(t *testing.T) {
	priBytes, err := hex.DecodeString("73f4aea1869349a7911086d146d40d314e1b56bbb4bc87b40f55decdfcab1157")
	if err != nil {
		panic(err.Error())
	}
	var pri32 [32]byte
	copy(pri32[:], priBytes)
	private, err := sr25519.NewMiniSecretKeyFromRaw(pri32)
	if err != nil {
		panic(err.Error())
	}
	puben := private.Public().Encode()
	fmt.Printf("%x\n", puben)
	pub, err := ss58.DecodeToPub("5GvSop24VowjzX49dL2EP1N3PK6JFEEXPh1sDNHr7rmzismK")
	if err != nil {
		panic(err.Error())
	}
	addr, err := PubKeyToAddress(pub, []byte{42})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(addr)
	fmt.Printf("%x\n", pub)
	ph, _ := hex.DecodeString("390284d6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d051")
	addre, _ := ss58.Encode(ph, []byte{42})
	fmt.Println(addre)
	pri2, err := sr255192.NewPrivateKey(priBytes)
	if err != nil {
		panic(err.Error())
	}
	pub2, err := pri2.Public()
	if err != nil {
		//panic(err.Error())
	}
	fmt.Printf("%x", pub2.Encode())
	//0x390284d6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d051 83ac460140f3b0b7b6654bdcbcdc6bac7f27e57789055ef71248404c1dbabb0ea27ba27b23bfa0171f59ce912ba5de0d25186da752993c94f5f33ee1348ef3f95b2ae883f500000004033678a39d3bafa04d2c5835080d1f0c1bd8c1d07e9a902947d1bb4346ca25306c0700e8764817
	//0x390284d6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d051 83ac460152b1424affd22e99b97f9a027f413eb71f37aecd934d97dd1e015198a86f05204f5e2a9aee88c88979864eba1800df6ec7024217b5ca6a3a620226fcdbc18c8f9503040004033678a39d3bafa04d2c5835080d1f0c1bd8c1d07e9a902947d1bb4346ca25306c0700e8764817
	//0x39028492e0feb85e225ee7c1800966ab1e69d2e0afe8021304421f9ca9bf5ea9 f9784601b418c283bdadb2243dedb254d46af218c82e51048600432d473c2d0c6621d57860867599803edcc62c2bb57967dbb14b44c096bae7125142e8518be9e17e418df50200000500704f74890129ae1e780a1dcaa27fd395bfce5744d4e377dd197174537df36702070010a5d4e8
	//0xd6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d05183ac46 3678a39d3bafa04d2c5835080d1f0c1bd8c1d07e9a902947d1bb4346ca25306c01000000000000000100000000000000ecc3b4faa9f33ba19d43ee332dc1daa837eecd9fca31dc245cd787a49dd6ac3f3c9c54227536a5bcb481b53c24775ce97098340be175edc9e77c4fbe4e82c881
}
func Test_pub(t *testing.T) {
	ph, _ := hex.DecodeString("390284d6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d051")
	addre, _ := ss58.Encode(ph, []byte{42})
	t.Log(addre)
	t.Log(ss58.EncodeByPubHex("390284d6e08b2f8136867468b31c08ad61401e3310b817bc403315aae8f9d051", []byte{42}))
	t.Log(ss58.EncodeByPubHex("cace3edb9d5fceaa51c5d52a9d0a12b287bc8efa05dfe7704e4a79594d55b17b", []byte{42}))
}
func Test_pria(t *testing.T) {
	addr, err := PrivateToAddress("33a6f3093f158a7109f679410bef1a0c54168145e0cecb4df006c1c2fffb1f09", []byte{42})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(addr)
}
