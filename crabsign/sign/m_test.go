package gsrpc

import (
	"encoding/hex"
	"fmt"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"wallet-sign/sign/signature"
	"wallet-sign/sign/types"
	//"wallet-sign/sign/types"
	"math/big"
	"testing"
)

func TestTransfer1(t *testing.T) {
	api, err := NewSubstrateAPI("https://rpc.test.azero.dev/")
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	// Create a call, transferring 12345 units to Bob
	bob, err := types.NewMultiAddressFromHexAccountID("0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48")
	if err != nil {
		panic(err)
	}

	// 1 unit of transfer
	bal, ok := new(big.Int).SetString("100000000000000", 10)
	if !ok {
		panic(fmt.Errorf("failed to convert balance"))
	}

	c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompact(bal))
	if err != nil {
		panic(err)
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}
	//mn := "renew among rocket damage ancient pen rhythm warrior maze mango when session"
	//k, _ := sr25519.NewKeypairFromSeed([]byte(mn))
	//
	//alice,err :=hex.DecodeString("24db192aba6c87b95cbab378c79129d43d73cbc9b00b1c6eac2a19ad4dcd535f")
	//if err!=nil{
	//	fmt.Println(err)
	//}

	//key, err := types.CreateStorageKey(meta, "System", "Account", k.Public().Encode())
	//if err != nil {
	//	panic(err)
	//}
	key, err := types.CreateStorageKey(meta, "System", "Account", signature.TestKeyringPairAlice.PublicKey)
	if err != nil {
		panic(err)
	}
	var accountInfo types.AccountInfo
	ok, err = api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		panic(err)
	}
	fmt.Println("balance:", accountInfo.Data.Free)
	nonce := uint32(accountInfo.Nonce)
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account

	err = ext.Sign(signature.TestKeyringPairAlice, o)
	if err != nil {
		panic(err)
	}

	// Send the extrinsic
	_, err = api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Balance transferred from Alice to Bob: %v\n", bal.String())
}

func TestGenAddr(t *testing.T) {
	mn := "renew among rocket damage ancient pen rhythm warrior maze mango when session"
	//k, _ := sr25519.NewKeypairFromSeed([]byte(mn))
	p, err := signature.KeyringPairFromSecret(mn, 42)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(p.Address)
}

func TestTransfer(t *testing.T) {
	//wss://ws.test.azero.dev
	//	api, err := NewSubstrateAPI("https://rpc.test.azero.dev/")
	api, err := NewSubstrateAPI("wss://ws.test.azero.dev")
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}
	recvPub, err := ss58.DecodeToPub("5DqnJetgimKbn86prxZg68UutGhrTaC36NS7rXi1vzjw5AHJ")
	if err != nil {
		panic(err)
	}
	pub, err := sr25519.NewPublicKey(recvPub)
	fmt.Println(pub.Address())
	bob := types.NewMultiAddressFromAccountID(pub.Encode())
	// Create a call, transferring 12345 units to Bob
	//bob, err := types.NewMultiAddressFromHexAccountID("0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48")
	//if err != nil {
	//	panic(err)
	//}

	// 1 unit of transfer
	bal, ok := new(big.Int).SetString("100000000000000", 10)
	if !ok {
		panic(fmt.Errorf("failed to convert balance"))
	}

	c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompact(bal))
	if err != nil {
		panic(err)
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		panic(err)
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		panic(err)
	}
	mn := "renew among rocket damage ancient pen rhythm warrior maze mango when session"
	//k, _ := sr25519.NewKeypairFromSeed([]byte(mn))
	p, err := signature.KeyringPairFromSecret(mn, 42)
	//
	//alice,err :=hex.DecodeString("24db192aba6c87b95cbab378c79129d43d73cbc9b00b1c6eac2a19ad4dcd535f")
	//if err!=nil{
	//	fmt.Println(err)
	//}

	key, err := types.CreateStorageKey(meta, "System", "Account", p.PublicKey)
	if err != nil {
		panic(err)
	}

	var accountInfo types.AccountInfo
	ok, err = api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		panic(err)
	}
	fmt.Println("balance:", accountInfo.Data.Free)
	nonce := uint32(accountInfo.Nonce)
	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(p, o)
	//err = ext.SignV1(k.Private().Encode(), o)
	if err != nil {
		panic(err)
	}

	// Do the transfer and track the actual status
	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	for {
		status := <-sub.Chan()
		fmt.Printf("Transaction status: %#v\n", status)

		if status.IsInBlock {
			fmt.Printf("Completed at block hash: %#x\n", status.AsInBlock)
			return
		}
	}

	// Send the extrinsic
	//hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Println("txId:",hash.Hex())

	//fmt.Printf("Balance transferred from Alice to Bob: %v\n", bal.String())
}

func TestBlance(t *testing.T) {
	api, err := NewSubstrateAPI("https://rpc.test.azero.dev/")
	if err != nil {
		panic(err)
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		panic(err)
	}

	//alice := signature.TestKeyringPairAlice.PublicKey
	alice, err := hex.DecodeString("24db192aba6c87b95cbab378c79129d43d73cbc9b00b1c6eac2a19ad4dcd535f")
	if err != nil {
		fmt.Println(err)
	}
	key, err := types.CreateStorageKey(meta, "System", "Account", alice)
	if err != nil {
		panic(err)
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		panic(err)
	}

	previous := accountInfo.Data.Free

	fmt.Printf("%#x has a balance of %v\n", alice, previous)
}
