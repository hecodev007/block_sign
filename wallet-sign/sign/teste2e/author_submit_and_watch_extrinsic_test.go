// Go Substrate RPC Client (GSRPC) provides APIs and types around Polkadot and any Substrate-based chain RPC calls
//
// Copyright 2019 Centrifuge GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package teste2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/group-coldwallet/wallet-sign/sign/rpc/author"
	"github.com/group-coldwallet/wallet-sign/sign/signature"
	"github.com/group-coldwallet/wallet-sign/sign/types"
	"github.com/stretchr/testify/assert"
)

func TestAuthor_SubmitAndWatchExtrinsic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode.")
	}

	from, ok := signature.LoadKeyringPairFromEnv()
	if !ok {
		from = signature.TestKeyringPairAlice
	}

	api := subscriptionsAPI

	meta, err := api.RPC.State.GetMetadataLatest()
	assert.NoError(t, err)

	bob, err := types.NewMultiAddressFromHexAccountID("0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48")
	assert.NoError(t, err)

	c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompactFromUInt(6969))
	assert.NoError(t, err)

	ext := types.NewExtrinsic(c)
	assert.NoError(t, err)

	era := types.ExtrinsicEra{IsMortalEra: false}
	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	assert.NoError(t, err)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	assert.NoError(t, err)

	key, err := types.CreateStorageKey(meta, "System", "Account", from.PublicKey)
	assert.NoError(t, err)

	var sub *author.ExtrinsicStatusSubscription
	for {

		var accountInfo types.AccountInfo
		ok, err = api.RPC.State.GetStorageLatest(key, &accountInfo)
		assert.NoError(t, err)
		assert.True(t, ok)

		nonce := uint32(accountInfo.Nonce)
		o := types.SignatureOptions{
			// BlockHash:   blockHash,
			BlockHash:          genesisHash, // BlockHash needs to == GenesisHash if era is immortal. // TODO: add an error?
			Era:                era,
			GenesisHash:        genesisHash,
			Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
			SpecVersion:        rv.SpecVersion,
			Tip:                types.NewUCompactFromUInt(0),
			TransactionVersion: rv.TransactionVersion,
		}

		err = ext.Sign(from, o)
		assert.NoError(t, err)

		sub, err = api.RPC.Author.SubmitAndWatchExtrinsic(ext)
		if err != nil {
			continue
		}

		break
	}

	defer sub.Unsubscribe()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case status := <-sub.Chan():
			fmt.Printf("%#v\n", status)

			if status.IsInBlock {
				assert.False(t, types.Eq(status.AsInBlock, types.ExtrinsicStatus{}.AsInBlock),
					"expected AsFinalized not to be empty")
				return
			}
		case <-timeout:
			assert.FailNow(t, "timeout of 10 seconds reached without getting finalized status for extrinsic")
			return
		}
	}
}
