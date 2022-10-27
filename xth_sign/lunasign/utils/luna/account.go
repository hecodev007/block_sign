package luna

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	//"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/terra-project/terra.go/key"
)
func init(){
	//sdkConfig := sdk.GetConfig()
	//sdkConfig.SetCoinType(core.CoinType)
	//sdkConfig.SetFullFundraiserPath(core.FullFundraiserPath)
	//sdkConfig.SetBech32PrefixForAccount(core.Bech32PrefixAccAddr, core.Bech32PrefixAccPub)
	//sdkConfig.SetBech32PrefixForValidator(core.Bech32PrefixValAddr, core.Bech32PrefixValPub)
	//sdkConfig.SetBech32PrefixForConsensusNode(core.Bech32PrefixConsAddr, core.Bech32PrefixConsPub)
	//sdkConfig.SetAddressVerifier(core.AddressVerifier)
	//sdkConfig.Seal()
}


func GenAccount() (address string, private string, err error) {
	privKeyBz := secp256k1.GenPrivKey()
	privKey, err := PrivKeyGen(privKeyBz)
	if err != nil {
		return "","",err
	}
	addr := privKey.PubKey().Address()
	return sdk.AccAddress(addr).String(), hex.EncodeToString(privKeyBz), nil
}


func PrivKeyGen(bz []byte) (types.PrivKey, error) {
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyring.SigningAlgoList{hd.Secp256k1})
	if err != nil {
		return nil, err
	}

	return algo.Generate()(bz), nil
}
