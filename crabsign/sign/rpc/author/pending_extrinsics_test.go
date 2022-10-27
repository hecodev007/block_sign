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

package author_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	gsrpc "wallet-sign"
	"wallet-sign/sign/config"
)

func TestAuthor_PendingExtrinsics(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	assert.NoError(t, err)
	res, err := api.RPC.Author.PendingExtrinsics()
	assert.NoError(t, err)
	for _, ext := range res {
		fmt.Printf("Pending txn from %v with nonce %v\n", ext.Signature.Signer, ext.Signature.Nonce)
	}
}
