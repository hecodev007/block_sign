// Generated by: main
// TypeWriter: wrapper
// Directive: +gen on SignatureInner

package crypto

import (
	"btmSign/bytom/lib/github.com/tendermint/go-wire/data"
)

// Auto-generated adapters for happily unmarshaling interfaces
// Apache License 2.0
// Copyright (c) 2017 Ethan Frey (ethan.frey@tendermint.com)

type Signature struct {
	SignatureInner "json:\"unwrap\""
}

var SignatureMapper = data.NewMapper(Signature{})

func (h Signature) MarshalJSON() ([]byte, error) {
	return SignatureMapper.ToJSON(h.SignatureInner)
}

func (h *Signature) UnmarshalJSON(data []byte) (err error) {
	parsed, err := SignatureMapper.FromJSON(data)
	if err == nil && parsed != nil {
		h.SignatureInner = parsed.(SignatureInner)
	}
	return err
}

// Unwrap recovers the concrete interface safely (regardless of levels of embeds)
func (h Signature) Unwrap() SignatureInner {
	hi := h.SignatureInner
	for wrap, ok := hi.(Signature); ok; wrap, ok = hi.(Signature) {
		hi = wrap.SignatureInner
	}
	return hi
}

/*** below are bindings for each implementation ***/

func init() {
	SignatureMapper.RegisterImplementation(SignatureEd25519{}, "ed25519", 0x1)
}

func (hi SignatureEd25519) Wrap() Signature {
	return Signature{hi}
}
