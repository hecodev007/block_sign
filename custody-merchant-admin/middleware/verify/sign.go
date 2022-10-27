package verify

import (
	"crypto/hmac"
	"crypto/sha256"
	"custody-merchant-admin/module/errcode"
	"fmt"
	"net/url"
	"sort"
	"strings"

	. "custody-merchant-admin/config"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"strconv"
	"time"
)

//由于nonce为随机字符串 不能与上次请求所使用相同，使用map临时存储
//key为client_id value为nonce
var SignNonceMap map[string]string

const (
	//ClientId       = "d28fa2b0-d36a-4b5f-a7ff-0612bdc620d7"
	//ApiSecret      = "31ywhtAGwh74ThyfnGHj788aVWhbViKhpZ"
	CallBack       = "/custody/blockchain/callback"       //提现回调地址
	InComeCallBack = "/custody/blockchain/incomecallback" //充值回调地址
)

func init() {
	SignNonceMap = make(map[string]string, 0)
}

type CustodyApiSignParams map[string]interface{}

//入口所有参数签名
func CheckApiParamSign() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var (
				err error
				ok  bool
			)

			data, _ := ioutil.ReadAll(ctx.Request().Body)
			log.Infof("传入内容体：%s", string(data))
			////获取完之后要重新设置进去，不然会丢失
			//ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

			clientId := ctx.Param("client_id")
			ts, _ := strconv.ParseInt(ctx.Param("ts"), 10, 64)
			nonce := ctx.Param("nonce")
			signStr := ctx.Param("sign")
			nowTs := time.Now().Unix()
			log.Infof("client_id:%s", clientId)
			log.Infof("ts:%d", ts)
			log.Infof("now:%d", nowTs)
			log.Infof("nonce:%s", nonce)
			log.Infof("signStr:%s", signStr)
			//单位秒
			if (nowTs-ts) > 60 || (nowTs-ts) < -30 {
				//超过服务器时间60秒自动失败
				err = fmt.Errorf("服务器时间不一致，签名失败")
				ctx.JSON(errcode.ParamSignError.StatusCode(), map[string]interface{}{
					"code": errcode.ParamSignError.StatusCode(),
					"msg":  err.Error(),
				})
				return err
			}
			if clientId == "" || ts == 0 || nonce == "" || signStr == "" {
				err = fmt.Errorf("参数异常，签名失败")
				ctx.JSON(errcode.ParamSignError.StatusCode(), map[string]interface{}{
					"code": errcode.ParamSignError.StatusCode(),
					"msg":  err.Error(),
				})
				return err
			}

			form := CustodyApiSignParams{}

			if err = json.Unmarshal(data, &form); err != nil {
				err = fmt.Errorf("参数Unmarshal异常，签名失败")
				ctx.JSON(errcode.ParamSignError.StatusCode(), map[string]interface{}{
					"code": errcode.ParamSignError.StatusCode(),
					"msg":  err.Error(),
				})
				return err
			}
			form["client_id"] = clientId
			form["ts"] = ts
			form["nonce"] = nonce
			form["sign"] = signStr

			log.Infof("内容：%+v", form)
			if ok, err = CustodyVerifyApiSign(form); !ok {
				if err != nil {
					log.Error(err.Error())
				}
				err = fmt.Errorf("参数Unmarshal异常，签名失败")
				ctx.JSON(errcode.ParamSignError.StatusCode(), map[string]interface{}{
					"code": errcode.ParamSignError.StatusCode(),
					"msg":  err.Error(),
				})
				return err
			}
			return next(ctx)

		}
	}
}

func CustodyVerifyApiSign(params CustodyApiSignParams) (bool, error) {

	clientId := params["client_id"].(string)
	if clientId != Conf.BlockchainCustody.ClientId {
		return false, errors.New("ClientId错误")
	}
	paramSign := params["sign"].(string)

	params["api_secret"] = Conf.BlockchainCustody.ApiSecret

	sign, err := params.GetSign()
	if err != nil {
		return false, err
	}
	if sign != paramSign {
		log.Errorf("传入的sign：%s,验证的sign:%s", paramSign, sign)
		return false, errors.New("签名错误")
	}
	return true, nil
}

//获取签名sign
func (s *CustodyApiSignParams) GetSign() (sign string, err error) {
	param := *s
	apiKey := param["api_key"].(string)
	apiSecret := param["api_key"].(string)
	nonce := param["api_key"].(string)
	ts := fmt.Sprintf("%v", param["api_key"])

	if apiKey == "" || apiSecret == "" || nonce == "" || ts == "" {
		return "", errors.New("params error")
	}
	nonceStr := SignNonceMap[apiKey]
	if nonce == nonceStr {
		return "", errors.New("same nonce as last time")
	} else {
		SignNonceMap[apiKey] = nonce
	}

	str := EncodeQueryInterface(*s)
	sign = ComputeHmac256(str, apiSecret)
	return
}

// 拼接除sign以外的所有query字符串
func EncodeQueryInterface(query map[string]interface{}) string {
	keys := make([]string, 0)
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines = make([]string, 0)
	for _, item := range keys {
		if item == "sign" || item == "api_secret" {
			continue
		}
		s := interface2String(query[item])
		log.Infof("interface2String %v,%v\n", query[item], s)
		lines = append(lines, url.QueryEscape(item)+"="+s)
		//lines = append(lines, url.QueryEscape(item)+"="+url.QueryEscape(fmt.Sprintf("%v", query[item])))
	}
	return strings.Join(lines, "&")
}

//签名
func ComputeHmac256(data string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func interface2String(inter interface{}) string {
	switch inter.(type) {
	case string:
		s := inter.(string)
		return s
	case int:
		i := inter.(int)
		s := fmt.Sprintf("%v", i)
		return s
	case int64:
		i := inter.(int64)
		s := fmt.Sprintf("%v", i)
		return s
	case float64:
		i := inter.(float64)
		//s := strconv.FormatFloat(i, 'E', -1, 64)
		s := fmt.Sprintf("%.f", i)
		return s
	}
	s := fmt.Sprintf("%v", inter)
	return s

}
