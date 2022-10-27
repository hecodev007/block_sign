package ghost

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type ChainParams struct {
	PubHashPrefixes    []byte
	ScriptHashPrefixes []byte
}

var net = ChainParams{
	PubHashPrefixes:    []byte{0x26},
	ScriptHashPrefixes: []byte{0x1C, 0xBD},
}

type Address struct {
	hash    [ripemd160.Size]byte
	Address string
}

func (addr *Address) String() string {
	return addr.Address
}
func (addr *Address) EncodeAddress() string {
	return addr.Address
}
func (addr *Address) ScriptAddress() []byte {
	return addr.hash[:]
}
func (addr *Address) IsForNet(chainParam *chaincfg.Params) bool {
	return true
}
func DecodeAddress(address string, netName string) (btcutil.Address, error) {
	var decoded = base58.Decode(address)
	if len(decoded) != 25 {
		return nil, base58.ErrInvalidFormat
	}

	var cksum [4]byte
	copy(cksum[:], decoded[len(decoded)-4:])

	if addrChecksum(decoded[:len(decoded)-4]) != cksum {
		return nil, base58.ErrChecksum
	}

	if len(decoded)-5 != ripemd160.Size {
		return nil, errors.New("incorrect payload len")
	}
	fmt.Printf(address+"address decode:%x\n", decoded[1:len(decoded)-4])
	addr := new(Address)
	copy(addr.hash[:], decoded[1:len(decoded)-4])
	addr.Address = address
	return addr, nil
}
func addrChecksum(input []byte) (cksum [4]byte) {
	var (
		h  = sha256.Sum256(input)
		h2 = sha256.Sum256(h[:])
	)

	copy(cksum[:], h2[:4])

	return
}
func PayToPubKeyHashScript(pubKeyHash []byte) ([]byte, error) {
	return txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
}
