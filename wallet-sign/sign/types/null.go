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
	"github.com/group-coldwallet/wallet-sign/sign/scale"
)

// Null is a type that does not contain anything (apart from null)
type Null byte

// NewNull creates a new Null type
func NewNull() Null {
	return Null(0x00)
}

// Encode implements encoding for Null, which does nothing
func (n Null) Encode(encoder scale.Encoder) error {
	return nil
}

// Decode implements decoding for Null, which does nothing
func (n *Null) Decode(decoder scale.Decoder) error {
	return nil
}

// String returns a string representation of the value
func (n Null) String() string {
	return ""
}
