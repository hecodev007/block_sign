package custody

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"time"
)

//VerifyParamFromCustody
func VerifyParamFromCustody(c *gin.Context) {

	verifyData := c.PostForm("verify_data")
	log.Infof("verifyData:%+v", verifyData)
	verifyMap := util.CustodyApiSignParams{}
	err := json.Unmarshal([]byte(verifyData),&verifyMap)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("校验商户数据错误： %v", err.Error()).Error(), nil)
		return
	}
	log.Infof("verifyMap:%+v", verifyMap)
	err =  VerifyParam(verifyMap)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL,  fmt.Errorf("校验商户数据不通过： %v", err.Error()).Error(), nil)
		return
	}
	httpresp.HttpRespOK(c, "success", nil)
	return
}


func VerifyParam(verifyMap util.CustodyApiSignParams)(err error){
	var ok bool
	clientId := verifyMap["client_id"].(string)
	tsStr :=verifyMap["ts"].(float64)
	ts := int64(tsStr)
	nonce := interface2String(verifyMap["nonce"])
	signStr :=  interface2String(verifyMap["sign"])
	nowTs := time.Now().Unix()
	log.Infof("client_id:%s", clientId)
	log.Infof("ts:%d", ts)
	log.Infof("now:%d", nowTs)
	log.Infof("nonce:%s", nonce)
	log.Infof("signStr:%s", signStr)
	//单位秒
	if (nowTs-ts) > 60 || (nowTs-ts) < -60 {
		//超过服务器时间60秒自动失败
		//ctx.String(403, "服务器时间不一致，签名失败")
		err = fmt.Errorf("服务器时间不一致，签名失败")
		return
	}
	if clientId == "" || ts == 0 || nonce == "" || signStr == "" {
		//ctx.String(403, "参数异常，签名失败")
		err = fmt.Errorf("参数缺失，签名失败")
		return
	}
	//ShouldBind 无效
	//form := util.CustodyApiSignParams{}
	//if err = json.Unmarshal(data,&form) ; err != nil {
	//	log.Infof("err内容：%+v", err)
	//	err = fmt.Errorf("参数Unmarshal异常，签名失败")
	//	return
	//}
	//verifyMap["client_id"] = clientId
	//verifyMap["ts"]=ts
	//verifyMap["nonce"]=nonce
	//verifyMap["sign"]=signStr

	log.Infof("内容：%+v", verifyMap)
	if ok, err = service.CustodyVerifyApiSign(verifyMap); !ok {
		if err != nil {
			log.Error(err.Error())
		}
		return
	}
	return

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