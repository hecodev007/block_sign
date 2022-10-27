package controllers

import (
	"bnbsign/conf"
	"bnbsign/crypto"
	"bnbsign/secure"
	"bnbsign/service"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
)

type GenAddressController struct {
	beego.Controller
}

// request json
//{
// 	"result":[],
//}
func (c *GenAddressController) Post() {
	// 返回数据
	newresp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)

		if jsonObj["count"] == nil || jsonObj["mch"] == nil {
			newresp["code"] = 1
			newresp["message"] = "Request param error"
			beego.Debug("Request param error")
			break
		}

		resp := map[string]interface{}{
			"address": nil,
		}

		count := int(jsonObj["count"].(float64))
		mch := jsonObj["mch"].(string)
		ret, list := service.GenAddress(count)
		if ret == false {
			newresp["code"] = 1
			newresp["message"] = "Generate address fail"
			break
		}

		kmsReq := &secure.RequestPushKey{CoinCode: "bnb", Mch: mch}
		addrList := make([]string, 0)
		for k, v := range list {
			cryptoText, err := crypto.AesBase64Str(v, conf.Global.Secret.TransportSecureKey, true)
			if err != nil {
				newresp["code"] = 1
				newresp["message"] = fmt.Sprintf("[%s]failed to aes crypto: %v", k, err)
				return
			}
			addrList = append(addrList, k)
			kmsReq.Data = append(kmsReq.Data, secure.PushKeyData{Address: k, PrivateKey: cryptoText})
		}

		if err := secure.PushKeyToKMS(kmsReq); err != nil {
			newresp["code"] = 1
			newresp["message"] = err.Error()
			return
		}

		resp["address"] = addrList
		newresp["data"] = resp

		break
	}

	c.Data["json"] = newresp
	c.ServeJSON()

}
