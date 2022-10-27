package controllers

import (
	"bnbsign/secure"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/binance-chain/go-sdk/keys"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
)

type MainController struct {
	beego.Controller
}

//{
//	"coinName": "bnb",
//	"hash": "7a25a174963529a53d9f0e394ab25ef3",
//	"mchId": "1233",
//	"orderId": "123",
//	"data":{
//		"from":"",
//		"to":"",
//		"amount":0,
//		"memo":"",
// 		"denom":"BNB",
//		"AccountNumber":0,
//		"Sequence":0
//	}
//}
func (c *MainController) Post() {
	// 返回数据
	newresp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	// 返回数据
	resp := map[string]interface{}{
		"result":   nil,
		"hash":     "",
		"mchId":    "",
		"orderId":  "",
		"coinName": "",
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		beego.Trace(jsonObj)

		if jsonObj["mchId"] != nil {
			resp["mchId"] = jsonObj["mchId"]
		}
		if jsonObj["orderId"] != nil {
			resp["orderId"] = jsonObj["orderId"]
		}
		if jsonObj["coinName"] != nil {
			resp["coinName"] = jsonObj["coinName"]
		}

		if jsonObj["data"] == nil || jsonObj["hash"] == nil || jsonObj["mchId"] == nil || jsonObj["orderId"] == nil || jsonObj["coinName"] == nil {
			newresp["code"] = 1
			newresp["message"] = "Request param error, data or hash or mchId or orderId or coinName is null"
			break
		}

		mch := jsonObj["mchId"].(string)
		data := jsonObj["data"].(map[string]interface{})
		denom := "BNB"
		if data["denom"] != nil {
			denom = data["denom"].(string)
		}
		coins := ctypes.Coins{ctypes.Coin{Denom: denom, Amount: int64(data["amount"].(float64))}}
		fromaddr, err := ctypes.AccAddressFromBech32(data["from"].(string))
		if err != nil {
			newresp["code"] = 1
			newresp["message"] = "from address format error"
			break
		}
		toaddr, err := ctypes.AccAddressFromBech32(data["to"].(string))
		if err != nil {
			newresp["code"] = 1
			newresp["message"] = "to address format error"
			break
		}

		txmsg := msg.CreateSendMsg(fromaddr, coins, []msg.Transfer{{toaddr, coins}})

		signMsg := tx.StdSignMsg{
			ChainID:       "Binance-Chain-Tigris",
			AccountNumber: int64(data["account_number"].(float64)),
			Sequence:      int64(data["sequence"].(float64)),
			Memo:          data["memo"].(string),
			Msgs:          []msg.Msg{txmsg},
			Source:        0,
		}

		from := data["from"].(string)

		//wif, _ := comm.AesDecrypt(EncryptWifMap[from], []byte(WifKeyListMap[from]))
		key, err := secure.GetPrivateKey(mch, from)
		if err != nil {
			beego.Debug("decrypt key fail !", from)
			newresp["code"] = 1
			newresp["message"] = "decrypt key fail"
			break
		}

		keyManager, err := keys.NewPrivateKeyManager(key)
		if err != nil {
			beego.Error(err)
		}

		signResult, err := keyManager.Sign(signMsg)
		if err != nil {
			beego.Debug(err)
		}
		hexs := hex.EncodeToString(signResult[:])

		_md5data, err := json.Marshal(hexs)
		if err != nil {
			newresp["code"] = 1
			newresp["message"] = "hash error"
			break
		}
		has := md5.Sum(_md5data)
		resp["hash"] = hex.EncodeToString(has[:])
		resp["result"] = hexs

		break
	}

	newresp["data"] = resp
	c.Data["json"] = newresp
	c.ServeJSON()
}
