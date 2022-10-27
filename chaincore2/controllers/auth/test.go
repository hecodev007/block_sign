package auth

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"
)

type TestController struct {
	beego.Controller
}

func (c *TestController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"status":     500,
		"total":      0,
		"tx_hex":     "123",
	}

	// 解密
	{
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)
		log.Debug(c.Ctx.Request.Header)

		//authorization := c.Ctx.Request.Header.Get("auth")
		//authArray := strings.Split(authorization, ":")
		//
		////data_bytes, _ := c.Ctx.Input.RequestBody
		//data_sha256 := sha256.Sum256(c.Ctx.Input.RequestBody)
		//to_sign := []byte(fmt.Sprintf("%s:%s:%s", authArray[0], authArray[1], string(data_sha256[:])))
		//log.Debug(hex.EncodeToString(to_sign))
		//
		//// 私钥解密
		//decresult := common.RSADecrypt("my_rsa_private.pem", []byte(authArray[2]))
		//log.Debug(string(decresult), hex.EncodeToString(decresult))
	}


	{
		// 签名
		//data_bytes, _ := json.Marshal(resp)
		//data_sha256 := sha256.Sum256(data_bytes)
		//to_sign := []byte(fmt.Sprintf("%d:%d:%s", 0, time.Now().Unix() + 60, string(data_sha256[:])))
		//log.Debug(hex.EncodeToString(to_sign))
		//
		//signresult := common.RSASign("private.pem", to_sign)
		//auth := fmt.Sprintf("%d:%d:%s", 0, 0, hex.EncodeToString(signresult))
		//log.Debug(auth)
		//
		//c.Ctx.ResponseWriter.Header().Set("auth", auth)
		//c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
