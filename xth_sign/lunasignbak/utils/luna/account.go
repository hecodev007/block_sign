package luna


import (
	"encoding/hex"
	//"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-project/terra.go/key"
)

func init() {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
}

func GenAccount2() (address string, private string, err error) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	acc := sdk.AccAddress(pubKey.Address())
	return acc.String(), hex.EncodeToString(privKey[:]), nil
}


func GenAccount() (address string, private string, err error) {
	privKeyBz := secp256k1.GenPrivKey()
	privKey, err := key.PrivKeyGen(privKeyBz)

	addr ,err := privKey.PubKey().Address()
	return addr, privKeyBz, nil
}


func PrivKeyGen(bz []byte) (types.PrivKey, error) {
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyring.SigningAlgoList{hd.Secp256k1})
	if err != nil {
		return nil, err
	}

	return algo.Generate()(bz), nil
}
