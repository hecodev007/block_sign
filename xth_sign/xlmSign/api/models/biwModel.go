package models

import (
	"errors"
	"fmt"
	"xlmSign/common"
	"xlmSign/common/conf"
	"xlmSign/common/log"
	"xlmSign/common/validator"
	"xlmSign/utils/keystore"
	util "xlmSign/utils/xlm"

	"github.com/stellar/go/txnbuild"
)

type BiwModel struct{}

func (m *BiwModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchId + "_" + params.OrderId)
	defer common.Unlock(params.MchId + "_" + params.OrderId)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchId, params.OrderId) {
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
		log.Info(private)
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: address, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: address, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: private}) //string(keystore.Base64Encode([]byte(private)))})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})
	}

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchId, params.OrderId); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return adds, nil
}

func (m *BiwModel) Sign(params *validator.SignParams) (tx *txnbuild.Transaction, err error) {
	tx, err = util.BuildTx(params)
	if err != nil {
		return nil, err
	}
	pri, err := m.GetPrivate(params.MchId, params.From)
	if err != nil {
		return nil, err
	}
	//log.Info(tx.Base64())
	tx, err = util.SignTx(tx, string(pri))
	return tx, err
}

//获取私钥
func (m *BiwModel) GetPrivate(mchName, address string) (private []byte, err error) {
	//todo:删除调试
	//if address == "12Dni1tZ6E6DPtPAmfQ9ey5381ZexVGRvA" {
	//	return []byte("L4SdL6gRUfDfDGg2wDptnPxXgacWcRuWq5xmUwnxWu2fjpagyiwg"), nil
	//}
	if err != nil {
		return nil, err
	}
	//get mch akey
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
