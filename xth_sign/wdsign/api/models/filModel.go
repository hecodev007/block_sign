package models

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/chain/types"
	//"wdsign/utils/types"

	"github.com/shopspring/decimal"
	"wdsign/common"
	"wdsign/common/conf"
	"wdsign/common/log"
	util "wdsign/utils/fil"
	"wdsign/utils/keystore"
)

type FilModel struct{}

func (m *FilModel) NewAccount(num int, MchName, OrderNo string) (adds []string, err error) {
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
func (m *FilModel) genAccount() (address string, wtfPri string, err error) {
	return util.CreateAddress()
}

//todo:56粉尘找零,fee大小限制,from地址过滤
//todo:toaddr compare
func (m *FilModel) SignTx(param *SignParams) (signedMsg interface{}, rawTx string, txid string, err error) {
	from, err := util.AddrFromString(param.From)
	if err != nil {
		log.Info(err.Error())
		return nil, "", "", err
	}
	fromBuffer := new(bytes.Buffer)
	err = from.MarshalCBOR(fromBuffer)
	log.Info("from", from.String(), hex.EncodeToString(fromBuffer.Bytes()), err)
	//550145c7fdefef0431140e842be3e2688fd2efb98d77
	//550145c7fdefef0431140e842be3e2688fd2efb98d77
	//550145c7fdefef0431140e842be3e2688fd2efb98d77
	to, err := util.AddrFromString(param.To)
	if err != nil {
		return nil, "", "", err
	}
	value, err := decimal.NewFromString(param.Amount)
	if err != nil {
		return nil, "", "", err
	}
	amount := value.BigInt()
	params, _ := base64.StdEncoding.DecodeString("Ynl0ZSBiYXNl")
	msg := &types.Message{
		Version:    0,
		From:       from,
		To:         to,
		Nonce:      uint64(param.Nonce),
		Value:      big.NewFromGo(amount),
		GasPremium: abi.NewTokenAmount(param.GasPremium),
		GasFeeCap:  abi.NewTokenAmount(param.GasFeeCap),
		GasLimit:   param.GasLimit,
		Params:     params,
	}
	//j, _ := msg.MarshalJSON()
	//m2 := (*types.RawMessage)(msg)
	//params,_:=base64.StdEncoding.DecodeString("Ynl0ZSBiYXNl")
	//println(string(j))
	//println(json.Marshal(m2.Params))
	//panic("")
	//获取private
	private, err := m.GetPrivate(param.MchName, param.From)
	if err != nil {
		return nil, "", "", err
	}
	pri, err := hex.DecodeString(string(private))
	if err != nil {
		return nil, "", "", err
	}

	msg.Cid().Bytes()
	sig, err := new(util.SecpSigner).Sign(pri, msg.Cid().Bytes())
	if err != nil {
		return nil, "", "", err
	}
	Signature := &crypto.Signature{
		Type: crypto.SigTypeSecp256k1,
		Data: sig,
	}
	SignedMessage := types.SignedMessage{
		Signature: *Signature,
		Message:   *msg,
	}
	sigBytes, err := SignedMessage.Serialize()
	if err != nil {
		return nil, "", "", nil
	}
	//log.Info(SignedMessage.Message.Cid().String(), SignedMessage.Cid().String())
	return SignedMessage, hex.EncodeToString(sigBytes), SignedMessage.Message.Cid().String(), nil
}

func (m *FilModel) GetAccount(mchName, address string) (*btcutil.WIF, error) {
	if privateWif, err := m.GetPrivate(mchName, address); err != nil {
		return nil, err
	} else if wif, err := btcutil.DecodeWIF(string(privateWif)); err != nil {
		return nil, err
	} else {
		return wif, nil
	}
}

//获取私钥
func (m *FilModel) GetPrivate(mchName, address string) (private []byte, err error) {
	//return []byte("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"), nil
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

//828a0055012ad874611e9ee116675882f61cb45ca4577728050145c7fdefef0431140e842be3e2688fd2efb98d77004200641a000a122044000249f044000186a000405842012f4c72cb76303bbe9e0dc03fc3b272bd95a702669d1b3abd0aa44dbb0ebdde752e8b267745413cd6e9c366e653f9decfdc1ca92e13e596247d4299f443476e8b01
//828a0055012ad874611e9ee116675882f61cb45ca457772806550145c7fdefef0431140e842be3e2688fd2efb98d77004200641a000a122044000249f044000186a000405842012f4c72cb76303bbe9e0dc03fc3b272bd95a702669d1b3abd0aa44dbb0ebdde752e8b267745413cd6e9c366e653f9decfdc1ca92e13e596247d4299f443476e8b01
