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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexDecodeString(t *testing.T) {
	s := HexEncodeToString([]byte{0, 128, 255})
	assert.Equal(t, "0x0080ff", s)

	b, err := HexDecodeString(s)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0, 128, 255}, b)

	b, err = HexDecodeString("0xa")
	assert.NoError(t, err)
	assert.Equal(t, []byte{10}, b)

	b, err = HexDecodeString("f")
	assert.NoError(t, err)
	assert.Equal(t, []byte{15}, b)
}
