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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "wallet-sign/sign/types"
)

func TestExtrinsicEra_Immortal(t *testing.T) {
	var e ExtrinsicEra
	err := DecodeFromHexString("0x00", &e)
	assert.NoError(t, err)
	assert.Equal(t, ExtrinsicEra{IsImmortalEra: true}, e)
}

func TestExtrinsicEra_Mortal(t *testing.T) {
	var e ExtrinsicEra
	err := DecodeFromHexString("0x4e9c", &e)
	assert.NoError(t, err)
	assert.Equal(t, ExtrinsicEra{
		IsMortalEra: true, AsMortalEra: MortalEra{78, 156},
	}, e)
}

func TestExtrinsicEra_EncodeDecode(t *testing.T) {
	var e ExtrinsicEra
	err := DecodeFromHexString("0x4e9c", &e)
	assert.NoError(t, err)
	assertRoundtrip(t, e)
}
