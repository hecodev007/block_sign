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

package state

import (
	"testing"

	"github.com/group-coldwallet/wallet-sign/sign/types"
	"github.com/stretchr/testify/assert"
)

func TestState_GetKeysLatest(t *testing.T) {
	prefix := types.NewStorageKey(types.MustHexDecodeString(mockSrv.storageKeyHex))[:8]
	keys, err := state.GetKeysLatest(prefix)
	assert.NoError(t, err)
	assert.Equal(t, []types.StorageKey{types.MustHexDecodeString(mockSrv.storageKeyHex)}, keys)
}

func TestState_GetKeys(t *testing.T) {
	prefix := types.NewStorageKey(types.MustHexDecodeString(mockSrv.storageKeyHex))[:8]
	keys, err := state.GetKeys(prefix, mockSrv.blockHashLatest)
	assert.NoError(t, err)
	assert.Equal(t, []types.StorageKey{types.MustHexDecodeString(mockSrv.storageKeyHex)}, keys)
}
