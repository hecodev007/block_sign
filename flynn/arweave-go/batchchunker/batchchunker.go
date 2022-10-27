package batchchunker

import (
	"context"
	"encoding/json"
	"fmt"
	arweave_go "github.com/JFJun/arweave-go"
	"github.com/JFJun/arweave-go/chunker"
	"github.com/JFJun/arweave-go/transactor"
	"github.com/JFJun/arweave-go/wallet"

	"io"
	"strings"
)

const chunkerVersion = "0.0.1"

// BatchMaker struct
type BatchMaker struct {
	ar        *transactor.Transactor
	wallet    *wallet.Wallet
	reader    io.Reader
	totalSize int64
}

// ChunkInformation is the extra data we add to the tags to inform us about the last chunk, it's position, whether it's the
// head of the chunk and it's data version
type ChunkInformation struct {
	PreviousChunk string `json:"previous_chunk"`
	IsHead        bool   `json:"is_head"`
	Version       string `json:"version"`
	Position      int64  `json:"position"`
}

// NewBatch creates a NewBatch struct
func NewBatch(ar *transactor.Transactor, w *wallet.Wallet, reader io.Reader, totalSize int64) *BatchMaker {
	return &BatchMaker{
		ar:        ar,
		wallet:    w,
		reader:    reader,
		totalSize: totalSize,
	}
}

// SendBatchTransaction chunks, sends and waits for all the transactions to get mined
func (b *BatchMaker) SendBatchTransaction() ([]string, error) {
	txList := []string{}
	ch, err := chunker.NewChunker(b.reader, b.totalSize)
	if err != nil {
		return nil, err
	}
	for i := int64(0); i < ch.TotalChunks(); i++ {
		chunk, err := ch.Next()
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(chunk)
		if err != nil {
			return nil, err
		}

		txBuilder, err := b.ar.CreateTransaction(context.TODO(), b.wallet, "0", data, "")
		if err != nil {
			return nil, err
		}
		previousChunk := ""
		if len(txList) > 0 {
			previousChunk = txList[len(txList)-1]
		}
		isHead := false
		if i+1 == ch.TotalChunks() {
			isHead = true
		}
		chunkerInfo := ChunkInformation{
			PreviousChunk: previousChunk,
			IsHead:        isHead,
			Version:       chunkerVersion,
			Position:      chunk.Position,
		}
		tagValue, err := json.Marshal(chunkerInfo)
		if err != nil {
			return nil, err
		}
		txBuilder.AddTag(arweave_go.BatchChunkerAppName, string(tagValue))
		tx, err := txBuilder.Sign(b.wallet)
		if err != nil {
			return nil, err
		}
		resp, err := b.ar.SendTransaction(context.TODO(), tx)
		if err != nil {
			return nil, err
		}
		fmt.Println(resp)
		minedTx, err := b.ar.WaitMined(context.TODO(), tx)
		txList = append(txList, minedTx.Hash())
		fmt.Printf("Successfully sent transaction %d/%d with hash %s \n", chunk.Position+1, ch.TotalChunks(), minedTx.Hash())

	}
	fmt.Printf("Successfully sent batch transactions with head transaction %s and list of transactions: \n - %s \n", txList[len(txList)-1], strings.Join(txList, "\n - "))

	return txList, nil
}
