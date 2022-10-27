package models

import (
	"errors"
	"fmt"
	"fioprotocol/common/log"
	"fioprotocol/common"
	"fioprotocol/common/conf"
	"fioprotocol/common/keystore"
	"fioprotocol/common/validator"
	"fioprotocol/utils"
	coin "fioprotocol/utils/okt"
	"strings"

	gosdk "github.com/okex/exchain-go-sdk"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

//from:https://bos.eosn.io/v1/chain/get_info
//mainet
var CHAINID = []byte("\xd5\xa3\xd1\x8f\xbb\x3c\x08\x4e\x3b\x1f\x3f\xa9\x8c\x21\x01\x4b\x5f\x3d\xb5\x36\xcc\x15\xd0\x8f\x9f\x64\x79\x51\x7c\x6a\x3d\x86")

type signParams struct {
	Id          int64           `json:"id,omitempty"`
	FromAddress string          `json:"from_address" binding:"required"`
	ToAddress   string          `json:"to_address" binding:"required"`
	Token       string          `json:"token" binding:"required"`
	Quantity    string          `json:"quantity" binding:"required"`
	Memo        string          `json:"memo,omitempty" binding:"required"`
	SignPubKey  string          `json:"sign_pubkey" binding:"required"`
	BlockID     eos.Checksum256 `json:"block_id" binding:"required"`
}

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
			pub, pri, err := coin.GentAccount()
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
func (m *EosModel) quantityToSymbol(quantity string) (sbol eos.Symbol, err error) {
	quantity = strings.Trim(quantity, " ")
	if len(quantity) == 0 {
		return sbol, fmt.Errorf("quantity cannot be empty")
	}
	parts := strings.Split(quantity, " ")
	if len(parts) <= 1 {
		return sbol, fmt.Errorf("eror quantity: eg. \"1.001 TLOS\"")
	}
	values := strings.Split(parts[0], ".")
	sbol.Symbol = parts[len(parts)-1]
	if len(values) > 1 {
		sbol.Precision = uint8(len(values[1]))
	}
	return

}
func (m *EosModel) SignTx(client gosdk.Client, params *validator.SignParams) (p interface{}, txhash string, err error) {
	return nil, "", nil
}
