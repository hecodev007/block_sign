package fil

import (
	"encoding/hex"
	"github.com/filecoin-project/go-address"
)

const WorldbPrefix = 'w'

func init() {
	address.CurrentNetwork = address.Testnet
}

//这里得到的是testnet 地址
func CreateAddress() (addre string, private string, err error) {
	pri, err := new(SecpSigner).GenPrivate()
	if err != nil {
		return "", "", err
	}
	pub, err := new(SecpSigner).ToPublic(pri)
	if err != nil {
		return "", "", err
	}
	addr, err := address.NewSecp256k1Address(pub)
	if err != nil {
		return "", "", err
	}
	//taddre := []byte(addr.String())
	//taddre[0] = uint8(WorldbPrefix)
	return addr.String(), hex.EncodeToString(pri), nil
}

//func CreateAddress2() (addre string, private string, err error) {
//	pri, err := new(BlsSigner).GenPrivate()
//	if err != nil {
//		return "", "", err
//	}
//	pub, err := new(BlsSigner).ToPublic(pri)
//	if err != nil {
//		return "", "", err
//	}
//	addr, err := address.NewBLSAddress(pub)
//	if err != nil {
//		return "", "", err
//	}
//	return addr.String(), hex.EncodeToString(pri), nil
//}
func AddrFromString(addr string) (address.Address, error) {
	return address.NewFromString(addr)
}
