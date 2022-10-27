package neputil

import "github.com/zwjlink/neo-thinsdk-go/neo"

func CreateAddr() (address, privkey string) {
	priv, _ := neo.NewSigningKey()
	privkey = neo.PrivateToWIF(priv)
	address = neo.PublicToAddress(&priv.PublicKey)
	return
}

// Validate NEO address
func ValidateNEOAddress(address string) bool {
	//NEO address version is 23
	//https://github.com/neo-project/neo/blob/427a3cd08f61a33e98856e4b4312b8147708105a/neo/protocol.json#L4
	ver, _, err := neo.Base58CheckDecode(address)
	if err != nil {
		return false
	}
	if ver != 23 {
		return false
	}
	return true
}
