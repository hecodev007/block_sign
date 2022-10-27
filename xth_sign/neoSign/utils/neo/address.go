package neo

import (
	"crypto/rand"
	"github.com/o3labs/neo-utils/neoutils/btckey"
)
func GenAccount()(address,private string,err error){
	priv, err := btckey.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return priv.ToNeoAddress(),priv.ToWIF(),nil
}