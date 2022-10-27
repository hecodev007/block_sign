package models

import (
	"cruSign/common"
	"cruSign/common/conf"
	util "cruSign/utils/crust"
	"cruSign/utils/keystore"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil"
)

//5GvSop24VowjzX49dL2EP1N3PK6JFEEXPh1sDNHr7rmzismK 0x73f4aea1869349a7911086d146d40d314e1b56bbb4bc87b40f55decdfcab1157
var prefix []byte

type CommonModel struct{}

func init() {
	prefix = []byte{conf.GetConfig().Node.Prefix}
}
func (m *CommonModel) NewAccount(num int, MchName, OrderNo string) (adds []string, err error) {
	common.Lock(MchName + "_" + OrderNo)
	defer common.Unlock(MchName + "_" + OrderNo)

	if keystore.Have(conf.GetConfig().Csv.Dir, MchName, OrderNo) {
		return nil, errors.New("address already created")
	}

	var (
		cvsKeysA []*keystore.CsvKey
		cvsKeysB []*keystore.CsvKey
		cvsKeysC []*keystore.CsvKey
		cvsKeysD []*keystore.CsvKey
	)
	for i := 1; i <= num; i++ {
		address, private, err := m.genAccount()
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

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, MchName, OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}

	return adds, nil
}
func (m *CommonModel) genAccount() (address string, wtfPri string, err error) {
	return util.CreateAddress(prefix)
}

func (m *CommonModel) SignTx(param *SignParams) (rawTx string, txid string, err error) {
	pri, err := m.GetPrivate(param.MchName, param.From)
	if err != nil {
		return "", "", err
	}
	//SignTx(from, to string, amount, nonce, fee uint64, pri string, genesisHash, blockHash string, blockNumber uint64, specVersion, transactionVersion uint32, callId string) (string, error) {
	rawTx, err = util.SignTx(param.From, param.To, param.Amount, param.Nonce, param.Fee, string(pri), param.GenesisHash, param.BlockHash,
		param.BlockNumber, param.SpecVersion, param.TransactionVersion, param.CallId)
	return rawTx, "", err
}

func (m *CommonModel) GetAccount(mchName, address string) (*btcutil.WIF, error) {
	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := btcutil.DecodeWIF(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}
}

//获取私钥
func (m *CommonModel) GetPrivate(mchName, address string) (private []byte, err error) {
	if address == "5GvSop24VowjzX49dL2EP1N3PK6JFEEXPh1sDNHr7rmzismK" {
		return []byte("73f4aea1869349a7911086d146d40d314e1b56bbb4bc87b40f55decdfcab1157"), nil
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
