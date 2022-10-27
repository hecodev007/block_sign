package models

import (
	"chiaSign/common/validator"
	"chiaSign/utils/keystore"
	"fmt"
)

type DagModel struct{}

func (m *DagModel) NewAccount(num int, MchName, OrderNo string) (pubs []string, err error) {
	return nil, err
}

//获取私钥
func (m *DagModel) GetPrivate(mchName, pubkey string) (private []byte, err error) {
	//todo:注释
	//if pubkey == "YTA7dzEV7m9cWajox8tUgRt3Cu17kAg5MuMKMeW6cv16NEgMCHuye" {
	//return []byte("12b13f6bf3354d3a3b7fb740f2cd82dcbbbddb4e9f9e3bfbc5c34518c4f17331"), nil
	//}
	//get mch akey
	if tmpA, err := keystore.KeystoreGetKeyA(mchName, pubkey); err != nil {
		return nil, fmt.Errorf("doesn't find keyA for mch : %s , address : %s", mchName, pubkey)
	} else if akey, err := keystore.Base64Decode([]byte(tmpA)); err != nil {
		return nil, fmt.Errorf("keyA base64 decode err:%v", err)
	} else if bkey, err := keystore.KeystoreGetKeyB(mchName, pubkey); err != nil {
		return nil, fmt.Errorf("doesn't find keyB for mch : %s , address : %s", mchName, pubkey)
	} else if privkey, err := keystore.AesCryptCfb([]byte(akey), []byte(bkey), false); err != nil {
		return nil, fmt.Errorf("aes crypt cfb failed : %s , address : %s", mchName, pubkey)
	} else {
		return privkey, nil
	}

}

func (m *DagModel) SignTx(params *validator.TelosSignParams) (txhash string, rawTx []byte, err error) {

	return
}

func (m *DagModel) SendRawTransaction(rawTx string) error {
	return nil
}
