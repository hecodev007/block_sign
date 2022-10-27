package okt

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	"strings"
)

func GentAccount2() (addr string, pri string, err error) {
	//st := time.Now()
	mnemonic, err := utils.GenerateMnemonic()
	if err != nil {
		return "", "", err
	}
	//log.Info(time.Since(st))
	pri, err = utils.GeneratePrivateKeyFromMnemo(mnemonic)
	if err != nil {
		return "", "", err
	}
	//log.Info(time.Since(st))
	key, err := utils.CreateAccountWithPrivateKey(pri, "", "")
	if err != nil {
		return "", "", err
	}
	//log.Info(time.Since(st))
	return strings.ToLower(common.BytesToAddress(key.GetAddress()).String()), pri, nil
}

func GentAccount()(addr string,pri string,err error) {
	privkey, err := ethsecp256k1.GenerateKey()
	if err != nil {
		return
	}
	address := ethcrypto.PubkeyToAddress(privkey.ToECDSA().PublicKey)
	return strings.ToLower(address.String()),strings.ToUpper(hexutil.Encode(ethcrypto.FromECDSA(privkey.ToECDSA()))[2:]),nil
}