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

	"github.com/stretchr/testify/assert"
	"wallet-sign/sign/types"
)

func TestState_QueryStorageLatest(t *testing.T) {
	key := types.NewStorageKey(types.MustHexDecodeString(mockSrv.storageKeyHex))
	hash := types.NewHash(types.MustHexDecodeString("0xdd1816b6f6889f46e23b0d6750bc441af9dad0fda8bae90677c1708d01035fbe"))
	data, err := state.QueryStorageLatest([]types.StorageKey{key}, hash)
	assert.NoError(t, err)
	assert.Equal(t, mockSrv.storageChangeSets, data)
}

func TestState_QueryStorage(t *testing.T) {
	key := types.NewStorageKey(types.MustHexDecodeString(mockSrv.storageKeyHex))
	hash := types.NewHash(types.MustHexDecodeString("0xdd1816b6f6889f46e23b0d6750bc441af9dad0fda8bae90677c1708d01035fbe"))
	data, err := state.QueryStorage([]types.StorageKey{key}, hash, mockSrv.blockHashLatest)
	assert.NoError(t, err)
	assert.Equal(t, mockSrv.storageChangeSets, data)
}
