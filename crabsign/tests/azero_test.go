package tests

import (
	"encoding/hex"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"testing"
)

func TestEcdsaCreateKey(t *testing.T) {
	priv, pub, err := crypto.GenerateSubstrateKey(crypto.EcdsaType)
	if err != nil {
		t.Log(err)
		return
	}
	address, err := crypto.CreateSubstrateAddress(pub, ss58.ChainXPrefix)
	if err != nil {
		t.Log(err)
		return
	}
	fmt.Println("PrivateKey: ", hex.EncodeToString(priv))
	fmt.Println("Address1: ", address)
	fmt.Println("===============================================")
	pub2, err := crypto.GenerateSubstrateKeyBySeed(priv, crypto.EcdsaType)
	if err != nil {
		t.Log(err)
		return
	}
	address2, err := crypto.CreateSubstrateAddress(pub2, ss58.ChainXPrefix)
	if err != nil {
		t.Log(err)
		return
	}
	fmt.Println("Address2: ", address2)
}

func TestSr25519CreateKey(t *testing.T) {
	priv, pub, err := crypto.GenerateSubstrateKey(crypto.Sr25519Type)
	if err != nil {
		t.Log(err)
		return
	}
	//address, err := crypto.CreateSubstrateAddress(pub, ss58.PolkadotPrefix)
	address, err := crypto.CreateSubstrateAddress(pub, ss58.SubstratePrefix)
	if err != nil {
		t.Log(err)
		return
	}

	fmt.Println("PrivateKey: ", hex.EncodeToString(priv))
	fmt.Println("Pubkey: ", hex.EncodeToString(pub))
	fmt.Println("Address1: ", address)
	fmt.Println("===============================================")
	pub2, err := crypto.GenerateSubstrateKeyBySeed(priv, crypto.Sr25519Type)
	if err != nil {
		t.Log(err)
		return
	}
	//address2, err := crypto.CreateSubstrateAddress(pub2, ss58.PolkadotPrefix)
	address2, err := crypto.CreateSubstrateAddress(pub2, ss58.SubstratePrefix)
	if err != nil {
		t.Log(err)
		return
	}
	fmt.Println("Address2: ", address2)
}

func TestSign(t *testing.T) {
	priv, pub, err := crypto.GenerateSubstrateKey(crypto.EcdsaType)
	if err != nil {
		t.Log(err)
		return
	}
	sig, err := crypto.Sign(priv, pub, crypto.EcdsaType)
	if err != nil {
		t.Log(err)
		return
	}
	fmt.Println("EcdsaSig: ", hex.EncodeToString(sig))
	fmt.Println(len(sig))
	fmt.Println("=========================================")

	priv2, pub2, err := crypto.GenerateSubstrateKey(crypto.Sr25519Type)
	if err != nil {
		t.Log(err)
		return
	}
	sig2, err := crypto.Sign(priv2, pub2, crypto.Sr25519Type)
	if err != nil {
		t.Log(err)
		return
	}
	fmt.Println("Sr25519Sig: ", hex.EncodeToString(sig2))
	fmt.Println(len(sig2))
}

func Test_DecodeAddress(t *testing.T) {
	address := "12h1EPMr8dt34jLGPzEXsagptARn9xzwruXaCAaXqKQ3GqKc"
	pub, err := ss58.DecodeToPub(address)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(pub)
	fmt.Println(hex.EncodeToString(pub))
}

func Test_EncodeAddress(t *testing.T) {
	pub := "ff188d85f3a34eb6830e0a52299c0451d5a0b615d70967a2db8dee107877fe2d"
	data, _ := hex.DecodeString(pub)
	fmt.Println(len(data))
	address, err := ss58.Encode(data, ss58.PolkadotPrefix)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(address)
}

func Test_ValidAddress(t *testing.T) {
	err := ss58.VerityAddress("GRrZ2nUrLmzRoEWrYBUVN98TVfTKCdCKHNrCKEQnhXbKX1N", ss58.KsmPrefix)
	if err != nil {
		panic(err)
	}
}
