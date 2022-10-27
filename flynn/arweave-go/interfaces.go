// Package arweave defines interfaces for interacting with the Arweave Blockchain.
package arweave_go

import "math/big"

// WalletSigner is the interface needed to be able to sign an arweave
type WalletSigner interface {
	Sign(msg []byte) ([]byte, error)
	Verify(msg []byte, sig []byte) error
	Address() string
	PubKeyModulus() *big.Int
}

// BatchChunkerAppName is the application name for the batchchunker. It is added to transaction tags to retrieve them easily.
const BatchChunkerAppName = "arweave-go-batcher"
