package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"btmSign/bytom/encoding/blockchain"
	"btmSign/bytom/testutil"
)

func TestReadWriteBlockCommitment(t *testing.T) {
	cases := []struct {
		bc        BlockCommitment
		hexString string
	}{
		{
			bc: BlockCommitment{
				TransactionsMerkleRoot: testutil.MustDecodeHash("35a2d11158f47a5c5267630b2b6cf9e9a5f79a598085a2572a68defeb8013ad2"),
			},
			hexString: "35a2d11158f47a5c5267630b2b6cf9e9a5f79a598085a2572a68defeb8013ad2",
		},
		{
			bc: BlockCommitment{
				TransactionsMerkleRoot: testutil.MustDecodeHash("8ec3ee7589f95eee9b534f71fcd37142bcc839a0dbfe78124df9663827b90c35"),
			},
			hexString: "8ec3ee7589f95eee9b534f71fcd37142bcc839a0dbfe78124df9663827b90c35",
		},
	}

	for _, c := range cases {
		buff := []byte{}
		buffer := bytes.NewBuffer(buff)
		if err := c.bc.writeTo(buffer); err != nil {
			t.Fatal(err)
		}

		hexString := hex.EncodeToString(buffer.Bytes())
		if hexString != c.hexString {
			t.Errorf("test write block commitment fail, got:%s, want:%s", hexString, c.hexString)
		}

		bc := &BlockCommitment{}
		if err := bc.readFrom(blockchain.NewReader(buffer.Bytes())); err != nil {
			t.Fatal(err)
		}

		if !testutil.DeepEqual(*bc, c.bc) {
			t.Errorf("test read block commitment fail, got:%v, want:%v", *bc, c.bc)
		}
	}
}
