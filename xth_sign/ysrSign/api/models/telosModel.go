package models

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	eos "github.com/ystar-foundation/yta-go"
	"github.com/ystar-foundation/yta-go/ecc"
	"github.com/ystar-foundation/yta-go/token"
	"strings"
	"time"
	"ysrSign/common"
	"ysrSign/common/conf"
	"ysrSign/utils/keystore"
)

//from:https://github.com/telosnetwork/node-template/tree/master/mainnet
//testnet
//var CHAINID = []byte("\x1e\xaa\x08\x24\x70\x7c\x8c\x16\xbd\x25\x14\x54\x93\xbf\x06\x2a\xec\xdd\xfe\xb5\x6c\x73\x6f\x6b\xa6\x39\x7f\x31\x95\xf3\x3c\x9f")

//mainet
var CHAINID = []byte("\x9d\x7b\xec\x4b\xf1\x67\xa7\xb1\x36\xd0\xb4\x5d\x8a\xac\x77\xbd\x45\xe7\x61\xe3\x5c\xbd\x2b\x7d\x0e\x88\xdf\xe0\x5e\xbf\x3d\x62")

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

type TelocModel struct{}

func (m *TelocModel) NewAccount(num int, MchName, OrderNo string) (pubs []string, err error) {
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
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: pub, Key: string(keystore.Base64Encode([]byte(pri.String())))})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: pub, Key: ""})
	}
	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, MchName, OrderNo); err != nil {
		return nil, fmt.Errorf("generateCvsABC err: %v", err)
	}
	return pubs, err
}

func (m *TelocModel) GetAccount(mchName, address string) (*ecc.PrivateKey, error) {

	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := ecc.NewPrivateKey(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}

}

//获取私钥
func (m *TelocModel) GetPrivate(mchName, pubkey string) (private []byte, err error) {
	//todo:注释
	if pubkey == "YTA7dzEV7m9cWajox8tUgRt3Cu17kAg5MuMKMeW6cv16NEgMCHuye" {
		return []byte("5Ki5H9RjiyEJng6uVnD1mdoTJss4jZVWU4yMnA3ZyMVE5AgcKuy"), nil
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
func (m *TelocModel) quantityToSymbol(quantity string) (sbol eos.Symbol, err error) {
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
func (m *TelocModel) SignTx(mchName string, pubkey string, txData interface{}) (p interface{}, txhash string, err error) {
	var tx = &eos.Transaction{TransactionHeader: eos.TransactionHeader{MaxCPUUsageMS: 0, MaxNetUsageWords: 0}}
	tx.SetExpiration(3000 * time.Second) //50分钟签名过期，最多只能一个小时

	var txParams = new(signParams)
	txBytes, _ := json.Marshal(txData)
	if err := json.Unmarshal(txBytes, txParams); err != nil {
		return nil, "", err
	}

	tx.RefBlockNum = uint16(binary.BigEndian.Uint32(txParams.BlockID[:4]))
	tx.RefBlockPrefix = binary.LittleEndian.Uint32(txParams.BlockID[8:16])
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
	action.Name = eos.ActN("transfer") //yrctransfer
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

	keyBag := eos.NewKeyBag()
	if err := keyBag.Add(string(wif)); err != nil {
		return nil, "", err
	}
	//log.Info(keyBag.Keys[0].PublicKey().String())
	signTx := eos.NewSignedTransaction(tx)
	fmt.Println(signTx)
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
