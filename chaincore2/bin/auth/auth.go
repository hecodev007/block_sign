package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
	"runtime"
)

func main() {
	maxCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(maxCPU)

	privPath, pubPath := common.GenerateRSAKey(512)
	log.Debug(common.GetRSAKey(privPath))
	log.Debug(common.GetRSAKey(pubPath))

	data_bytes, _ := json.Marshal(map[string]interface{}{
		"a": 1,
		"b": "bbbb",
		"c": 0.5,
	})
	data_sha256 := sha256.Sum256(data_bytes)
	to_sign := []byte(fmt.Sprintf("%d:%d:%s", 0, 0, string(data_sha256[:])))
	log.Debug(string(to_sign), hex.EncodeToString(to_sign))

	// server -> client
	{
		// 签名
		signresult := common.RSASign(privPath, to_sign)
		//auth := fmt.Sprintf("%d:%d:%s", 0, 0, hex.EncodeToString(signresult))
		//log.Debug(auth)

		// 验证签名
		if common.RSAVerifySign(pubPath, signresult, to_sign) {
			log.Debug("验证成功")
		} else {
			log.Debug("验证失败")
		}
	}

	// client -> server
	{
		// 公钥加密
		encresult := common.RSAEncrypt(pubPath, to_sign)

		// 私钥解密
		decresult := common.RSADecrypt(privPath, encresult)
		log.Debug(string(decresult), hex.EncodeToString(decresult))
	}

	beego.Run()
}
