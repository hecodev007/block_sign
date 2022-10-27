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

	. "github.com/group-coldwallet/wallet-sign/sign/types"
	"github.com/stretchr/testify/assert"
)

func TestData_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{
		{Data([]byte{12, 251, 42}), 3},
		{Data([]byte{}), 0},
	})
}

func TestData_Encode(t *testing.T) {
	bz := []byte{12, 251, 42}
	data := Data(bz)
	encoded, err := EncodeToBytes(data)
	assert.NoError(t, err)
	assert.Equal(t, bz, encoded)
}

func TestData_Decode(t *testing.T) {
	bz := []byte{12, 251, 42}
	var decoded Data
	err := DecodeFromBytes(bz, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, Data(bz), decoded)
}

func TestData_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{Data([]byte{0, 42, 254}), MustHexDecodeString(
			"0x537db36f5b5970b679a28a3df8d219317d658014fb9c3d409c0c799d8ecf149d")},
		{Data([]byte{}), MustHexDecodeString(
			"0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")},
	})
}

func TestData_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{Data([]byte{0, 0, 0}), "0x000000"},
		{Data([]byte{171, 18, 52}), "0xab1234"},
		{Data([]byte{0, 1}), "0x0001"},
		{Data([]byte{18, 52, 86}), "0x123456"},
	})
}

func TestData_String(t *testing.T) {
	assertString(t, []stringAssert{
		{Data([]byte{0, 0, 0}), "[0 0 0]"},
		{Data([]byte{171, 18, 52}), "[171 18 52]"},
		{Data([]byte{0, 1}), "[0 1]"},
		{Data([]byte{18, 52, 86}), "[18 52 86]"},
	})
}

func TestData_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{Data([]byte{1, 0, 0}), Data([]byte{1, 0}), false},
		{Data([]byte{0, 0, 1}), Data([]byte{0, 1}), false},
		{Data([]byte{0, 0, 0}), Data([]byte{0, 0}), false},
		{Data([]byte{12, 48, 255}), Data([]byte{12, 48, 255}), true},
		{Data([]byte{0}), Data([]byte{0}), true},
		{Data([]byte{1}), NewBool(true), false},
		{Data([]byte{0}), NewBool(false), false},
	})
}
