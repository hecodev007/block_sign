package models

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go/ecc"
	"flowSign/common"
	"flowSign/common/conf"
	"flowSign/common/keystore"
	"flowSign/common/log"
	"flowSign/common/validator"
	"flowSign/utils"
	util "flowSign/utils/stx"
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
			pub, pri, err := util.GentAccount()
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

func (m *EosModel) SignTx(params *validator.SignParams) (p []byte, txhash string, err error) {
	tx,err :=util.BuildTx(params)
	if err != nil{
		return nil,"",err
	}
	pri,err :=m.GetPrivate(params.MchName,params.FromAddress)
	if err != nil {
		return nil,"",err
	}
	priBytes ,err := hex.DecodeString(string(pri))
	if err != nil {
		return nil,"",err
	}

	err = tx.Sign(priBytes)
	//rawtx,_:=json.Marshal(tx)
	//log.Info(string(rawtx))
	return tx.Serialize(),tx.Txid(),err
}
