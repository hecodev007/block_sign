package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/service"
	"io/ioutil"
	"strconv"
	"time"
)

func CheckApiSignSecureOldVersion() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			err error
			ok  bool
		)

		data, _ := ioutil.ReadAll(ctx.Request.Body)
		log.Infof("CheckApiSignSecureOldVersion 传入内容体：%s", string(data))
		//获取完之后要重新设置进去，不然会丢失
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

		clientId := ctx.PostForm("client_id")
		nonce := ctx.PostForm("nonce")
		timestampStr := ctx.PostForm("ts")
		sfrom := ctx.PostForm("mch")
		outOrderId := ctx.PostForm("outOrderId")
		coinName := ctx.PostForm("coinName")
		amount := ctx.PostForm("amount")
		toAddress := ctx.PostForm("toAddress")
		memo := ctx.PostForm("memo")
		callBack := ctx.PostForm("callBack")
		contractAddress := ctx.PostForm("contractAddress")
		tokenName := ctx.PostForm("tokenName")
		walletId := ctx.PostForm("walletId")
		isCustody := ctx.PostForm("isCustody")

		hash := ctx.PostForm("hash")
		sign := ctx.PostForm("sign")

		ts, _ := strconv.ParseInt(timestampStr, 10, 64)
		nowTs := time.Now().Unix()
		log.Infof("client_id:%s", clientId)
		log.Infof("ts:%d", ts)
		log.Infof("now:%d", nowTs)
		log.Infof("nonce:%s", nonce)
		log.Infof("hash:%s", hash)
		log.Infof("sign:%s", sign)
		//单位秒
		if (nowTs-ts) > 60 || (nowTs-ts) < -60 {
			//超过服务器时间60秒自动失败
			//ctx.String(403, "服务器时间不一致，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "服务器时间不一致，签名失败", nil)
			ctx.Abort()
			return
		}
		if clientId == "" || ts == 0 || nonce == "" || sign == "" {
			//ctx.String(403, "参数异常，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "参数异常，签名失败", nil)
			ctx.Abort()
			return
		}

		params := map[string]string{}
		params["client_id"] = clientId
		params["nonce"] = nonce
		params["ts"] = timestampStr
		params["sfrom"] = sfrom
		params["outOrderId"] = outOrderId
		params["coinName"] = coinName
		params["amount"] = amount
		params["toAddress"] = toAddress
		params["memo"] = memo
		params["callBack"] = callBack
		params["contractAddress"] = contractAddress
		params["tokenName"] = tokenName
		params["walletId"] = walletId
		params["isCustody"] = isCustody

		if ok, err = service.VerifyApiSignV2(clientId, hash, sign, params); !ok {
			errMsg := "invalid sign"
			if err != nil {
				log.Error(err.Error())
				errMsg = err.Error()
			}
			//ctx.String(403, err.Error())
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, errMsg, nil)
			ctx.Abort()
			return
		}
		//不需要像json向下传递
		ctx.Next()
	}
}

type ApplyTrxRequest struct {
	ClientId        string `json:"clientId"`
	Nonce           string `json:"nonce"`
	Timestamp       string `json:"timestamp"`
	OutOrderId      string `json:"outOrderId"`
	CoinName        string `json:"coinName"`
	Amount          string `json:"amount"`
	ToAddress       string `json:"toAddress"`
	Memo            string `json:"memo"`
	CallBack        string `json:"callBack"`
	ContractAddress string `json:"contractAddress"`
	TokenName       string `json:"tokenName"`
	Hash            string `json:"hash"`
	Signature       string `json:"signature"`
}

func getRawData(ctx *gin.Context) ([]byte, error) {
	data, err := ctx.GetRawData()
	if err != nil {
		log.Errorf("getRawData获取原始入参数据失败 %v", err)
		return nil, err
	}
	return data, err
}

func jsonStructToMap(entity *ApplyTrxRequest) (map[string]string, error) {
	// 结构体转json
	strRet, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}
	// json转map
	var mRet map[string]string
	err1 := json.Unmarshal(strRet, &mRet)
	if err1 != nil {
		return nil, err1
	}
	return mRet, nil
}

func CheckApiSignSecure() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, _ := getRawData(ctx)
		request := &ApplyTrxRequest{}
		if err := json.Unmarshal(data, request); err != nil {
			httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "入参有误", nil)
			ctx.Abort()
			return
		}
		log.Infof("CheckApiSignSecure 传入内容体：%+v", request)
		//获取完之后要重新设置进去，不然会丢失
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

		mapParam, err := jsonStructToMap(request)
		if err != nil {
			log.Errorf("jsonStructToMap出错 %v", err)
			httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "入参有误", nil)
			ctx.Abort()
			return
		}
		timestamp, _ := strconv.ParseInt(request.Timestamp, 10, 64)
		now := time.Now().Unix()
		//单位秒
		if (now-timestamp) > 60 || (now-timestamp) < -60 {
			//超过服务器时间60秒自动失败
			//ctx.String(403, "服务器时间不一致，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "服务器时间不一致，签名失败", nil)
			ctx.Abort()
			return
		}
		if request.ClientId == "" || timestamp == 0 || request.Nonce == "" || request.Signature == "" {
			//ctx.String(403, "参数异常，签名失败")
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, "参数异常，签名失败", nil)
			ctx.Abort()
			return
		}

		if ok, err := service.VerifyApiSignV2(request.ClientId, request.Hash, request.Signature, mapParam); !ok {
			errMsg := "invalid sign"
			if err != nil {
				log.Error(err.Error())
				errMsg = err.Error()
			}
			//ctx.String(403, err.Error())
			httpresp.HttpRespCodeError(ctx, httpresp.NO_PERMISSION, errMsg, nil)
			ctx.Abort()
			return
		}
		//不需要像json向下传递
		ctx.Next()
	}
}
