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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gsrpc "wallet-sign"
	"wallet-sign/sign/config"
	"wallet-sign/sign/types"
)

func TestChain_SubscribeBeefyJustifications(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode.")
	}

	api, err := gsrpc.NewSubstrateAPI(config.Default().RPCURL)
	assert.NoError(t, err)

	ch := make(chan interface{})
	sub, err := api.Client.Subscribe(context.Background(), "beefy", "subscribeJustifications", "unsubscribeJustifications", "justifications", ch)
	if err != nil && err.Error() == "Method not found" {
		t.Skip("skipping since beefy module is not available")
	}

	assert.NoError(t, err)
	defer sub.Unsubscribe()

	timeout := time.After(40 * time.Second)
	received := 0

	for {
		select {
		case msg := <-ch:
			fmt.Printf("encoded msg: %#v\n", msg)

			s := &types.SignedCommitment{}
			err := types.DecodeFromHexString(msg.(string), s)
			if err != nil {
				panic(err)
			}
			fmt.Printf("decoded msg: %#v\n", s)

			received++

			if received >= 2 {
				return
			}
		case <-timeout:
			assert.FailNow(t, "timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}
