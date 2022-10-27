package okt

import (
	"github.com/okex/okchain-go-sdk/utils"
)

func GentAccount() (addr string, pri string, err error) {
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
	return key.GetAddress().String(), pri, nil
}
