package models

import (
	"bncsign/common"
	"bncsign/common/conf"
	"bncsign/common/validator"
	util "bncsign/utils/bifrost"
	"bncsign/utils/keystore"
	"errors"
	"fmt"

	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
)

type HdxModel struct{}

func (m *HdxModel) NewAccount(num int, MchName, OrderNo string) (pubs []string, err error) {
	//同一个商户keystore保存不能并发
	common.Lock(MchName + "_" + OrderNo)
	defer common.Unlock(MchName + "_" + OrderNo)
	if keystore.Have(conf.GetConfig().Csv.Dir, MchName, OrderNo) {
		//log.Debug("address already created")
		return nil, errors.New("address already created")
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 0; i < num; i++ {
		pub, pri, err := util.GenAccount()
		if err != nil {
			return nil, err
		}
		pubs = append(pubs, pub)
		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(pri), aesKey, true)
		if err != nil {
			return nil, err
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: pub, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: pub, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: pub, Key: pri})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: pub, Key: ""})
	}
	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, MchName, OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return pubs, err
}

//获取私钥
func (m *HdxModel) GetPrivate(mchName, pubkey string) (private []byte, err error) {
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

func (m *HdxModel) SignTx(params *validator.TelosSignParams) (string, error) {

	pri, err := m.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		return "", err
	}
	extri, err := util.BuildTx(params, string(pri))
	if err != nil {
		return "", err
	}

	return types.EncodeToHexString(extri)
}

func (m *HdxModel) SendRawTransaction(rawTx string) error {
	return nil
}
