package models

import (
	"avaxSign/common/validator"
	"errors"
	"fmt"
	"log"

	"avaxSign/common"
	"avaxSign/common/conf"
	"avaxSign/utils/avax"
	"avaxSign/utils/keystore"
	"github.com/ava-labs/gecko/ids"
)

var ASSETID = ""

func init() {
	ASSETID = conf.Cfg.Node.AssetID
}

type AvaxModel struct{}

func (m *AvaxModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchName + "_" + params.OrderNo)
	defer common.Unlock(params.MchName + "_" + params.OrderNo)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchName, params.OrderNo) {
		return nil, errors.New("address already created")
	}

	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 1; i <= params.Num; i++ {
		address, private, err := avax.GenAccount()
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
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: private}) //string(keystore.Base64Encode([]byte(private)))})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})
	}

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return adds, nil
}

func (m *AvaxModel) Sign(params *validator.SignParams) (rawtx string, err error) {
	//privkey, err = m.GetPrivate(params.MchName, params.FromAddr)
	//if err != nil {
	//	return "", err
	//}
	w, err := avax.NewWallet("", params.Fee)
	if err != nil {
		return "", err
	}
	if err := w.SetChangeAddr(params.ChangeAddr); err != nil {
		log.Println(err.Error())
		return "", err
	}
	if err = avax.AddUtxos(w, params.Utxos, params.MchName); err != nil {
		log.Println(err.Error())
		return "", err
	}
	toAddr, err := avax.AddressToShot(params.ToAddr)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	assetID, err := ids.FromString(ASSETID)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	tx, err := w.CreateTx(assetID, params.Amount, *toAddr)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return avax.MarsonTx(tx)
}

//获取私钥
func (m *AvaxModel) GetPrivate(mchName, address string) (private []byte, err error) {
	//return []byte("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"), nil
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
