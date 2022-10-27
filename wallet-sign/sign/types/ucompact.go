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
	"math/big"

	"github.com/group-coldwallet/wallet-sign/sign/scale"
)

type UCompact big.Int

func NewUCompact(value *big.Int) UCompact {
	return UCompact(*value)
}

func (u *UCompact) Int64() int64 {
	i := big.Int(*u)
	return i.Int64()
}

func NewUCompactFromUInt(value uint64) UCompact {
	return NewUCompact(new(big.Int).SetUint64(value))
}

func (u *UCompact) Decode(decoder scale.Decoder) error {
	ui, err := decoder.DecodeUintCompact()
	if err != nil {
		return err
	}

	*u = UCompact(*ui)
	return nil
}

func (u UCompact) Encode(encoder scale.Encoder) error {
	err := encoder.EncodeUintCompact(big.Int(u))
	if err != nil {
		return err
	}
	return nil
}
