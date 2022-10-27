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

func TestOptionBool_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewOptionBool(NewBool(true)))
	assertRoundtrip(t, NewOptionBool(NewBool(false)))
	assertRoundtrip(t, NewOptionBoolEmpty())
}

func TestOptionBool_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{
		{NewOptionBool(NewBool(false)), 1},
		{NewOptionBool(NewBool(true)), 1},
		{NewOptionBoolEmpty(), 1},
	})
}

func TestOptionBool_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewOptionBool(NewBool(false)), MustHexDecodeString("0x02")},
		{NewOptionBool(NewBool(true)), MustHexDecodeString("0x01")},
		{NewOptionBoolEmpty(), MustHexDecodeString("0x00")},
	})
}

func TestOptionBool_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewOptionBool(NewBool(true)), MustHexDecodeString(
			"0xee155ace9c40292074cb6aff8c9ccdd273c81648ff1149ef36bcea6ebb8a3e25")},
		{NewOptionBool(NewBool(false)), MustHexDecodeString(
			"0xbb30a42c1e62f0afda5f0a4e8a562f7a13a24cea00ee81917b86b89e801314aa")},
		{NewOptionBoolEmpty(), MustHexDecodeString(
			"0x03170a2e7597b7b7e3d84c05391d139a62b157e78786d8c082f29dcf4c111314")},
	})
}

func TestOptionBool_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewOptionBool(NewBool(true)), NewBool(true), false},
		{NewOptionBool(NewBool(false)), NewOptionBool(NewBool(false)), true},
		{NewOptionBoolEmpty(), NewOptionBoolEmpty(), true},
	})
}

func TestOptionBool(t *testing.T) {
	bz := NewOptionBool(NewBool(true))
	assert.False(t, bz.IsNone())
	assert.True(t, bz.IsSome())
	ok, val := bz.Unwrap()
	assert.True(t, ok)
	assert.Equal(t, val, NewBool(true))
	bz.SetNone()
	assert.True(t, bz.IsNone())
	assert.False(t, bz.IsSome())
	ok2, val2 := bz.Unwrap()
	assert.False(t, ok2)
	assert.Equal(t, val2, NewBool(false))
	bz.SetSome(NewBool(false))
	assert.False(t, bz.IsNone())
	assert.True(t, bz.IsSome())
	ok3, val3 := bz.Unwrap()
	assert.True(t, ok3)
	assert.Equal(t, val3, NewBool(false))
}
