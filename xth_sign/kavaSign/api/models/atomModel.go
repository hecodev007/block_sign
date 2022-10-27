package models

import (
	"errors"
	"fmt"
	"kavaSign/common"
	"kavaSign/common/conf"
	"kavaSign/common/validator"
	util "kavaSign/utils/kava"
	"kavaSign/utils/keystore"
)

type AtomModel struct{}

func (m *AtomModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchName + "_" + params.OrderNo)
	defer common.Unlock(params.MchName + "_" + params.OrderNo)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchName, params.OrderNo) {
		return nil, errors.New("address already created")
	}

	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 1; i <= params.Num; i++ {
		address, private, err := util.GenAccount()
		if err != nil {
			return nil, err
		}
		adds = append(adds, address)

		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(private), aesKey, true)
		if err != nil {
			return nil, err
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: address, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: address, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: string(keystore.Base64Encode([]byte(private)))}) //string(keystore.Base64Encode([]byte(private)))})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})
	}

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return adds, nil
}

func (m *AtomModel) Sign(params *validator.SignParams) (rawtx string, err error) {
	if params.Data.ChainID == "" {
		params.Data.ChainID = conf.GetConfig().Node.ChainID
	}
	if params.Data.Fee > 10000 {
		return "", errors.New("tx.Fee > 0.1")
	}
	pri, err := m.GetPrivate(params.MchName, params.Data.FromAddr)
	if err != nil {
		return "", err
	}
	return util.SignTx(params, pri)
}

//获取私钥
func (m *AtomModel) GetPrivate(mchName, address string) (private []byte, err error) {
	if tmpA, err := keystore.KeystoreGetKeyA(mchName, address); err != nil {
		return nil, fmt.Errorf("doesn't find keyA for mch : %s , address : %s", mchName, address)
	} else if akey, err := keystore.Base64Decode([]byte(tmpA)); err != nil {
		return nil, fmt.Errorf("keyA base64 decode err:%v", err)
	} else if bkey, err := keystore.KeystoreGetKeyB(mchName, address); err != nil {
		return nil, fmt.Errorf("doesn't find keyB for mch : %s , address : %s", mchName, address)
	} else if privkey, err := keystore.AesCryptCfb([]byte(akey), []byte(bkey), false); err != nil {
		return nil, fmt.Errorf("aes crypt cfb failed : %s , address : %s", mchName, address)
	} else {
		return privkey, nil
	}
}
