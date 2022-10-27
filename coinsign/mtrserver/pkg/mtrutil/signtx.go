package mtrutil

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
	"github.com/zwjlink/meterutil/meter"
	"github.com/zwjlink/meterutil/tx"
	"log"
	"math/big"
	"strings"
	"time"
)

type MtrTpl struct {
	//ChainTag（创世区块 ID 的最后2位字符，可以动态获取，不过一般也不会变，除非清链）
	ChainTag int `json:"chainTag"`
	//过期区块数量
	Expiration uint32 `json:"expiration"`
	//21000，类似gaslimit
	Gas uint64 `json:"gas"`
	//gasprice
	GasPriceCoef uint8 `json:"gasPriceCoef"`
	//BlockRef 必须指向一个已经存在的区块。
	//可以指定一个之前的区块,但是需要注意区块如果交易没在引用区块+（指定数量）之内打包，将会失败
	//因此个人建议只使用未来区块，+18
	BlockRef        uint32  `json:"blockRef"`
	Outs            []ToTpl `json:"outs"`
	FromAddr        string  `json:"fromAddr"`
	FromPrivKey     string  `json:"fromPrivKey"`
	FeeAddr         string  `json:"feeAddr,omitempty"`
	FeePayerPrivKey string  `json:"feePayerPrivKey,omitempty"` //手续费代付者，如果为空则使用from的gas 预留现在官方不支持
}

type ToTpl struct {
	ToAddress   string       `json:"toAddress"`
	ToAmountInt *big.Int     `json:"toAmountInt"`
	Token       tx.TokenType `json:"token"`
}

func (mtrTpl *MtrTpl) Check() error {
	if mtrTpl.ChainTag <= 0 {
		return errors.New("error ChainTag")
	}
	if mtrTpl.BlockRef <= 0 {
		return errors.New("error BlockRef")
	}
	if len(mtrTpl.Outs) == 0 {
		return errors.New("error outs")
	}
	mtrTpl.FromAddr = strings.ToLower(mtrTpl.FromAddr)
	mtrTpl.FeeAddr = strings.ToLower(mtrTpl.FeeAddr)

	from, err := meter.ParseAddress(mtrTpl.FromAddr)
	if err != nil {
		err = fmt.Errorf("from:address error:%s,%s", mtrTpl.FromAddr, err.Error())
		return err
	}
	if strings.ToLower(from.String()) != mtrTpl.FromAddr {
		err = fmt.Errorf("from:inconsistent address,from:%s,parse:%s", mtrTpl.FromAddr, from.String())
		return err
	}

	for _, v := range mtrTpl.Outs {
		toAm := decimal.NewFromBigInt(v.ToAmountInt, -18)
		if toAm.LessThanOrEqual(decimal.Zero) {
			err = fmt.Errorf("to address:%s,amount error:%s", v.ToAddress, toAm.String())
			return err
		}
		to, err := meter.ParseAddress(v.ToAddress)
		if err != nil {
			err = fmt.Errorf("to: address error:%s,%s", v.ToAddress, err.Error())
			return err
		}
		if strings.ToLower(to.String()) != strings.ToLower(v.ToAddress) {
			err = fmt.Errorf("to:inconsistent address,to:%s,parse:%s", v.ToAddress, to.String())
			return err
		}
	}

	if len(mtrTpl.FromPrivKey) == 66 && strings.HasPrefix(mtrTpl.FromPrivKey, "0x") {
		mtrTpl.FromPrivKey = mtrTpl.FromPrivKey[2:]
	} else if len(mtrTpl.FromPrivKey) == 64 {
		mtrTpl.FromPrivKey = mtrTpl.FromPrivKey
	} else {
		log.Println(len(mtrTpl.FromPrivKey))
		return errors.New("error privkey len")
	}

	//私钥转换地址
	fromByte, _ := hexutil.Decode("0x" + mtrTpl.FromPrivKey)
	fromP, err := GetAccountFromBytes(fromByte)
	if err != nil {
		return fmt.Errorf("from :address：%s,privkey error", mtrTpl.FromAddr)
	}
	if strings.ToLower(fromP.Address.String()) != mtrTpl.FromAddr {
		err = fmt.Errorf("from :address,from:%s,privkey parse addr:%s", mtrTpl.FromAddr, fromP.Address.String())
		return err
	}

	if len(mtrTpl.FeePayerPrivKey) == 66 && strings.HasPrefix(mtrTpl.FeePayerPrivKey, "0x") {
		mtrTpl.FeePayerPrivKey = mtrTpl.FeePayerPrivKey[2:]
	}
	if mtrTpl.FeeAddr != "" && mtrTpl.FeePayerPrivKey != "" {
		feeAddr, err := meter.ParseAddress(mtrTpl.FeeAddr)
		if err != nil {
			err = fmt.Errorf("feepayer: address error:%s,%s", mtrTpl.FeeAddr, err.Error())
			return err
		}
		if strings.ToLower(feeAddr.String()) != mtrTpl.FeeAddr {
			err = fmt.Errorf("feepayer:inconsistent address,feepayer:%s,parse:%s", mtrTpl.FeeAddr, feeAddr.String())
			return err
		}

		feeByte, _ := hexutil.Decode("0x" + mtrTpl.FeePayerPrivKey)
		feeP, err := GetAccountFromBytes(feeByte)
		if err != nil {
			return fmt.Errorf("fee :address：%s,privkey error", mtrTpl.FeeAddr)
		}
		if feeP.Address.String() != mtrTpl.FeeAddr {
			err = fmt.Errorf("feepayer :address,feepayer:%s,privkey parse addr:%s", mtrTpl.FeeAddr, feeP.Address.String())
			return err
		}

	}

	if mtrTpl.Gas < 21000 {
		mtrTpl.Gas = 21000
	}
	//128
	if mtrTpl.GasPriceCoef < 0 {
		mtrTpl.GasPriceCoef = 0
	}
	return nil
}

func (mtrTpl *MtrTpl) SignTxTpl() (string, error) {
	var delegator []byte //代付者

	err := mtrTpl.Check()
	if err != nil {
		return "", err
	}
	chainTag := byte(mtrTpl.ChainTag) // chainTag is NOT the same across chains
	var expiration = mtrTpl.Expiration
	var gas = mtrTpl.Gas

	txBuild := new(tx.Builder).
		BlockRef(tx.NewBlockRef(mtrTpl.BlockRef)).
		ChainTag(chainTag).
		Expiration(expiration).
		GasPriceCoef(mtrTpl.GasPriceCoef).
		Gas(gas).
		Nonce(uint64(time.Now().UnixNano()))
	for _, v := range mtrTpl.Outs {
		toAddr, err := meter.ParseAddress(v.ToAddress)
		if err != nil {
			log.Println("111")
			return "", err
		}
		txBuild.Clause(tx.NewClause(&toAddr).
			WithValue(v.ToAmountInt). // value in Wei
			WithToken(byte(v.Token)))
	}

	if mtrTpl.FeePayerPrivKey != "" {
		//手续费代付
		var feat tx.Features
		feat.SetDelegated(true)
		txBuild.Features(feat)
	}
	tx := txBuild.Build()

	privKey, err := crypto.HexToECDSA(mtrTpl.FromPrivKey)
	if err != nil {
		return "", err
	}
	sig, err := crypto.Sign(tx.SigningHash().Bytes(), privKey)
	if err != nil {
		return "", err
	}

	//判断是否是代付手续费
	if tx.Features().IsDelegated() {
		o := crypto.PubkeyToAddress(privKey.PublicKey)
		hash := tx.DelegatorSigningHash(meter.Address(o))
		log.Println(mtrTpl.FeePayerPrivKey)
		delegator, err = hex.DecodeString(mtrTpl.FeePayerPrivKey)
		if err != nil {
			return "", err
		}
		priv2, err := crypto.ToECDSA(delegator)
		if err != nil {
			return "", err
		}
		delegatorSig, err := crypto.Sign(hash.Bytes(), priv2)
		if err != nil {
			return "", err
		}
		sig = append(sig, delegatorSig...)
	}

	tx = tx.WithSignature(sig)
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}
	log.Println("Built Tx: ", tx.String())
	return hexutil.Encode(rlpTx), nil
}
