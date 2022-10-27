package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"neoSign/common"
	"neoSign/common/conf"
	"neoSign/common/log"
	"neoSign/common/validator"
	"neoSign/utils/keystore"
	util "neoSign/utils/neo"
	"strings"
)

type NeoModel struct{}

func (m *NeoModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchName + "_" + params.OrderId)
	defer common.Unlock(params.MchName + "_" + params.OrderId)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchName, params.OrderId) {
		return nil, errors.New("address already created")
	}

	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 1; i <= params.Num; i++ {
		address, private, err := util.GenAccount2()
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

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderId); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return adds, nil
}

func (m *NeoModel) Sign(params *validator.SignParams) (rawtx string, txid string, err error) {
	log.Info("sign")
	if params.Type != "" && params.Type!="claim"{
		return "", "", errors.New("错误交易类型")
	}
	if params.Type == "claim"{
		tx,err := util.BuildClaim(params)
		if err != nil {
			return "", "", err
		}
		var pris []string
		pris = pris[0:0]
		for _, in := range params.TxOuts {
			pri, err := m.GetPrivate(params.MchName, in.ToAddr)
			if err != nil {
				return "", "", err
			}
			pris = append(pris, string(pri))
		}
		rawtx, txid, err = util.SignClaim(tx, pris)
		txr,_:=json.Marshal(tx)
		log.Info(string(txr))
		if !strings.HasPrefix(txid,"0x"){
			txid = "0x"+txid
		}
		return rawtx, txid, err
	}
	tx, err := util.BuildTx2(params)
	if err != nil {
		return "", "", err
	}
	var pris []string
	for _, in := range params.TxIns {
		pri, err := m.GetPrivate(params.MchName, in.FromAddr)
		if err != nil {
			return "", "", err
		}
		pris = append(pris, string(pri))
	}
	rawtx, txid, err = util.Sign2(tx, pris)
	txr,_:=json.Marshal(tx)
	log.Info(string(txr))
	if !strings.HasPrefix(txid,"0x"){
		txid = "0x"+txid
	}
	return rawtx, txid, err
}

//获取私钥
func (m *NeoModel) GetPrivate(mchName, address string) (private []byte, err error) {
	return []byte("KyRWDNQUo4RfpFegNcSgEPsoKzAtUSqLw69bNd5Qyv11CxgCSAt5"), nil
	//get mch akey
	log.Info(mchName,address)
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
