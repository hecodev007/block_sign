package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"google.golang.org/grpc"
	"time"
)

func GenerateKeys(sigAlgoName string) (string, string) {
	seed := make([]byte, crypto.MinSeedLength)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoName)
	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.PublicKey()

	pubKeyHex := hex.EncodeToString(publicKey.Encode())
	privKeyHex := hex.EncodeToString(privateKey.Encode())

	return pubKeyHex, privKeyHex
}

func CreateAccount(node string, publicKeyHex string, sigAlgoName string, hashAlgoName string,
	code string, serviceAddressHex string, servicePrivKeyHex string, serviceSigAlgoName string,
	gasLimit uint64) string {
	ctx := context.Background()

	sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoName)
	publicKey, err := crypto.DecodePublicKeyHex(sigAlgo, publicKeyHex)
	if err != nil {
		panic(err)
	}

	hashAlgo := crypto.StringToHashAlgorithm(hashAlgoName)

	accountKey := flow.NewAccountKey().
		SetPublicKey(publicKey).
		SetSigAlgo(sigAlgo).
		SetHashAlgo(hashAlgo).
		SetWeight(flow.AccountKeyWeightThreshold)

	//accountCode := []byte(nil)
	//if strings.TrimSpace(code) != "" {
	//	accountCode = []byte(code)
	//}

	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	serviceSigAlgo := crypto.StringToSignatureAlgorithm(serviceSigAlgoName)
	servicePrivKey, err := crypto.DecodePrivateKeyHex(serviceSigAlgo, servicePrivKeyHex)
	if err != nil {
		panic(err)
	}

	serviceAddress := flow.HexToAddress(serviceAddressHex)
	serviceAccount, err := c.GetAccountAtLatestBlock(ctx, serviceAddress)
	if err != nil {
		panic(err)
	}
	serviceAccountKey := serviceAccount.Keys[0]
	serviceSigner := crypto.NewInMemorySigner(servicePrivKey, serviceAccountKey.HashAlgo)

	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, serviceAddress)
	tx.SetProposalKey(serviceAddress, serviceAccountKey.Index, serviceAccountKey.SequenceNumber)
	tx.SetPayer(serviceAddress)
	tx.SetGasLimit(uint64(gasLimit))
	block, err := c.GetLatestBlockHeader(ctx, true)
	if err != nil {
		panic(err)
	}
	tx.SetReferenceBlockID(block.ID)
	err = tx.SignEnvelope(serviceAddress, serviceAccountKey.Index, serviceSigner)
	if err != nil {
		panic(err)
	}

	err = c.SendTransaction(ctx, *tx)
	if err != nil {
		panic(err)
	}

	return tx.ID().String()
}

func GetAddress(node string, txIDHex string) string {
	ctx := context.Background()
	c, err := client.New(node, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to node")
	}

	txID := flow.HexToID(txIDHex)
	result, err := c.GetTransactionResult(ctx, txID)
	if err != nil {
		panic("failed to get transaction result")
	}

	var address flow.Address

	if result.Status == flow.TransactionStatusSealed {
		for _, event := range result.Events {
			if event.Type == flow.EventAccountCreated {
				accountCreatedEvent := flow.AccountCreatedEvent(event)
				address = accountCreatedEvent.Address()
			}
		}
	}

	return address.Hex()
}

func main() {
	pubKey, privKey := GenerateKeys("ECDSA_P256")
	fmt.Println(pubKey)
	fmt.Println(privKey)

	node := "127.0.0.1:3569"

	sigAlgoName := "ECDSA_P256"
	hashAlgoName := "SHA3_256"
	code := `
			pub contract HelloWorld {

				pub let greeting: String

				init() {
				    self.greeting = "Hello, World!"
				}

				pub fun hello(): String {
				    return self.greeting
				}
			}
		`

	serviceAddressHex := "f8d6e0586b0a20c7"
	servicePrivKeyHex := "0ab0b3c92adf319ab118f6c073003f7029bb6fa8eb986f47f9b139fbb189e655"
	serviceSigAlgoHex := "ECDSA_P256"

	gasLimit := uint64(100)

	txID := CreateAccount(node, pubKey, sigAlgoName, hashAlgoName, code, serviceAddressHex,
		servicePrivKeyHex, serviceSigAlgoHex, gasLimit)

	fmt.Println(txID)

	blockTime := 10 * time.Second
	time.Sleep(blockTime)

	address := GetAddress(node, txID)
	fmt.Println(address)
}
