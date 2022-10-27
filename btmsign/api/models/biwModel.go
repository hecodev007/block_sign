package models

import (
	"btmSign/bytom/account"
	"btmSign/bytom/blockchain/pseudohsm"
	"btmSign/bytom/blockchain/signers"
	"btmSign/bytom/crypto/ed25519/chainkd"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/protocol/bc/types"
	"btmSign/bytom/protocol/validation"
	"btmSign/bytom/test"
	"btmSign/common"
	"btmSign/common/conf"
	"btmSign/common/log"
	"btmSign/common/validator"
	"btmSign/net"
	util "btmSign/utils/btc"
	"btmSign/utils/keystore"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/status-im/keycard-go/hexutils"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

type BiwModel struct{}

func (m *BiwModel) NewAccount(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchName + "_" + params.OrderId)
	defer common.Unlock(params.MchName + "_" + params.OrderId)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchName, params.OrderId) {
		return nil, errors.New("address already created")
	}

	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 1; i <= params.Num; i++ {
		address, private, err := util.GenAccountBtm()
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

func (m *BiwModel) NewAccountBtm(params *validator.CreateAddressParams) (adds []string, err error) {
	common.Lock(params.MchName + "_" + params.OrderId)
	defer common.Unlock(params.MchName + "_" + params.OrderId)

	if keystore.Have(conf.GetConfig().Csv.Dir, params.MchName, params.OrderId) {
		return nil, errors.New("address already created")
	}

	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey

	for i := 1; i <= params.Num; i++ {
		alias := params.OrderId + strconv.Itoa(i)
		sum := md5.Sum([]byte(alias))
		address, private, err := util.GenAccountBtm2(hexutils.BytesToHex(sum[:]))
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

func (m *BiwModel) Sign(params *validator.SignParams) (rawtx string, err error) {
	wiretx, err := util.BuildTx(params)
	if err != nil {
		return "", err
	}
	for index, _ := range wiretx.TxIn {
		pri, err := m.GetPrivate(params.MchName, params.Ins[index].FromAddr)

		if err != nil {
			return "", err
		}
		wif, err := btcutil.DecodeWIF(string(pri))
		if err != nil {
			return "", err
		}
		addr, err := btcutil.DecodeAddress(params.Ins[index].FromAddr, util.NetParams)
		if err != nil {
			return "", err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		signraw, err := txscript.SignatureScript(wiretx, index, pkScript, txscript.SigHashAll, wif.PrivKey, true)
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

//{"mchId":"placemch5","orderId":"placeorder5","coinName":"sat",
//	"txIns":[{"fromAddr":"sat1qk2yxh2dycq852pje2uzrphut2v5z686z0a6tcf","fromTxid":"edbb5995f0e11b467eb2897d1e622991cc01acaf77a290c6e1f8950b25e3bb9e","fromIndex":0,"fromAmount":5000000000}],
//	"txOuts":[{"toAddr":"sat1qaqyygyl4dcdx33l20dw5z8mttgsm64f7k7vnvh","toAmount":2000000000},
//	{"toAddr":"sat1qk2yxh2dycq852pje2uzrphut2v5z686z0a6tcf","toAmount":2999000000}]}

//address:  bn1ql3j9pdam8e7fd8hhjlyngxz2et0e3nqlt989we
//mnemonic:  lawn stool card saddle venue slot step inspire october bargain myself cool
//encode:  0x6c61776e2073746f6f6c206361726420736164646c652076656e756520736c6f74207374657020696e7370697265206f63746f626572206261726761696e206d7973656c6620636f6f6c
//address:  bn1qyrhc9q8g3m7e7c4szgtltykqhn0ahcl9lr5yfn
//mnemonic:  hope rain ripple bridge fee pizza blade term hurt upset trip rely
//encode:  0x686f7065207261696e20726970706c6520627269646765206665652070697a7a6120626c616465207465726d206875727420757073657420747269702072656c79

//{"base_transaction":null,"actions":[{"account_id":"6b8bd994-960d-4658-897c-4f5fed4c7f5b","amount":1000000,"asset_id":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","type":"spend_account"},
//{"account_id":"6b8bd994-960d-4658-897c-4f5fed4c7f5b","amount":99,"asset_id":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","type":"spend_account"},
//{"amount":99,"asset_id":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","address":"bn1quzkzd38wyraftws4sku8g98f9hdm62e46v9enr","type":"control_address"}],"ttl":0,"time_range": 43432}
func (m *BiwModel) SignBtm2(params *validator.BtmSignParams) (string, error) {
	url := conf.GetConfig().Node.Url
	//id, password, err := m.GetAccount(params.MchName, params.FromAddress)
	id, password, err := m.GetAccountByAddress(params.FromAddress)
	if err != nil {
		return "", err
	}

	listBalancesRequest := net.ListBalancesRequest{
		AccountID: id,
	}
	listBalancesResult, err := net.Post(url+net.ListBalances, listBalancesRequest)
	if err != nil {
		return "", err
	}
	var lbr net.ListBalancesResult
	err = json.Unmarshal([]byte(listBalancesResult), &lbr)
	if err != nil {
		return "", err
	}
	if lbr.Status != "success" {
		return "", errors.New(listBalancesResult)
	}

	dAmount, err := decimal.NewFromString(params.Amount)
	if err != nil {
		return "", err
	}

	var balance int64
	for _, b := range lbr.Data {
		if b.AssetID == "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" && b.AccountID == id {
			balance = b.Amount
		}
	}

	if dAmount.IntPart() > balance {
		return "", errors.New("余额不足")
	}

	listUnspentOutputsRequest := net.ListUnspentOutputsRequest{
		AccountId: id,
	}
	listUnspentOutputsResult, err := net.Post(url+net.ListUnspentOutputs, listUnspentOutputsRequest)
	if err != nil {
		return "", err
	}
	var luor net.ListUnspentOutputsResult
	err = json.Unmarshal([]byte(listUnspentOutputsResult), &luor)
	if err != nil {
		return "", err
	}
	if luor.Status != "success" {
		return "", errors.New(listUnspentOutputsResult)
	}

	var ac []net.Action
	//for _, unspent := range luor.Data {
	//	spend := net.Action{
	//		AccountID: id,
	//		Amount:    unspent.Amount,
	//		AssetID:   "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	//		Type:      "spend_account",
	//	}
	//	ac = append(ac, spend)
	//}

	//449000
	spend := net.Action{
		AccountID: id,
		Amount:    dAmount.IntPart() + 1000000,
		AssetID:   "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		Type:      "spend_account",
	}
	ac = append(ac, spend)
	control := net.Action{
		Address: params.ToAddress,
		Amount:  dAmount.IntPart(),
		AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		Type:    "control_address",
	}
	ac = append(ac, control)

	buildTransactionRequest := net.BuildTransactionRequest{
		Actions:   ac,
		TTL:       1,
		TimeRange: 100000000,
	}
	buildTransactionResult, err := net.Post(url+net.BuildTransaction, buildTransactionRequest)
	if err != nil {
		return "", err
	}
	var btr net.BuildTransactionResult
	err = json.Unmarshal([]byte(buildTransactionResult), &btr)
	if err != nil {
		return "", err
	}
	if btr.Status != "success" {
		return "", errors.New(buildTransactionResult)
	}

	if btr.Data.Fee > 1000000000 {
		return "", errors.New("手续费过高")
	}

	//estimateTransactionGasRequest := net.EstimateTransactionGasRequest{
	//	TransactionTemplate: btr.Data,
	//}
	//estimateTransactionGasResult, err := net.Post(url+net.EstimateTransactionGas, estimateTransactionGasRequest)
	//if err != nil {
	//	return "", err
	//}
	//var etgr net.EstimateTransactionGasResult
	//err = json.Unmarshal([]byte(estimateTransactionGasResult), &etgr)
	//if err != nil {
	//	return "", err
	//}
	//if etgr.Status != "success" {
	//	return "", errors.New(estimateTransactionGasResult)
	//}
	//gas := etgr.Data.TotalNeu
	//var ac2 []net.Action
	//spend2 := net.Action{
	//	AccountID: id,
	//	Amount:    dAmount.IntPart() + gas,
	//	AssetID:   "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	//	Type:      "spend_account",
	//}
	//ac2 = append(ac2, spend2)
	//control2 := net.Action{
	//	Address: params.ToAddress,
	//	Amount:  dAmount.IntPart(),
	//	AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	//	Type:    "control_address",
	//}
	//ac2 = append(ac2, control2)

	signTransactionRequest := net.SignTransactionRequest{
		Password:    password,
		Transaction: btr.Data,
	}
	signTransactionResult, err := net.Post(url+net.SignTransaction, signTransactionRequest)
	if err != nil {
		return "", err
	}
	var str net.SignTransactionResult
	err = json.Unmarshal([]byte(signTransactionResult), &str)
	if err != nil {
		return "", err
	}
	if str.Status != "success" {
		return "", errors.New(signTransactionResult)
	}
	return str.Data.Transaction.RawTransaction, nil
}

func (m *BiwModel) GetAccountByAddress(address string) (accountId string, password string, err error) {
	pri, err := keystore.GetKeyByAddress(address)
	if err != nil {
		return "", "", err
	}
	acc, err := hexutil.Decode(string(pri))
	if err != nil {
		return "", "", err
	}
	accs := strings.Split(string(acc), "#")
	if len(accs) != 3 {
		return "", "", errors.New("获取账户信息失败")
	}
	return accs[0], accs[1], nil
}
func (m *BiwModel) GetAccount(mchId, address string) (accountId string, password string, err error) {
	pri, err := m.GetPrivate(mchId, address)
	if err != nil {
		return "", "", err
	}
	acc, err := hexutil.Decode(string(pri))
	if err != nil {
		return "", "", err
	}
	accs := strings.Split(string(acc), "#")
	if len(accs) != 3 {
		return "", "", errors.New("获取账户信息失败")
	}
	return accs[0], accs[1], nil
}

func (m *BiwModel) SignBtm(params *validator.SignParams) (rawtx string, err error) {
	//todo 手续费判断
	dirPath, err := ioutil.TempDir(".", "P2PKH")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dirPath)
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")
	//chain, _, _, err := test.MakeChain(testDB)
	//chain = chain
	//if err != nil {
	//	return "", err
	//}
	accountManager := account.NewManager(testDB, nil)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		return "", err
	}

	var u []*account.UTXO
	var a []*account.Account
	for _, in := range params.Ins {
		pri, err := m.GetPrivate(params.MchName, in.FromAddr)
		if err != nil {
			return "", err
		}
		mnemonic, err := hexutil.Decode(string(pri))
		if err != nil {
			return "", err
		}
		xpub, _, err := hsm.XCreate2(in.FromAddr, "", "en", string(mnemonic))
		if err != nil {
			return "", err
		}
		accounts, err := accountManager.Create([]chainkd.XPub{xpub.XPub}, 1, in.FromAddr, signers.BIP0044)
		if err != nil {
			return "", err
		}

		controlProg, err := accountManager.CreateAddress(accounts.ID, false)
		if err != nil {
			return "", err
		}
		fmt.Println("in.FromAddr: ", in.FromAddr)
		fmt.Println("controlProg.Address: ", controlProg.Address)
		utxo := test.MakeUTXO2(controlProg, in.FromTxid, uint64(in.FromAmountInt64), uint64(in.FromIndex))
		u = append(u, utxo)
		a = append(a, accounts)
	}
	tpl, tx, err := test.MakeTx2(u, a, params.Outs)
	if err != nil {
		return "", err
	}
	if _, err := test.MakeSign(tpl, hsm, ""); err != nil {
		return "", err
	}
	marshal, err := json.Marshal(tpl)
	if err != nil {
		return "", err
	}
	log.Infof("签名数据: %s", string(marshal))
	var txResult TxData
	err = json.Unmarshal(marshal, &txResult)
	if err != nil {
		return "", err
	}
	fmt.Println("txResult.RawTransaction: ", txResult.RawTransaction)
	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		return "", err
	}
	return string(marshal), nil
}

type TxData struct {
	RawTransaction      string `json:"raw_transaction"`
	SigningInstructions []struct {
		Position          int `json:"position"`
		WitnessComponents []struct {
			Type   string `json:"type"`
			Quorum int    `json:"quorum,omitempty"`
			Keys   []struct {
				Xpub           string   `json:"xpub"`
				DerivationPath []string `json:"derivation_path"`
			} `json:"keys,omitempty"`
			Signatures []string `json:"signatures,omitempty"`
			Value      string   `json:"value,omitempty"`
		} `json:"witness_components"`
	} `json:"signing_instructions"`
	Fee                    int  `json:"fee"`
	AllowAdditionalActions bool `json:"allow_additional_actions"`
}

//获取私钥
func (m *BiwModel) GetPrivate(mchName, address string) (private []byte, err error) {
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
