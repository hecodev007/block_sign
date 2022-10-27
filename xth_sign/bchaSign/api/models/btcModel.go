package models

import (
	"github.com/btcsuite/btcd/chaincfg"
	hd "github.com/btcsuite/btcutil/hdkeychain"
)

type BtcModel struct{}

func (m *BtcModel) NewAccount(net *chaincfg.Params) (address string, private string, err error) {
	seed, err := hd.GenerateSeed(64)
	if err != nil {
		return "", "", err
	}
	key, err := hd.NewMaster(seed, net)
	if err != nil {
		return "", "", nil
	}
	addressHash, err := key.Address(net)
	address = addressHash.EncodeAddress()
	return address, key.String(), nil
}
