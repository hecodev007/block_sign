package models

import (
	"bosSign/common"
	"bosSign/common/conf"
	"bosSign/common/log"
	"bosSign/common/validator"
	"bosSign/utils/keystore"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
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

	for i := 0; i < num; i++ {
		pri, err := ecc.NewRandomPrivateKey()
		if err != nil {
			return nil, err
		}
		pub := pri.PublicKey().String()
		pubs = append(pubs, pub)
		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(pri.String()), aesKey, true)
		if err != nil {
			return nil, err
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: pub, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: pub, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: pub, Key: pri.String()})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: pub, Key: ""})
	}
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
	if pubkey == "EOS6VeUZo93nzcmhK3HfQaXBsiw9tsd6hPfU2QwS2adpYQqM9G2Rt" {
		return []byte("5JFnwrLsvo6nmCRPQ2636U2zygHZ9nj2YrNHh5WrTyxC4vwJ9q7"), nil
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

func (m *EosModel) SignTx(mchName string, pubkey string, txParams *validator.SignParams_Data) (p interface{}, txhash string, err error) {
	var tx = &eos.Transaction{TransactionHeader: eos.TransactionHeader{MaxCPUUsageMS: 0, MaxNetUsageWords: 0}}
	tx.SetExpiration(3000 * time.Second) //50分钟签名过期，最多只能一个小时
	tx.Expiration = eos.JSONTime{time.Unix(1641386661, 0)}
	log.Info(tx.Expiration)
	log.Info(hex.EncodeToString(CHAINID))
	blockid, err := hex.DecodeString(txParams.BlockID)
	if err != nil {
		return nil, "", err
	}

	tx.RefBlockNum = uint16(binary.BigEndian.Uint32(blockid[:4]))
	tx.RefBlockPrefix = binary.LittleEndian.Uint32(blockid[8:16])
	log.Info(uint16(binary.BigEndian.Uint32(blockid[:4])), binary.LittleEndian.Uint32(blockid[8:16]))
	m.quantityToSymbol(txParams.Quantity)
	telosSymbol, err := m.quantityToSymbol(txParams.Quantity)
	if err != nil {
		return nil, "", err
	}
	quantity, err := eos.NewFixedSymbolAssetFromString(telosSymbol, txParams.Quantity)
	if err != nil {
		return nil, "", err
	}
	action := token.NewTransfer(eos.AN(txParams.FromAddress), eos.AN(txParams.ToAddress), quantity, txParams.Memo)
	action.Account = eos.AN(txParams.Token)
	action.Name = eos.ActN("transfer")
	action.Authorization = []eos.PermissionLevel{
		{Actor: eos.AN(txParams.FromAddress), Permission: eos.PN("active")},
	}
	actions := []*eos.Action{action}

	////////// todo:删除
	//var action2 eos.Action = *action
	//action2.Account = eos.AN("111111111111")
	//actions = append(actions, &action2)
	/////////

	tx.Actions = actions
	eccPubKey, err := ecc.NewPublicKey(pubkey)
	if err != nil {
		return nil, "", err
	}

	wif, err := m.GetPrivate(mchName, pubkey)
	if err != nil {
		return nil, "", err
	}
	log.Info(wif)
	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(string(wif)); err != nil {
		return nil, "", err
	}

	signTx := eos.NewSignedTransaction(tx)
	//签名
	if signTx, err = keyBag.Sign(signTx, CHAINID, eccPubKey); err != nil {
		return nil, "", err
		//打包
	} else if pack, err := signTx.Pack(eos.CompressionNone); err != nil {
		//fmt.Println("Pack")
		return nil, "", err
	} else if txhash, err := pack.ID(); err != nil {
		return nil, "", err
	} else {
		return pack, txhash.String(), nil
	}
}
