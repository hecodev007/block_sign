package models

import (
	"bytes"
	"demo/common"
	"demo/common/conf"
	"demo/common/validator"
	util "demo/utils/dash"
	"demo/utils/keystore"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

type DashModel struct{}

func (m *DashModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
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
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: private}) //string(keystore.Base64Encode([]byte(private)))})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})
	}

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, params.MchName, params.OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return adds, nil
}

func (m *DashModel) Sign(params *validator.SignParams) (rawtx string, err error) {
	wiretx, err := util.BuildTx(params)
	if err != nil {
		return "", err
	}
	for index, _ := range wiretx.TxIn {
		pri, err := m.GetPrivate(params.MchName, params.TxIns[index].FromAddr)
		if err != nil {
			return "", err
		}
		wif, err := btcutil.DecodeWIF(string(pri))
		if err != nil {
			return "", err
		}
		addr, err := btcutil.DecodeAddress(params.TxIns[index].FromAddr, util.NetParams)
		if err != nil {
			return "", err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		signraw, err := txscript.SignatureScript(wiretx, index, pkScript, txscript.SigHashAll, wif.PrivKey, false)
		if err != nil {
			return "", err
		}
		wiretx.TxIn[index].SignatureScript = signraw
	}
	buf := bytes.NewBuffer(make([]byte, 0, wiretx.SerializeSize()))
	wiretx.Serialize(buf)
	//wire.WriteVarBytes(&buf2, 0, wiretx.TxIn[0].SignatureScript)
	return hex.EncodeToString(buf.Bytes()), nil
}

//获取私钥
func (m *DashModel) GetPrivate(mchName, address string) (private []byte, err error) {
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
