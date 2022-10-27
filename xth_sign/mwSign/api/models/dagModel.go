package models

import (
	"errors"
	"fmt"
	"mwSign/common"
	"mwSign/common/conf"
	"mwSign/common/validator"
	"mwSign/utils/keystore"
	util "mwSign/utils/mw"
)

type DagModel struct{}

func (m *DagModel) NewAccount(num int, MchName, OrderNo string) (pubs []string, err error) {
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
func (m *DagModel) GetPrivate(mchName, pubkey string) (private []byte, err error) {
	//todo:注释
	if pubkey == "CDW-7JFQ-VBRS-CHH6-7BX3X" {
		return []byte("a8f62a289b29a5b641d5cd555f00ad26603f81d5e6374e81c29217c8e6ecca4d"), nil
	}
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
	unsignTx, err := util.BuildTx(params)
	if err != nil {
		return "", nil, err
	}
	pri, err := m.GetPrivate(params.MchName, params.FromAddress)
	if err != nil {
		return "", nil, err
	}
	txhash, err = util.SignTx(unsignTx, string(pri))
	//txhash2, rawTx2, err := util.SignTx2(unsignTx, string(pri))

	return
}

func (m *DagModel) SendRawTransaction(rawTx string) error {
	return nil
}
