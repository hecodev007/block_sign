package models

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"github.com/shopspring/decimal"
	"iotasign/common"
	"iotasign/common/conf"
	"iotasign/common/validator"
	"iotasign/utils/keystore"
)

type SatModel struct{}

type IOTAUnspent struct {
	//Input                  *iotago.UTXOInput
	Txid         string          `json:"txid"`
	Vout         uint16          `json:"vout"`
	Address      string          `json:"address"`
	AmountInt64  decimal.Decimal `json:"amount"`
	ScriptPubKey string          `json:"scriptPubKey"`
}

func (m *SatModel) ListUnSpent(addrs []string) ([]*IOTAUnspent, error) {
	var (
		err error
	)
	//"atoi1qrhacyfwlcnzkvzteumekfkrrwks98mpdm37cj4xx3drvmjvnep6x8x4r7t"
	list := make([]*IOTAUnspent, 0)

	for _, add := range addrs {
		_, addr, _ := iotago.ParseBech32(add)

		edAddr, err := iotago.ParseEd25519AddressFromHexString(addr.String())
		if err != nil {
			return nil, err
		}
		nodeAPI := iotago.NewNodeHTTPAPIClient(conf.GetConfig().Node.Url)
		resp, err := nodeAPI.OutputIDsByEd25519Address(context.Background(), edAddr, true)
		if err != nil {
			return nil, err
		}
		for _, id := range resp.OutputIDs {
			unspent := &IOTAUnspent{}
			outputRes, err := nodeAPI.OutputByID(context.Background(), id.MustAsUTXOInput().ID())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if outputRes.Spent {
				continue
			}
			unspent.Vout = id.MustAsUTXOInput().TransactionOutputIndex
			unspent.Txid = hex.EncodeToString(id.MustAsUTXOInput().TransactionID[:])
			output, err := outputRes.Output()
			if err != nil {
				continue
			}
			value, err := output.Deposit()
			if err != nil {
				continue
			}

			tar, err := output.Target()
			if err != nil {

				continue
			}
			addr := fmt.Sprintf("%v", tar)
			edAddr, err := iotago.ParseEd25519AddressFromHexString(addr)
			unspent.Address = edAddr.Bech32(iotago.PrefixMainnet)
			//unspent.Spent = outputRes.Spent
			unspent.AmountInt64 = decimal.NewFromInt(int64(value))
			list = append(list, unspent)

		}
	}

	return list, err
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
		_, pri, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}
		//one := hex.EncodeToString(identityOne)
		//two, _ := hex.DecodeString(one)

		inputAddr := iotago.AddressFromEd25519PubKey(pri.Public().(ed25519.PublicKey))
		//fmt.Println(inputAddr.String())
		//addrKeys := iotago.AddressKeys{Address: &inputAddr, Keys: pri}
		//fmt.Println(inputAddr.Bech32(iotago.PrefixTestnet))
		addr := inputAddr.Bech32(iotago.PrefixMainnet)
		//pub, pri, err := util.GenAccount()
		adds = append(adds, addr)

		aesKey := keystore.RandBase64Key()
		aesPrivKey, err := keystore.AesBase64CryptCfb(pri, aesKey, true)
		if err != nil {
			return nil, err
		}
		cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: addr, Key: string(aesPrivKey)})
		cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: addr, Key: string(aesKey)})
		cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: addr, Key: hex.EncodeToString(pri)})
		cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: addr, Key: ""})
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
	return "", nil
}

func (m *SatModel) Sign(params *validator.SignParams) (rawtx string, err error) {

	//
	//resp, err := nodeAPI.OutputIDsByEd25519Address(context.Background(), &inputAddr, true)
	builder := iotago.NewTransactionBuilder()

	//var identityOne []byte
	//var fromAddr iotago.Address
	keys := make([]iotago.AddressKeys, 0)
	for index, vin := range params.SignParams_data.Ins {
		pri, err := m.GetPrivate(params.MchId, params.Ins[index].FromAddr)
		if err != nil {
			return "", err
		}

		_, addr, err := iotago.ParseBech32(vin.FromAddr)
		if err != nil {
			return "", err
		}

		keys = append(keys, iotago.AddressKeys{Address: addr, Keys: ed25519.PrivateKey(pri)})
		//inputAddr, _ := iotago.ParseEd25519AddressFromHexString(addr.String())
		//vin.FromTxid
		transId, err := hex.DecodeString(vin.FromTxid)
		if err != nil {
			return "", err
		}

		input := &iotago.UTXOInput{
			TransactionOutputIndex: uint16(vin.FromIndex),
		}
		copy(input.TransactionID[:iotago.TransactionIDLength], transId[:])
		builder = builder.AddInput(&iotago.ToBeSignedUTXOInput{Address: addr, Input: input})
	}

	for _, vout := range params.SignParams_data.Outs {
		_, addr, err := iotago.ParseBech32(vout.ToAddr)
		if err != nil {
			return "", err
		}
		toAddr, _ := iotago.ParseEd25519AddressFromHexString(addr.String())
		builder = builder.AddOutput(&iotago.SigLockedSingleOutput{Address: toAddr, Amount: uint64(vout.ToAmountInt64)})
	}
	//addrKeys := iotago.AddressKeys{Address: fromAddr, Keys: ed25519.PrivateKey(identityOne)}
	singer := iotago.NewInMemoryAddressSigner(keys...)
	payload, err := builder.Build(singer)
	if err != nil {
		return "", err
	}

	//completeMsg := &iotago.Message{
	//	//Parents: tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
	//	Payload: payload,
	//	//Nonce:   3495721389537486,
	//}

	serializedPayload, err := payload.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(serializedPayload), nil
}

func (m *SatModel) GetPrivate(mchName, address string) (private []byte, err error) {

	return m.getPrivate(mchName, address)
}

//获取私钥
func (m *SatModel) getPrivate(mchName, address string) (private []byte, err error) {
	//todo:删除调试
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
