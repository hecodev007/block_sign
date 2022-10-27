package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

////校验结构中的签名
//func CheckSian() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		var (
//			err      error
//			data     []byte
//			signData map[string]string
//			mchName  string
//			ok       bool
//		)
//
//		if data, err = ioutil.ReadAll(ctx.Request.Body); err == nil {
//			if len(data) != 0 {
//				if signData, err = model.DecodeSignData(data); err == nil {
//					//trim
//					for k, v := range signData {
//						signData[k] = strings.Trim(v, " ")
//					}
//					mchName = signData["sfrom"]
//					if len(mchName) != 0 {
//						if len(signData[service.SignKey]) != 0 {
//							if ok, err = service.VerifySign(mchName, signData); ok {
//								//向下传递
//								ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
//								ctx.Next()
//							}
//						}
//					}
//
//				}
//			}
//		}
//		ctx.String(403, "无权限访问,签名错误")
//		ctx.Abort()
//	}
//}

//校验结构中的签名
func CheckSign() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err        error
			data       []byte
			signParams map[string]string
			mchName    string
			ok         bool
		)
		if data, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
			ctx.String(403, "缺少输入参数")
			ctx.Abort()
			return
		}
		if len(data) == 0 {
			ctx.String(403, "输入参数为空")
			ctx.Abort()
			return
		}
		if signParams, err = getApiSignParam(ctx, data); err != nil {
			//ctx.String(403, "无法识别的输入参数")
			log.Error("CheckSign getApiSignParam error: ", err)
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "无法识别的输入参数", nil)
			ctx.Abort()
			return
		}
		if mchName = signParams["sfrom"]; len(mchName) == 0 {
			//ctx.String(403, "缺少商户名")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "缺少商户名", nil)

			ctx.Abort()
			return
		}
		if len(signParams[service.SignKey]) == 0 {
			//ctx.String(403, "缺少签名参数")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "缺少商户名", nil)

			ctx.Abort()
			return
		}

		if ok, err = service.VerifySign(mchName, signParams); !ok {
			//ctx.String(403, "无权限访问,签名错误")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "签名错误", nil)
			ctx.Abort()
			return
		}
		//向下传递
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		ctx.Next()
	}
}

func getApiSignParam(ctx *gin.Context, data []byte) (map[string]string, error) {
	var (
		err        error
		signParams = map[string]string{}
	)
	switch ctx.HandlerName() {
	case "v1.Transfer":
		params := model.TransferParams{}
		if err = json.Unmarshal(data, &params); err != nil {
			return nil, fmt.Errorf("input data invalidate handler name: v1.Transfer error: %w", err)
		}
		signParams["sign"] = strings.Trim(params.Sign, " ")
		signParams["sfrom"] = strings.Trim(params.Sfrom, " ")
		signParams["outOrderId"] = strings.Trim(params.OutOrderId, " ")
		signParams["coinName"] = strings.Trim(params.CoinName, " ")
		signParams["amount"] = params.Amount.String()
		signParams["toAddress"] = strings.Trim(params.ToAddress, " ")
		signParams["tokenName"] = strings.Trim(params.TokenName, " ")
		signParams["contractAddress"] = strings.Trim(params.ContractAddress, " ")
		signParams["memo"] = strings.Trim(params.Memo, " ")
		signParams["fee"] = params.Fee.String()
		signParams["isForce"] = strconv.FormatBool(params.IsForce)
	default:
		return nil, fmt.Errorf("input data invalidate handler name: %s, data: %s", ctx.HandlerName(), string(data))
	}
	return signParams, nil
}

//校验结构中的签名,改良版本
func CheckSign2() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err        error
			data       []byte
			signParams map[string]interface{}
			ok         bool
		)
		if data, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
			//ctx.String(403, "缺少输入参数")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "缺少输入参数", nil)

			ctx.Abort()
			return
		}
		if len(data) == 0 {
			//ctx.String(403, "输入参数为空")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "输入参数为空", nil)
			ctx.Abort()
			return
		}
		if signParams, err = model.DecodeSignDataInterface(data); err != nil {
			//ctx.String(403, "无法识别的输入参数")
			log.Error("CheckSign getApiSignParam error: ", err)
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "无法识别的输入参数", nil)
			ctx.Abort()
			return
		}
		if _, ok = signParams["sfrom"]; !ok {
			//ctx.String(403, "缺少商户名")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "缺少商户名", nil)

			ctx.Abort()
			return
		}
		if _, ok = signParams[service.SignKey]; !ok {
			//ctx.String(403, "缺少签名参数")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "缺少签名参数", nil)
			ctx.Abort()
			return
		}
		if ok, err = service.VerifySignInterface(signParams); !ok {
			//ctx.String(403, "无权限访问,签名错误")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "无权限访问,签名错误", nil)

			ctx.Abort()
			return
		}
		//向下传递
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		ctx.Next()

	}
}

//新版入口签名
func CheckApiSign() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err error
			ok  bool
		)

		data, _ := ioutil.ReadAll(ctx.Request.Body)
		log.Infof("传入内容体：%s", string(data))
		//获取完之后要重新设置进去，不然会丢失
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

		clientId := ctx.PostForm("client_id")
		ts, _ := strconv.ParseInt(ctx.PostForm("ts"), 10, 64)
		nonce := ctx.PostForm("nonce")
		signStr := ctx.PostForm("sign")
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
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "服务器时间不一致，签名失败", nil)
			ctx.Abort()
			return
		}
		if clientId == "" || ts == 0 || nonce == "" || signStr == "" {
			//ctx.String(403, "参数异常，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "参数异常，签名失败", nil)
			ctx.Abort()
			return
		}
		//ShouldBind 无效
		form := &util.ApiSignParams{
			ClientId: clientId,
			Ts:       ts,
			Nonce:    nonce,
			Sign:     signStr,
		}
		log.Infof("内容：%+v", form)
		if ok, err = service.VerifyApiSign(*form); !ok {
			if err != nil {
				log.Error(err.Error())
			}
			//ctx.String(403, err.Error())
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, err.Error(), nil)
			ctx.Abort()
			return
		}
		//不需要像json向下传递
		ctx.Next()

	}

}

//入口所有参数签名
func CheckApiParamSign() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var (
			err error
			ok  bool
		)

		data, _ := ioutil.ReadAll(ctx.Request.Body)
		log.Infof("传入内容体：%s", string(data))
		//获取完之后要重新设置进去，不然会丢失
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))


		clientId := ctx.PostForm("client_id")
		ts, _ := strconv.ParseInt(ctx.PostForm("ts"), 10, 64)
		nonce := ctx.PostForm("nonce")
		signStr := ctx.PostForm("sign")
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
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "服务器时间不一致，签名失败", nil)
			ctx.Abort()
			return
		}
		if clientId == "" || ts == 0 || nonce == "" || signStr == "" {
			//ctx.String(403, "参数异常，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "参数异常，签名失败", nil)
			ctx.Abort()
			return
		}
		//ShouldBind 无效

		form := make(map[string]interface{})
		fmt.Printf("data = %+v\n",string(data))
		formStr := string(data)
		formArr := strings.Split(formStr,"&")
		if len(formArr)>0{
			for _,item := range formArr{
					kvArr := strings.Split(item,"=")
					if len(kvArr) >=2 {
						v,_ := url.QueryUnescape( kvArr[1])
						form[kvArr[0]] = v
					}
			}
		}else{
			if err = json.Unmarshal(data,&form) ; err != nil {
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "参数Unmarshal异常，签名失败", nil)
				ctx.Abort()
				return
			}
		}

		form["client_id"] = clientId
		form["ts"]=ts
		form["nonce"]=nonce
		form["sign"]=signStr

		log.Infof("内容：%+v", form)
		if ok, err = service.CustodyVerifyMapApiSign(form); !ok {
			if err != nil {
				log.Error(err.Error())
			}
			//ctx.String(403, err.Error())
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, err.Error(), nil)
			ctx.Abort()
			return
		}
		//不需要像json向下传递
		ctx.Next()

	}


}
