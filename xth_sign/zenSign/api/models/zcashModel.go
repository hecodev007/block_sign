package models

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"zenSign/common"
	"zenSign/common/conf"
	"zenSign/common/log"
	"zenSign/utils/keystore"
	"zenSign/utils/zcash"
	"zenSign/utils/zen"

	"github.com/HorizenOfficial/rosetta-zen/zenutil"

	txscript2 "github.com/HorizenOfficial/rosetta-zen/zend/txscript"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
)

const MaxZCashFee = int64(100000)
const MinZCashFee = int64(1000)

var chaincfgParams = &chaincfg.Params{Name: "main"}

func init() {
	//*chaincfgParams = chaincfg.MainNetParams
}

type ZcashModel struct{}

func (m *ZcashModel) NewAccount(num int, MchName, OrderNo string) (adds []string, err error) {
	//同一个商户keystore保存不能并发
	common.Lock(MchName + "_" + OrderNo)
	defer common.Unlock(MchName + "_" + OrderNo)
	if keystore.Have(conf.GetConfig().Csv.Dir, MchName, OrderNo) {
		//log.Debug("address already created")
		return nil, errors.New("address already created")
	}
	var (
		cvsKeysA []*keystore.CsvKey
		cvsKeysB []*keystore.CsvKey
		cvsKeysC []*keystore.CsvKey
		cvsKeysD []*keystore.CsvKey
	)
	for i := 1; i <= num; i++ {
		if address, private, err := zcash.GenAccount(); err != nil {
			return nil, err
		} else {
			adds = append(adds, address)

			aesKey := keystore.RandBase64Key()
			aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(private), aesKey, true)
			if err != nil {
				return nil, err
			}
			cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: address, Key: string(aesPrivKey)})
			cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: address, Key: string(aesKey)})
			cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: string(keystore.Base64Encode([]byte(private)))})
			cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})
		}

	}
	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, MchName, OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}

	return adds, nil
}
func (m *ZcashModel) genAccount() (address string, wtfPri string, err error) {

	if priv, err := btcec.NewPrivateKey(btcec.S256()); err != nil {
		return "", "", err
	} else if priWif, err := btcutil.NewWIF(priv, &chaincfg.MainNetParams, true); err != nil {
		return "", "", err
	} else if address, err = zecutil.Encode(priWif.PrivKey.PubKey().SerializeCompressed(), chaincfgParams); err != nil {
		return "", "", nil
	} else {
		return address, priWif.String(), nil
	}

}

func (m *ZcashModel) GetAccount(mchName, address string) (*btcutil.WIF, error) {

	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := btcutil.DecodeWIF(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}
}

//获取私钥
func (m *ZcashModel) GetPrivate(mchName, address string) (private []byte, err error) {
	if address == "znSwuhWSWwCDY6ctgPKja5SpYmmctJrErnY" {
		return []byte("18nokGXeCx3CYiJzrFvHyVnEZdqCoqoncsSHsKH5MZbDLdyLLK7b"), nil
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

func (m *ZcashModel) computeFee(txInput *common.Data) int64 {
	var inAmount, outAmount int64
	for _, v := range txInput.TxOuts {
		outAmount += v.ToAmount
	}
	for _, v := range txInput.TxIns {
		inAmount += v.FromAmount
	}
	return inAmount - outAmount
}

func (m *ZcashModel) SignTx2(params *common.ZenSignParams) (rawTx string, txhash string, err error) {

	newTx, err := zcash.BuildRawTx(params, chaincfgParams)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	}

	var signParam = params.Data

	for i := 0; i < len(signParam.TxIns); i++ {
		pubKeyScript, err := hex.DecodeString(signParam.TxIns[i].FromScript)
		if err != nil {
			return "", "", err
		}
		//p2, _ := zcash.PayToAddrScript(signParam.TxIns[i].FromAddr, signParam.BlockHash, signParam.BlockHeight)
		log.Info(hex.EncodeToString(pubKeyScript))
		//log.Info(hex.EncodeToString(p2))
		log.Info(signParam.TxIns[i].FromAddr, signParam.BlockHash, signParam.BlockHeight)
		wif, err := m.GetAccount(params.MchId, signParam.TxIns[i].FromAddr)
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}

		signScrpit, err := txscript.SignatureScript(newTx, 0, pubKeyScript, txscript.SigHashAll,
			wif.PrivKey, true)
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		newTx.TxIn[i].SignatureScript = signScrpit

	}
	var buf bytes.Buffer
	if err = newTx.Serialize(&buf); err != nil {
		log.Info(err.Error())
		return "", "", err
	}

	return hex.EncodeToString(buf.Bytes()), newTx.TxHash().String(), nil
}

func (m *ZcashModel) SignTx(params *common.ZenSignParams) (rawTx string, txhash string, err error) {

	newTx, err := zen.BuildRawTx(params)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	}

	var signParam = params.Data

	for i := 0; i < len(signParam.TxIns); i++ {
		pubKeyScript, err := hex.DecodeString(signParam.TxIns[i].FromScript)
		if err != nil {
			return "", "", err
		}
		//lookupKey := func(a zenutil.Address) (*btcec.PrivateKey, bool, error) {
		//	wif, err := m.GetAccount(params.MchId, a.String())
		//	if err != nil {
		//		return nil, true, err
		//	}
		//	return wif.PrivKey, true, nil
		//}
		wif, err := m.GetAccount(params.MchId, signParam.TxIns[i].FromAddr)
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		wif2, err := zenutil.DecodeWIF(wif.String())
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}

		signScrpit, err := txscript2.SignatureScript(newTx, i, pubKeyScript, txscript2.SigHashAll,
			wif2.PrivKey, true)
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		log.Info(i, len(newTx.TxIn), len(signParam.TxIns))
		newTx.TxIn[i].SignatureScript = signScrpit

	}
	var buf bytes.Buffer
	if err = newTx.Serialize(&buf); err != nil {
		log.Info(err.Error())
		return "", "", err
	}

	return hex.EncodeToString(buf.Bytes()), newTx.TxHash().String(), nil
}
