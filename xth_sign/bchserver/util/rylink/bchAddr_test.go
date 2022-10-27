package rylink

import (
	"testing"
)

func TestChangeAddressToBch(t *testing.T) {
	t.Log(ChangeAddressToBch("1533kbeN5BfRWW1nHSeN5QwpNGEmsDPv7C"))
}

func TestCreateAddressP2sh(t *testing.T) {
	t.Log(CreateAddressP2sh())
}

func TestCreateAddress(t *testing.T) {
	t.Log(CreateAddress())
}
