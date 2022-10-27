package models

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/eoscanada/eos-go/ecc"
	"marSign/common"
	"marSign/common/conf"
	"marSign/common/keystore"
	"marSign/common/log"
	"marSign/common/validator"
	"marSign/utils"
	util "marSign/utils/mars"
)



type EosModel struct{}

func (m *EosModel) NewAccount(num int, MchName, OrderNo string) (pubs []string, err error) {
	//同一个商户keystore保存不能并发
	common.Lock(MchName + "_" + OrderNo)
	defer common.Unlock(MchName + "_" + OrderNo)
	if keystore.Have(conf.GetConfig().Csv.Dir, MchName, OrderNo) {
		//log.Debug("address already created")
		return nil, errors.New("address already created")
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
	wp := utils.NewWorkPool(10)
	for i := 0; i < num; i++ {
		wp.Incr()
		go func() {
			defer wp.Dec()
			pub, pri, err := util.GenAccount()
			if err != nil {
				panic(err.Error())
			}
			log.Info(pub,pri)
			pubs = append(pubs, pub)
			aesKey := keystore.RandBase64Key()
			aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(pri), aesKey, true)
			if err != nil {
				panic(err.Error())
			}
			log.Info(string(aesPrivKey),string(aesKey))
			common.Lock(MchName + "apend" + OrderNo)
			defer common.Unlock(MchName + "apend" + OrderNo)
			cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: pub, Key: string(aesPrivKey)})
			cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: pub, Key: string(aesKey)})
			cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: pub, Key: pri})
			cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: pub, Key: ""})
		}()
	}
	wp.Wait()

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, MchName, OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return pubs, err
}

func (m *EosModel) GetAccount(mchName, address string) (*ecc.PrivateKey, error) {

	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := ecc.NewPrivateKey(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}

}

//获取私钥
func (m *EosModel) GetPrivate(mchName, pubkey string) (private []byte, err error) {
	//todo:注释
	if pubkey == "SP1MRQDFDNH3AM6VFYCAHFEJ01G68YFWR4MNRA26C"{
		return []byte("9a05741dbfeecbac7055aadceb8e192410c15b45b8379d5b503d29be294d97f1"),nil
	}
	if pubkey == "SP1Z03AMJRDJT1R9DP0X8GJ2ZXNQQG06RYEAGC2EH"{
		return []byte("KyBWnXv2Y7vrBB8QDtq9VgueVRPk8ciTEf23foMT1tPuM2hkvRbG"),nil
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

func (m *EosModel) SignTx(params *validator.SignParams) (p string, txhash string, err error) {
	wiretx, err := util.BuildTx(params)
	if err != nil {
		return "","", err
	}
	for index, _ := range wiretx.TxIn {
		pri, err := m.GetPrivate(params.MchId, params.Ins[index].FromAddr)
		if err != nil {
			return "","", err
		}
		wif, err := btcutil.DecodeWIF(string(pri))
		if err != nil {
			return "","", err
		}
		addr, err := btcutil.DecodeAddress(params.Ins[index].FromAddr, util.NetParams)
		if err != nil {
			return "","", err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "","", err
		}
		signraw, err := txscript.SignatureScript(wiretx, index, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return "","", err
		}
		wiretx.TxIn[index].SignatureScript = signraw
	}
	buf := bytes.NewBuffer(make([]byte, 0, wiretx.SerializeSize()))
	wiretx.Serialize(buf)
	//wire.WriteVarBytes(&buf2, 0, wiretx.TxIn[0].SignatureScript)
	return hex.EncodeToString(buf.Bytes()), wiretx.TxHash().String(),nil
}
