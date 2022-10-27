package blockChainsApi

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/verify"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//BlockChainGetCoinList 主链币/代币 列表
func BlockChainGetCoinList(limit, offset int, clientId, apiSecret string) (list []domain.BCCoinInfo, err error) {

	param := make(map[string]interface{})
	param["limit"] = strconv.Itoa(limit)
	param["offset"] = strconv.Itoa(offset)
	v := CreateSendUrlValues(param, clientId, apiSecret)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.CoinList, v)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(" blockchain form = %+v", string(form))
	log.Infof(" blockchain form = %+v", string(form))
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		coins := map[string][]domain.BCCoinInfo{}
		err = json.Unmarshal(resByte, &coins)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return coins["list"], err
	}
	return
}

//BlockChainCreateClientIdSecret 商户 创建client_id,安全secret
func BlockChainCreateClientIdSecret(req domain.BCMchReq, clientId, apiSecret string) (item domain.BCMchInfo, err error) {

	param := make(map[string]interface{})
	param["name"] = req.Phone
	param["phone"] = req.Phone
	param["email"] = req.Email
	param["company_img"] = req.CompanyImg

	v := CreateSendUrlValues(param, clientId, apiSecret)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.CreateMch, v)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" blockchain form = %+v", string(form))
	log.Infof(" blockchain form = %+v", string(form))

	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		err = json.Unmarshal(resByte, &item)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return item, err
	}
	return
}

//BlockChainReSecretClientIdSecret 商户重置密钥 client_id,安全secret
func BlockChainReSecretClientIdSecret(req domain.BCMchReq, clientId, apiSecret string) (item domain.BCMchInfo, err error) {

	param := make(map[string]interface{})
	param["api_key"] = req.ClientId
	v := CreateSendUrlValues(param, clientId, apiSecret)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.ResetMch, v)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" blockchain form = %+v", string(form))
	log.Infof(" blockchain form = %+v", string(form))

	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		err = json.Unmarshal(resByte, &item)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return item, err
	}
	return
}

//BlockChainSearchClientIdSecret 商户查询 client_id,安全secret
func BlockChainSearchClientIdSecret(req domain.BCMchReq, clientId, apiSecret string) (item domain.BCMchInfo, err error) {
	param := make(map[string]interface{})
	param["api_key"] = req.ClientId

	v := CreateSendUrlValues(param, clientId, apiSecret)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.ResetMch, v)
	if err != nil {
		fmt.Println(err)
		return
	}
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		err = json.Unmarshal(resByte, &item)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return item, err
	}
	return
}

//BlockChainVerifyParamFromCustody 验证托管后台接受来自商户的参数
func BlockChainVerifyParamFromCustody(params interface{}, clientId, apiSecret string) (result bool, err error) {
	param := make(map[string]interface{})
	paramsByte, err := json.Marshal(params)
	if err != nil {
		return false, err
	}
	param["verify_data"] = string(paramsByte)

	v := CreateSendUrlValues(param, clientId, apiSecret)

	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.VerifyParam, v)
	if err != nil {
		fmt.Println(err)
		return
	}
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Code == 0 {
		result = true
	}
	return
}

//BlockChainBatchCreateAddress 商户钱包地址创建（批量创建地址）
func BlockChainBatchCreateAddress(num int, userClientId, coinName, clientId, apiSecret string) (list []string, err error) {
	list = make([]string, 0)

	param := make(map[string]interface{})
	param["api_key"] = userClientId
	param["outOrderId"] = xkutils.NewUUId("outOrderId")
	param["coinName"] = coinName
	param["num"] = num

	v := CreateSendUrlValues(param, clientId, apiSecret)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.CreateAddress, v)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" blockchain form = %+v", string(form))
	log.Infof(" blockchain form = %+v", string(form))

	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		data := make(map[string]interface{})
		err = json.Unmarshal(resByte, &data)
		if err != nil {
			err = fmt.Errorf("json Unmarshal1 err:%v", err.Error())
			return
		}
		if _, ok := data["list"]; ok {
			l := data["list"]
			lByte, _ := json.Marshal(l)
			err = json.Unmarshal(lByte, &list)
			if err != nil {
				err = fmt.Errorf("json Unmarshal2 err:%v", err.Error())
				return
			}
		}
	}
	return
}

//BlockChainCreateLotCoinAddress 商户钱包地址创建（批量创建地址）
/*
coinName 逗号拼接的多币种字符串
*/
func BlockChainCreateLotCoinAddress(coinArr []string, userClientId, clientId, apiSecret string) (list map[string]string, err error) {
	coinName := strings.Join(coinArr, ",")
	param := make(map[string]interface{})
	param["api_key"] = userClientId
	param["outOrderId"] = xkutils.NewUUId("outOrderId")
	param["coinName"] = coinName
	param["num"] = 1

	v := CreateSendUrlValues(param, clientId, apiSecret)

	log.Errorf(" BlockChainCreateLotCoinAddress param = %+v", v)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.CreateLotCoinAddress, v)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" blockchain form = %+v", string(form))
	log.Errorf(" blockchain form = %+v", string(form))

	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		data := make(map[string]interface{})
		err = json.Unmarshal(resByte, &data)
		if err != nil {
			err = fmt.Errorf("json Unmarshal1 err:%v", err.Error())
			return
		}
		if _, ok := data["list"]; ok {
			l := data["list"]
			lByte, _ := json.Marshal(l)
			err = json.Unmarshal(lByte, &list)
			if err != nil {
				err = fmt.Errorf("json Unmarshal2 err:%v", err.Error())
				return
			}
		}
		return list, err
	}
	return
}

//BlockChainBindAddress 商户钱包地址绑定，用户充值回调
func BlockChainBindAddress(coinArr []string, userClientId, clientId, apiSecret string) (err error) {

	coinName := strings.Join(coinArr, ",")
	param := make(map[string]interface{})
	param["api_key"] = userClientId
	param["address"] = Conf.BlockchainCustody.CallBackBaseUrl + verify.InComeCallBack
	param["coin_name"] = coinName
	param["ip"] = Conf.BlockchainCustody.WhiteIp

	v := CreateSendUrlValues(param, clientId, apiSecret)

	log.Errorf(" blockchain form = %+v", v)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.BindAddress, v)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf(" blockchain form = %+v", string(form))
	log.Infof(" blockchain form = %+v", string(form))

	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	return
}

//BlockChainWithdrawCoin 提现接口
/*
clientId ，apiSecret 托管后台 id,密钥
*/
func BlockChainWithdrawCoin(req domain.BCWithDrawReq, clientId, apiSecret string) (err error) {
	param := make(map[string]interface{})
	//param["client_id"] = req.ApiKey
	param["api_key"] = req.ApiKey
	param["callBack"] = Conf.BlockchainCustody.CallBackBaseUrl + verify.CallBack
	param["outOrderId"] = req.OutOrderId
	param["coinName"] = req.CoinName
	param["amount"] = req.Amount
	param["toAddress"] = req.ToAddress
	param["tokenName"] = req.TokenName
	param["contractAddress"] = req.ContractAddress //合约地址
	param["memo"] = req.Memo

	//param["fee"] = req.Fee
	//param["isForce"] = req.IsForce

	v := CreateSendUrlValues(param, clientId, apiSecret)

	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.Withdraw, v)
	if err != nil {
		fmt.Println("BlockChainWithdrawCoin err1 := ", err)
		return
	}
	log.Errorf("BlockChainWithdrawCoin  string(form) := %+v", string(form))
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println("BlockChainWithdrawCoin err2 := ", err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
	}
	return
}

//BlockChainCoinBalance 余额查询 //TODO:未完成
func BlockChainCoinBalance(clientId, apiSecret string) (item []domain.BCMchInfo, err error) {

	param := make(map[string]interface{})
	v := CreateSendUrlValues(param, clientId, apiSecret)

	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.CreateMch, v)
	if err != nil {
		fmt.Println(err)
		return
	}
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		err = json.Unmarshal(resByte, &item)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return
	}
	return
}

//BlockChainOrderUpChainStatus 上链结果回调/查询
func BlockChainOrderUpChainStatus(userClientId, outOrderId, clientId, apiSecret string) (status int, err error) {
	param := make(map[string]interface{})
	param["api_key"] = userClientId
	param["outOrderId"] = outOrderId
	v := CreateSendUrlValues(param, clientId, apiSecret)

	fmt.Printf("data = %+v\n", v)
	form, err := xkutils.PostForm(Conf.BlockchainCustody.BaseUrl+Conf.BlockchainCustody.ChainStatus, v)
	if err != nil {
		fmt.Println(err)
		return
	}
	var f domain.BCBaseInfo
	err = json.Unmarshal(form, &f)
	if err != nil {
		fmt.Println(err)
		return
	}
	if f.Code != 0 {
		err = fmt.Errorf(f.Msg)
		return
	}
	if f.Data != nil {
		resByte, resByteErr := json.Marshal(f.Data)
		if resByteErr != nil {
			err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
			return
		}
		err = json.Unmarshal(resByte, &status)
		if err != nil {
			err = fmt.Errorf("json Unmarshal err:%v", err.Error())
			return
		}
		return
	}
	return
}

//CreateSendUrlValues 生成签名参数
func CreateSendUrlValues(req map[string]interface{}, clientId, apiSecret string) (result url.Values) {
	nowTs := time.Now().Unix()
	uuid := xkutils.NewUUId("nonce")

	param := make(map[string]interface{})
	param["nonce"] = uuid
	param["client_id"] = clientId
	//param["secret"] = apiSecret
	param["ts"] = fmt.Sprintf("%d", nowTs)

	for k, v := range req {
		param[k] = v
	}
	log.Infof("enstr param := %+v\n", param)
	str := verify.EncodeQueryInterface(param)
	log.Infof("enstr1 := %v\n", str)
	log.Infof("enstr2 := %v\n", apiSecret)
	sign := verify.ComputeHmac256(str, apiSecret)
	log.Infof("enstr3 := %v\n", sign)
	result = url.Values{}
	for k, v := range param {
		if k != "secret" {
			result.Set(k, fmt.Sprintf("%v", v))
		}
	}
	result.Set("sign", sign)
	return
}
