package zen

import "fmt"

const (
	TX_NONSTANDARD           = "nonstandard"
	TX_PUBKEY                = "pubkey"
	TX_PUBKEYHASH            = "pubkeyhash"
	TX_SCRIPTHASH            = "scripthash"
	TX_MULTISIG              = "multisig"
	TX_NULL_DATA             = "nulldata"
	TX_WITNESS_V0_KEYHASH    = "witness_v0_keyhash"
	TX_WITNESS_V0_SCRIPTHASH = "witness_v0_scripthash"
	TX_WITNESS_UNKNOWN       = "witness_unknown"
)

type scriptPubkey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

func (sp *scriptPubkey) GetAddress() ([]string, error) {

	switch sp.Type {
	case TX_PUBKEY, TX_PUBKEYHASH, TX_SCRIPTHASH, TX_WITNESS_V0_KEYHASH, TX_WITNESS_V0_SCRIPTHASH,"pubkeyhashreplay":
		return sp.Addresses, nil
	default:
		return nil, fmt.Errorf("don't support tx %s", sp.Type)
	}
}
