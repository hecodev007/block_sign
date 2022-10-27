package v1

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
	"testing"
)

func TestAddr(t *testing.T) {
	//privkey, _ := crypto.GenerateKey()
	//
	//privkey.
	privkey, _ := crypto.HexToECDSA("d2ebea0d40fa7885bdb19fcb90862f68cc723e13e84d3c0561931754cf8806b9")
	//wif := hex.EncodeToString(privkey.D.Bytes())
	address := strings.ToLower(crypto.PubkeyToAddress(privkey.PublicKey).Hex())
	fmt.Println(address)

}
