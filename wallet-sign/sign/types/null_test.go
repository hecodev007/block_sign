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
)

func TestNull_EncodeDecode(t *testing.T) {
	assertRoundtrip(t, NewNull())
}

func TestNull_EncodedLength(t *testing.T) {
	assertEncodedLength(t, []encodedLengthAssert{
		{NewNull(), 0},
	})
}

func TestNull_Encode(t *testing.T) {
	assertEncode(t, []encodingAssert{
		{NewNull(), MustHexDecodeString("0x")},
	})
}

func TestNull_Hash(t *testing.T) {
	assertHash(t, []hashAssert{
		{NewNull(), MustHexDecodeString(
			"0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")},
	})
}

func TestNull_Hex(t *testing.T) {
	assertEncodeToHex(t, []encodeToHexAssert{
		{NewNull(), ""},
	})
}

func TestNull_String(t *testing.T) {
	assertString(t, []stringAssert{
		{NewNull(), ""},
	})
}

func TestNull_Eq(t *testing.T) {
	assertEq(t, []eqAssert{
		{NewNull(), NewNull(), true},
		{NewNull(), NewBytes([]byte{}), false},
		{NewNull(), NewBool(true), false},
		{NewNull(), NewBool(false), false},
	})
}
