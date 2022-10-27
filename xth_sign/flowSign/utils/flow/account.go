package flow

import (
	"encoding/hex"
	flow "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"crypto/rand"
	"context"
	"errors"
	"github.com/onflow/flow-go-sdk/templates"
)
var (
	creatorAddress    flow.Address
	creatorAccountKey *flow.AccountKey
	creatorSigner     crypto.Signer
)
//http://access.mainnet.nodes.onflow.org:9000/
func GenAccount(client *client.Client)(addr string,pri string,err error){
	ctx := context.Background()
	seed := make([]byte, crypto.MinSeedLength)
	if _, err = rand.Read(seed);err != nil{
		return "","",err
	}
	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil{
		return "","",err
	}
	//account:=flowsdk.NewAccountKey().FromPrivateKey(privateKey)
	publicKey := privateKey.PublicKey()
	accountKey := flow.NewAccountKey().
		SetPublicKey(publicKey). // The signature algorithm is inferred from the public key
		SetHashAlgo(crypto.SHA3_256). // This key will require SHA3 hashes
		SetWeight(flow.AccountKeyWeightThreshold) // Give this key full signing weight

	tx := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, creatorAddress)
	tx.SetPayer(creatorAddress)
	tx.SetProposalKey(
		creatorAddress,
		creatorAccountKey.Index,
		creatorAccountKey.SequenceNumber,
	)
	latestBlock, err := client.GetLatestBlockHeader(context.Background(), true)
	if err != nil{
		return "","",err
	}
	tx.SetReferenceBlockID(latestBlock.ID)
	err = tx.SignEnvelope(creatorAddress, creatorAccountKey.Index, creatorSigner)

	err = client.SendTransaction(context.Background(), *tx)
	result, err := client.GetTransactionResult(ctx, tx.ID())
	if err != nil{
		return "","",err
	}

	var newAddress flow.Address

	if result.Status != flow.TransactionStatusSealed {
		return "","",errors.New("address not known until transaction is sealed")
	}
	for _, event := range result.Events {
		if event.Type == flow.EventAccountCreated {
			newAddress = flow.AccountCreatedEvent(event).Address()
			break
		}
	}
	return newAddress.String(),hex.EncodeToString(privateKey.Encode()),nil
}