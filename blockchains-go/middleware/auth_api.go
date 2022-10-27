package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"io/ioutil"
	"strings"
)

type MiddleAuthParams struct {
	Sign     string `json:"sign"`
	Sfrom    string `json:"sfrom"`
	CoinName string `json:"coinName"`
}

func AuthApi() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		//截取api路径
		req := ctx.Request
		//req.RequestURI 会带末尾参数
		//req.URL.Path 不会带末尾参数
		path := req.URL.Path
		ip := ctx.ClientIP()
		bodyByte, _ := ioutil.ReadAll(req.Body)
		params := new(MiddleAuthParams)
		json.Unmarshal(bodyByte, params)
		if strings.TrimSpace(params.Sfrom) != "" {
			mch, err := mchService.GetAppId(params.Sfrom)
			if err != nil {
				log.Errorf("获取mchId异常：%s", err.Error())
				ctx.String(403, "无权限访问")
				ctx.Abort()
				return
			}

			pathArr := strings.Split(path, "/")
			suffix := "/" + pathArr[len(pathArr)-1]

			if "/applyTransactionSecure" == suffix {
				suffix = "/applyTransaction"
			}

			//验证是否允许访问api 包含了ip验证
			ok, err := transferSecurityService.VerifyApiPermission(suffix, params.CoinName, ip, mch.Id)
			if err != nil {
				log.Errorf("验证路径权限异常：%s", err.Error())
				//ctx.String(403, "无权限访问")
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+","+err.Error(), nil)
				ctx.Abort()
				return
			}
			if ok {
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
				ctx.Next()
				return
			}
		}
		//ctx.String(403, "无权限访问")
		httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+",缺少商户名参数", nil)

		ctx.Abort()
		return
	}
}

func AuthApiV3() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//截取api路径
		req := ctx.Request
		//req.RequestURI 会带末尾参数
		//req.URL.Path 不会带末尾参数
		path := req.URL.Path
		ip := ctx.ClientIP()
		bodyByte, _ := ioutil.ReadAll(req.Body)
		clientId := ctx.PostForm("client_id")
		coinName := ctx.PostForm("coinName")
		coinName = strings.ToLower(coinName)
		if strings.TrimSpace(clientId) != "" {
			mch, err := mchService.GetAppIdByApiKey(clientId)
			if err != nil {
				log.Errorf("获取mchId异常：%s", err.Error())
				//ctx.String(403, "无权限访问")
				//ctx.Abort()
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+",商户异常", nil)
				ctx.Abort()
				return
			}
			pathArr := strings.Split(path, "/")
			suffix := "/" + pathArr[len(pathArr)-1]

			if "/applyTransactionSecure" == suffix {
				suffix = "/applyTransaction"
			}

			//验证是否允许访问api 包含了ip验证
			ok, err := transferSecurityService.VerifyApiPermission(suffix, coinName, ip, mch.Id)
			if err != nil {
				log.Errorf("验证路径权限异常：%s", err.Error())
				//ctx.String(403, "无权限访问")
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+","+err.Error(), nil)

				ctx.Abort()
				return
			}
			if ok {
				log.Infof("AuthApiV3 验证结束")
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
				ctx.Next()
				return
			}
		}
		//ctx.String(403, "无权限访问")
		httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+",缺少商户名参数", nil)
		ctx.Abort()
		return
	}
}

func AuthApiV4() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, _ := getRawData(ctx)
		request := &ApplyTrxRequest{}
		json.Unmarshal(data, request)
		req := ctx.Request
		path := req.URL.Path
		ip := ctx.ClientIP()
		clientId := request.ClientId
		coinName := strings.ToLower(request.CoinName)
		if strings.TrimSpace(clientId) != "" {
			mch, err := mchService.GetAppIdByApiKey(clientId)
			if err != nil {
				log.Errorf("获取mchId异常：%s", err.Error())
				//ctx.String(403, "无权限访问")
				//ctx.Abort()
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+",商户异常", nil)
				ctx.Abort()
				return
			}
			pathArr := strings.Split(path, "/")
			suffix := "/" + pathArr[len(pathArr)-1]

			//验证是否允许访问api 包含了ip验证
			ok, err := transferSecurityService.VerifyApiPermission(suffix, coinName, ip, mch.Id)
			if err != nil {
				log.Errorf("验证路径权限异常：%s", err.Error())
				//ctx.String(403, "无权限访问")
				httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+","+err.Error(), nil)

				ctx.Abort()
				return
			}
			if ok {
				log.Infof("AuthApiV4 验证结束")
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
				ctx.Next()
				return
			}
		}
		//ctx.String(403, "无权限访问")
		httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, httpresp.GetMsg(httpresp.NO_PERMISSION)+",缺少商户名参数", nil)
		ctx.Abort()
		return
	}
}
