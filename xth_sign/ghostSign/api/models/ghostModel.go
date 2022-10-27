package models

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"ghostSign/common"
	"ghostSign/common/conf"
	"ghostSign/common/log"
	"ghostSign/utils/btc"
	"ghostSign/utils/ghost"
	"ghostSign/utils/keystore"
	"ghostSign/utils/zcash"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/iqoption/zecutil"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/ripemd160"
)

type Prevtxs struct {
	Utxos
}
type Utxos []Utxo
type Utxo struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	ScriptPubkey string `json:"scriptPubKey"`
	Amount       string `json:"amount"`
}

const MaxGhostFee = int64(10000000)
const MinGhostFee = int64(1000)

var chaincfgParams = &chaincfg.Params{Name: "mainnet"}

func init() {
	*chaincfgParams = chaincfg.MainNetParams
	chaincfgParams.PubKeyHashAddrID = 0x26
	chaincfgParams.ScriptHashAddrID = 0x61
	chaincfgParams.Bech32HRPSegwit = "ghost"
	chaincfgParams.Net = 0xb4eff2fb
	chaincfgParams.PrivateKeyID = 0xA6
}

type GhostModel struct{}

func (m *GhostModel) NewAccount(num int, MchName, OrderNo string) (adds []string, err error) {
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
func (m *GhostModel) genAccount() (address string, wtfPri string, err error) {
	version := []byte{0x26}
	if priv, err := btcec.NewPrivateKey(btcec.S256()); err != nil {
		return "", "", err
	} else if priWif, err := btcutil.NewWIF(priv, chaincfgParams, true); err != nil {
		return "", "", err
	} else if addrPubKey, err := btcutil.NewAddressPubKey(priWif.PrivKey.PubKey().SerializeCompressed(), chaincfgParams); err != nil {
		return "", "", err
	} else {
		address, err := zecutil.EncodeHash(btcutil.Hash160(addrPubKey.ScriptAddress())[:ripemd160.Size], version)
		return address, priWif.String(), err
	}

}
func (m *GhostModel) computeFee(txInput *zcash.UtxoParams) int64 {
	var inAmount, outAmount int64
	for _, v := range txInput.TxOuts {
		outAmount += v.ToAmount
	}
	for _, v := range txInput.TxIns {
		inAmount += v.FromAmount
	}
	return inAmount - outAmount
}

//todo:56粉尘找零,fee大小限制,from地址过滤
//todo:toaddr compare
func (m *GhostModel) SignTx(mchName string, txData interface{}) (rawTx string, err error) {
	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return "", fmt.Errorf("err sign data:", err.Error())
	}

	newTx, err := ghost.BuildRawTx(txDataBytes, chaincfgParams)
	if err != nil {
		fmt.Println("BuildRawTx")
		return "", err
	}

	var signParam = new(zcash.UtxoParams)
	if err := json.Unmarshal(txDataBytes, signParam); err != nil {
		return "", err
	} else if signParam.Fee < MinGhostFee || signParam.Fee > MaxGhostFee { //手续费判断
		return "", fmt.Errorf("tx.fee between %d and %d", MinGhostFee, MaxGhostFee)
	} else if fee := m.computeFee(signParam); fee < signParam.Fee { //不够指定手续费
		return "", errors.New("insuffient fee")
	} else if fee > signParam.Fee+56 { //有多余的找零,56粉尘额度
		if ChangeAddr, err := ghost.DecodeAddress(signParam.ChangeAddr, chaincfgParams.Name); err != nil {
			return "", errors.New("chanageAddr DecodeAddress error:" + err.Error())
		} else if changeTxScript, err := ghost.PayToPubKeyHashScript(ChangeAddr.ScriptAddress()); err != nil {
			return "", errors.New("chanageAddr PayToAddrScript error:" + err.Error())
		} else {
			newTx.TxOut = append(newTx.TxOut, wire.NewTxOut(fee-signParam.Fee, changeTxScript))
		}
	}
	//txins sign
	for i := 0; i < len(newTx.TxIn); i++ {
		fromAddr, err := ghost.DecodeAddress(signParam.TxIns[i].FromAddr, chaincfgParams.Name)
		if err != nil {
			return "", err
		}
		pkScript, err := ghost.PayToPubKeyHashScript(fromAddr.ScriptAddress())
		if err != nil {
			return "", err
		}
		wif, err := m.GetAccount(mchName, signParam.TxIns[i].FromAddr)
		if err != nil {
			return "", err
		}

		script, err := txscript.SignatureScript(newTx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return "", err
		}
		newTx.TxIn[i].SignatureScript = script
	}
	var buf2 bytes.Buffer

	newTx.Serialize(&buf2)
	wire.WriteVarBytes(&buf2, 0, newTx.TxIn[0].SignatureScript)
	return hex.EncodeToString(buf2.Bytes()), nil
}
func (m *GhostModel) HotSignTx(mchName string, txData interface{}) (rawTx string, err error) {
	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return "", fmt.Errorf("err sign data:", err.Error())
	}
	log.Info("HotSignTx", string(txDataBytes))
	var signParam = new(zcash.UtxoParams)
	if err := json.Unmarshal(txDataBytes, signParam); err != nil {
		return "", err
	}

	client := btc.NewRpcClient(conf.GetConfig().Node.Url, conf.GetConfig().Node.RPCKey, conf.GetConfig().Node.RPCSecret)
	crtq := &btc.CreateRawTransactionRequest{Txouts: make(map[string]string)}
	var allin, allout int64
	var keys []string
	var prevtxs []Utxo
	for i := 0; i < len(signParam.TxIns); i++ {

		tempUtxo := Utxo{Txid: signParam.TxIns[i].FromTxid, Amount: decimal.NewFromInt(signParam.TxIns[i].FromAmount).Div(decimal.NewFromInt(1e8)).String(), Vout: signParam.TxIns[i].FromIndex}
		tmpAddr, err := ghost.DecodeAddress(signParam.TxIns[i].FromAddr, chaincfgParams.Name)
		if err != nil {
			return "", err
		}
		changeTxScript, err := ghost.PayToPubKeyHashScript(tmpAddr.ScriptAddress())
		if err != nil {
			return "", err
		}
		tempUtxo.ScriptPubkey = hex.EncodeToString(changeTxScript)
		prevtxs = append(prevtxs, tempUtxo)
		if signParam.TxIns[i].FromAddr[0:1] != "G" {
			return "", errors.New("unsuport address:" + signParam.TxIns[i].FromAddr)
		}
		prikey, err := m.GetPrivate(mchName, signParam.TxIns[i].FromAddr)
		if err != nil {
			return "", err
		}
		log.Info(string(prikey))
		keys = append(keys, string(prikey))
		allin += signParam.TxIns[i].FromAmount
		txin := btc.Txin{Txid: signParam.TxIns[i].FromTxid, Vout: signParam.TxIns[i].FromIndex}
		crtq.Txins = append(crtq.Txins, txin)
	}
	for i := 0; i < len(signParam.TxOuts); i++ {
		if signParam.TxOuts[i].ToAddr[0:1] != "G" {
			return "", errors.New("unsuport address:" + signParam.TxOuts[i].ToAddr)
		}
		allout += signParam.TxOuts[i].ToAmount
		amount := decimal.NewFromInt(signParam.TxOuts[i].ToAmount).Div(decimal.NewFromInt(1e8))
		crtq.Txouts[signParam.TxOuts[i].ToAddr] = amount.String()
	}
	if allin < allout+signParam.Fee {
		return "", errors.New("insuffient amount of txins")
	}
	if allin > allout+signParam.Fee+56 {

		if signParam.ChangeAddr == "" {
			return "", errors.New("empty changeAddr")
		}
		if signParam.ChangeAddr[0:1] != "G" {
			return "", errors.New("unsuport address:" + signParam.ChangeAddr)
		}
		amount := decimal.NewFromInt(allin - allout - signParam.Fee).Div(decimal.NewFromInt(1e8))
		crtq.Txouts[signParam.ChangeAddr] = amount.String()
	}

	hextx, err := client.CreateRawTransaction(crtq.Txins, crtq.Txouts)
	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	rawtx, err := client.SignRawTransaction(hextx, keys, prevtxs)
	return rawtx.Hex, err
}

func (m *GhostModel) GetAccount(mchName, address string) (*btcutil.WIF, error) {
	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := btcutil.DecodeWIF(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}
}

//获取私钥
func (m *GhostModel) GetPrivate(mchName, address string) (private []byte, err error) {
	if address=="GWbNrKRNVSf63R5gh15StEAiMrWreyccns"{
		return []byte("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"), nil
	}
	if address=="GVJsqdbbwGF4xxtgqzhWbkmJLBuwTfqSuV" {
		return []byte("L2yuSWXr4B2x47dALUbYe8fkqEPbgyY8ZL7XWrJpHjtFKwVR8uKw"),nil
	}
	if address=="GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG"{
		return []byte("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"),nil
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
