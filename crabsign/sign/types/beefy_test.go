package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"wallet-sign/sign/types"
)

func TestBeefySignature(t *testing.T) {
	empty := types.NewOptionBeefySignatureEmpty()
	assert.True(t, empty.IsNone())
	assert.False(t, empty.IsSome())

	sig := types.NewOptionBeefySignature(types.BeefySignature{})
	sig.SetNone()
	assert.True(t, sig.IsNone())
	sig.SetSome(types.BeefySignature{})
	assert.True(t, sig.IsSome())
	ok, _ := sig.Unwrap()
	assert.True(t, ok)
	assertRoundtrip(t, sig)
}
