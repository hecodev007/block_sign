package models

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"satSign/common"
	"satSign/common/conf"
	"satSign/common/validator"
	util "satSign/utils/btc"
	"satSign/utils/keystore"
)

type SatModel struct{}

func (m *SatModel) ImportAddress(addr, label string, rescan bool) (*util.SatImportAddressResult, error) {
	var (
		result *util.SatImportAddressResult
		err    error
	)
	client := util.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
	result, err = client.ImportAddress(addr, label, rescan)
	if err != nil {
		return nil, err
	}
	return result, err
}
func (m *SatModel) ListUnSpent(addrs []string) (*util.SatListUnSpentResult, error) {
	var (
		result *util.SatListUnSpentResult
		err    error
	)
	client := util.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
	result, err = client.GetUnSpends2(addrs)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (m *SatModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	MchName := params.MchId
	OrderNo := params.OrderId
	num := params.Num
	common.Lock(MchName + "_" + OrderNo)
	defer common.Unlock(MchName + "_" + OrderNo)

	if keystore.Have(conf.GetConfig().Csv.Dir, MchName, OrderNo) {
		//log.Debug("address already created")
		return nil, errors.New("address already created")
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
	//beginTime := time.Now()

	for i := 0; i < num; i++ {
		pub, pri, err := util.GenAccount()
		adds = append(adds, pub)

		if err != nil {
			return nil, err
		}
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

	//endTime := time.Since(beginTime)
	//log.Info("generate %s : %d keys,used time : %d ns", MchName, num, endTime)
	//return fmt.Sprintf("generate %s : %d keys,used time : %f s", MchName, num, endTime.Seconds()), nil

	return adds, nil
}

func (m *SatModel) SignE(params *validator.SignParams) (rawtx string, err error) {
	wiretx, err := util.BuildTx(params)
	if err != nil {
		return "", err
	}

	for index, _ := range wiretx.TxIn {
		pri, err := m.GetPrivate(params.MchId, params.Ins[index].FromAddr)
		if err != nil {
			return "", err
		}
		if _, err := util.SignTx(wiretx, index, params.Ins[index].FromAmountInt64, string(pri)); err != nil {
			return "", err
		}
	}
	//, params.Ins[index].FromAddr
	buf := bytes.NewBuffer(make([]byte, 0, wiretx.SerializeSize()))
	wiretx.Serialize(buf)
	//wire.WriteVarBytes(&buf2, 0, wiretx.TxIn[0].SignatureScript)
	return hex.EncodeToString(buf.Bytes()), nil
}

func (m *SatModel) Sign(params *validator.SignParams) (rawtx string, err error) {
	wiretx, err := util.BuildTx2(params)
	if err != nil {
		return "", err
	}

	for index, _ := range wiretx.TxIn {
		pri, err := m.GetPrivate(params.MchId, params.Ins[index].FromAddr)
		if err != nil {
			return "", err
		}
		if _, err := util.SignTx4(wiretx, index, params.Ins[index].FromAmountInt64, string(pri), params.Ins[index].FromAddr); err != nil {
			return "", err
		}
	}
	//, params.Ins[index].FromAddr
	buf := bytes.NewBuffer(make([]byte, 0, wiretx.SerializeSize()))
	wiretx.Serialize(buf)
	//wire.WriteVarBytes(&buf2, 0, wiretx.TxIn[0].SignatureScript)
	return hex.EncodeToString(buf.Bytes()), nil
}

func (m *SatModel) GetPrivate(mchName, address string) (private []byte, err error) {
	//address, err = util.ToAddr(address)
	//if err != nil {
	//	return nil, err
	//}
	//if private, err = m.getPrivate(mchName, address); err == nil {
	//	return private, err
	//}
	//address, err = util.ToCashAddr(address)
	//if err != nil {
	//	return nil, err
	//}
	return m.getPrivate(mchName, address)
}

//????????????
func (m *SatModel) getPrivate(mchName, address string) (private []byte, err error) {
	//todo:????????????
	//if address == "12Dni1tZ6E6DPtPAmfQ9ey5381ZexVGRvA" {
	//	return []byte("L4SdL6gRUfDfDGg2wDptnPxXgacWcRuWq5xmUwnxWu2fjpagyiwg"), nil
	//}
	//address, err = util.ToCashAddr(address)
	//if err != nil {
	//	return nil, err
	//}
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
