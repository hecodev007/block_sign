package blockChainsApi

import (
	"crypto/hmac"
	"crypto/sha256"
	. "custody-merchant-admin/config"
	"custody-merchant-admin/util/crypt"
	"custody-merchant-admin/util/xkutils"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

func ValidInsideAddress(clientId, apiSecret, address string) (bool, error) {

	nowTs := time.Now().Unix()
	uuid := xkutils.NewUUId("nonce")
	sg := &crypt.ApiSign{
		ApiKey:    clientId,
		ApiSecret: apiSecret,
		Ts:        fmt.Sprintf("%d", nowTs), //ts为当前时间戳 与服务器时间差正负5秒会被拒绝
		Nonce:     uuid,                     //nonce为随机字符串 不能与上次请求所使用相同
	}
	sign, err := sg.GetSign()
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	v := url.Values{
		"address":   {address},
		"client_id": {clientId},
		"ts": {
			fmt.Sprintf("%d", nowTs),
		},
		"nonce": {uuid},
		"sign":  {sign},
	}
	form, err := xkutils.PostForm(Conf.Blockchain.Url, v)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	var f map[string]interface{}
	json.Unmarshal(form, &f)
	if _, ok := f["data"]; ok {
		return f["data"].(bool), err
	}
	return false, err
}

func ComputeHmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	fmt.Println(h.Sum(nil))
	sha := hex.EncodeToString(h.Sum(nil))
	fmt.Println(sha)
	//	hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(sha))
}
