package rylink

import (
	"testing"
)

func TestChangeAddressToBch(t *testing.T) {
	t.Log(ChangeAddressToBch("35noKrGr9sz5zL2cBE1ZJ6PsDnpRYLp8DB"))
}

func TestCreateAddressP2sh(t *testing.T) {
	t.Log(CreateAddressP2sh())
}

func TestCreateAddress(t *testing.T) {
	t.Log(CreateAddress())
}
