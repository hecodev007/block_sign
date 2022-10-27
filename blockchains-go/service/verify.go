package service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/crypto"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
)

const (
	SignKey = "sign"
)

//验证提交信息
func VerifySign(mchName string, data map[string]string) (bool, error) {
	if len(mchName) == 0 {
		return false, errors.New("mch name invalidate")
	}
	if len(data[SignKey]) == 0 {
		return false, errors.New("sign data is empty")
	}
	mvi, err := dao.GetMchVerifyInfo(mchName)
	if err != nil {
		//do logger record error processs
		return false, errors.New("get mch info exception")
	}
	signScript := util.Sign(data, mvi.ApiKey)
	if signScript != data[SignKey] {
		return false, errors.New("sign data no match")
	}
	return true, nil
}

//验证提交信息
func VerifySignInterface(data map[string]interface{}) (bool, error) {
	if _, ok := data["sfrom"]; !ok {
		return false, errors.New("mch name invalidate")
	}
	if _, ok := data[SignKey]; !ok {
		return false, errors.New("sign data is empty")
	}
	mchName := fmt.Sprintf("%v", data["sfrom"])
	mvi, err := dao.GetMchVerifyInfo(mchName)
	if err != nil {
		//do logger record error processs
		return false, errors.New("get mch info exception")
	}
	signScript := util.SignInterface(data, mvi.ApiKey)
	if signScript != data[SignKey] {
		return false, errors.New("sign data no match")
	}
	return true, nil
}

func VerifyApiSign(params util.ApiSignParams) (bool, error) {
	mch, err := dao.FcMchFindByApikey(params.ClientId)
	if err != nil {
		return false, err
	}
	apiSign := &util.ApiSign{
		ApiKey:    mch.ApiKey,
		ApiSecret: mch.ApiSecret,
		Ts:        fmt.Sprintf("%v", params.Ts),
		Nonce:     fmt.Sprintf("%v", params.Nonce),
	}
	sign, err := apiSign.GetSign()
	if err != nil {
		return false, err
	}
	if sign != params.Sign {
		log.Errorf("传入的sign：%s,验证的sign:%s", params.Sign, sign)
		return false, errors.New("签名错误")
	}
	return true, nil
}

func CustodyVerifyApiSign(params util.CustodyApiSignParams) (bool, error) {
	clientId := params["client_id"].(string)
	paramSign := params["sign"].(string)
	mch, err := dao.FcMchFindByApikey(clientId)
	if err != nil {
		return false, err
	}
	params["api_secret"] = mch.ApiSecret

	sign, err := params.GetSign()
	if err != nil {
		return false, err
	}
	if sign != paramSign{
		log.Errorf("传入的sign：%s,验证的sign:%s", paramSign, sign)
		return false, errors.New("签名错误")
	}
	return true, nil
}


func CustodyVerifyMapApiSign(params map[string]interface{}) (bool, error) {
	clientId := params["client_id"].(string)
	paramSign := params["sign"].(string)
	mch, err := dao.FcMchFindByApikey(clientId)
	if err != nil {
		return false, err
	}
	params["api_secret"] = mch.ApiSecret

	sign, err :=GetMapParamSign(params)
	if err != nil {
		return false, err
	}
	if sign != paramSign{
		log.Errorf("传入的sign：%s,验证的sign:%s", paramSign, sign)
		return false, errors.New("签名错误")
	}
	return true, nil
}

//获取签名sign
func GetMapParamSign(param map[string]interface{}) (sign string, err error) {

	apiKey := param["client_id"].(string)
	apiSecret := param["api_secret"].(string)
	nonce := interface2String(param["nonce"])
	ts := fmt.Sprintf("%v", param["api_key"])


	if apiKey == "" || apiSecret == "" || nonce == "" || ts == "" {
		return "", errors.New("params error")
	}
	nonceStr := util.SignNonceMap[apiKey]
	if nonce == nonceStr {
		return "", errors.New("same nonce as last time")
	} else {
		util.SignNonceMap[apiKey] = nonce
	}


	str := util.EncodeQueryInterface(param)

	log.Infof("enstr1 := %v\n", str)
	log.Infof("enstr2 := %v\n", apiSecret)
	sign = util.ComputeHmac256(str, apiSecret)
	log.Infof("enstr3 sign:= %v\n", sign)
	return
}



func VerifyApiSignV2(clientId string, hash string, sign string, params map[string]string) (bool, error) {
	mch, err := dao.FcMchFindByApikey(clientId)
	if err != nil {
		return false, err
	}

	log.Infof("mch.ApiPublicKey=%s", mch.ApiPublicKey)
	if mch.ApiPublicKey == "" {
		return false, errors.New("public key empty")
	}

	publicKey, err := hex.DecodeString(mch.ApiPublicKey)
	if err != nil {
		return false, err
	}

	origin := util.EncodeQueryString(params)
	log.Infof("origin=%s", origin)

	hashBytes, _ := hex.DecodeString(hash)
	computedHash := crypto.Sha256([]byte(origin))

	if !bytes.Equal(hashBytes, computedHash) {
		return false, errors.New("invalid hash")
	}

	signBytes, _ := hex.DecodeString(sign)

	return crypto.Verify(publicKey, computedHash, signBytes), nil
}

func interface2String(inter interface{}) string{
	switch inter.(type) {
	case string:
		s := inter.(string)
		return s
	case int:
		i :=  inter.(int)
		s := fmt.Sprintf("%v",i)
		return s
	case int64:
		i :=  inter.(int64)
		s := fmt.Sprintf("%v",i)
		return s
	case float64:
		i :=  inter.(float64)
		s := fmt.Sprintf("%v",i)
		return s
	}
	return ""
}