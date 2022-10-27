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
	"fmt"
	"io"

	"github.com/group-coldwallet/wallet-sign/sign/scale"
)

// StorageDataRaw contains raw bytes that are not decoded/encoded.
// Be careful using this in your own structs – it only works as the last value in a struct since it will consume the
// remainder of the encoded data. The reason for this is that it does not contain any length encoding, so it would
// not know where to stop.
type StorageDataRaw []byte

// NewStorageDataRaw creates a new StorageDataRaw type
func NewStorageDataRaw(b []byte) StorageDataRaw {
	return StorageDataRaw(b)
}

// Encode implements encoding for StorageDataRaw, which just unwraps the bytes of StorageDataRaw
func (s StorageDataRaw) Encode(encoder scale.Encoder) error {
	return encoder.Write(s)
}

// Decode implements decoding for StorageDataRaw, which just reads all the remaining bytes into StorageDataRaw
func (s *StorageDataRaw) Decode(decoder scale.Decoder) error {
	for i := 0; true; i++ {
		b, err := decoder.ReadOneByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		*s = append((*s)[:i], b)
	}
	return nil
}

// Hex returns a hex string representation of the value
func (s StorageDataRaw) Hex() string {
	return fmt.Sprintf("%#x", s)
}
