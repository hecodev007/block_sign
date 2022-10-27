package custody

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	v3 "github.com/group-coldwallet/blockchains-go/router/api/v3"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"time"
)

//OperateBalance /提现接口
func OperateBalance(ctx *gin.Context) {
	clientId := ctx.PostForm("api_key")
	callBack := ctx.PostForm("callBack")
	outOrderId := ctx.PostForm("outOrderId")
	coinName := ctx.PostForm("coinName") //主链币
	coinName = strings.ToLower(coinName)
	amount, _ := decimal.NewFromString(ctx.PostForm("amount"))
	toAddress := ctx.PostForm("toAddress")
	tokenName := strings.ToLower(ctx.PostForm("tokenName")) //代币
	contractAddress := strings.ToLower(ctx.PostForm("contractAddress"))
	memo := ctx.PostForm("memo")
	fee, _ := decimal.NewFromString(ctx.PostForm("fee"))
	isForce := false
	if ctx.PostForm("isForce") == "true" {
		isForce = true
	}
	log.Infof("clientId :%s", clientId)
	mchName := ""
	mchVal, ok1 := global.MchBaseInfo[clientId]
	if !ok1 {
		global.ReloadMchBaseInfo()
		mchVal, ok1 = global.MchBaseInfo[clientId]
		if !ok1 {
			log.Errorf("clientId 无配置 clientId:%s", clientId)
			httpresp.HttpRespCodeError(ctx, httpresp.UNKNOWN_ERROR, fmt.Sprintf("商户不存在"), nil)
			return
		}
	}

	if mchVal != nil {
		mchName = mchVal.MchName
	}
	log.Infof("mchName :%s", mchName)
	params := &model.TransferParams{
		Sfrom:           mchName,
		CallBack:        callBack,
		OutOrderId:      outOrderId,
		CoinName:        coinName,
		Amount:          amount,
		ToAddress:       toAddress,
		TokenName:       tokenName,
		ContractAddress: contractAddress,
		Memo:            memo,
		Fee:             fee,
		IsForce:         isForce,
	}
	log.Infof("TransferParams : %+v", params)

	//校验主链币是否一致
	//转换为小写
	//如果存在tokenName直接替换
	mainName := strings.ToLower(coinName)
	if tokenName != "" {
		//校验主链币是否一致
		tokenInfo := global.CoinDecimal[tokenName]
		if tokenInfo == nil {
			log.Errorf("代币无配置 main:%s，token：%s", coinName, tokenName)
			httpresp.HttpRespCodeError(ctx, httpresp.UNKNOWN_ERROR, fmt.Sprintf("代币无配置 main:%s，token：%s", coinName, tokenName), nil)
			return
		}
		if strings.ToLower(tokenInfo.Token) != contractAddress {
			log.Errorf("token 地址不匹配：%s <> %s ", strings.ToLower(tokenInfo.Token), contractAddress)
			httpresp.HttpRespCodeError(ctx, httpresp.UNKNOWN_ERROR, fmt.Sprintf("token 地址不匹：%s <> %s ", strings.ToLower(tokenInfo.Token), contractAddress), nil)
			return
		}
		mainInfo := global.CoinDecimal[coinName]
		if mainInfo == nil {
			log.Errorf("币种错误 main:%s，token：%s", coinName, tokenName)
			httpresp.HttpRespCodeError(ctx, httpresp.UNKNOWN_ERROR, fmt.Sprintf("币种错误 main:%s，token：%s", coinName, tokenName), nil)
			return
		}
		if mainInfo.Id != tokenInfo.Pid {
			//币种关联不对称
			log.Errorf("币种不对称 main:%s，token：%s", coinName, tokenName)
			httpresp.HttpRespCodeError(ctx, httpresp.UNKNOWN_ERROR, fmt.Sprintf("币种不对称 main:%s，token：%s", coinName, tokenName), nil)
			return
		}
		mainName = strings.ToLower(tokenName)
	}
	//基本参数校验
	err := params.CheckParams()
	if err != nil {
		log.Errorf("Transfer params异常:%s", err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.PARAM_ERROR, httpresp.GetMsg(httpresp.PARAM_ERROR), nil)
		return
	}

	//验证流程（9步流程）
	mch, err := api.MchService.GetAppId(params.Sfrom)
	if err != nil || mch.Id == 0 {
		log.Errorf("Transfer 查询商户ID异常:%s", err.Error())
		//商户错误
		httpresp.HttpRespCodeError(ctx, httpresp.MERCHANT_ERROR, httpresp.GetMsg(httpresp.MERCHANT_ERROR), nil)
		return
	}
	log.Infof("根据sfrom查询到mch信息: %+v", mch)
	mchId := mch.Id
	//1：验证币种是否支持
	ok, _ := api.TransferSecurityService.VerifyCoin(mainName, mchId)
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.COIN_ERROR, httpresp.GetMsg(httpresp.COIN_ERROR), nil)
		return
	}
	//2：验证币种关闭还是开放
	ok, _ = api.TransferSecurityService.VerifyCoinPermission(mainName, mchId)
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.NOT_POWER, httpresp.GetMsg(httpresp.NOT_POWER), nil)
		return
	}

	//3.验证商户是否已经分配地址
	isAssign, err := api.TransferSecurityService.IsAssignAddress(params.CoinName, mchId)
	if err != nil {
		log.Errorf("验证商户是否已经分配地址异常,币种：%s,商户：%s, error: %s", params.CoinName, params.Sfrom, err.Error())
	}
	if !isAssign {
		log.Errorf("商户未分配地址,币种：%s,商户：%s", params.CoinName, params.Sfrom)
		httpresp.HttpRespCodeError(ctx, httpresp.ADR_NONE, httpresp.GetMsg(httpresp.ADR_NONE), nil)
		return
	}
	//4：验证币种精度
	ok, _ = api.TransferSecurityService.VerifyCoinDecimal(mainName, params.Amount)
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.AmountDecimalError, httpresp.GetMsg(httpresp.AmountDecimalError), nil)
		return
	}
	//5：验证合法地址
	ok, err = api.TransferSecurityService.VerifyAddress(params.ToAddress, params.CoinName)
	if err != nil {
		log.Errorf("接收地址验证,币种：%s,商户：%s, error: %s", params.CoinName, params.Sfrom, err.Error())
	}
	if !ok {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单：%s,接收地址验证异常,币种：%s,地址：%s,商户：%s, error: %v", params.OutOrderId, mainName, params.ToAddress, params.Sfrom, err))
		httpresp.HttpRespCodeError(ctx, httpresp.ERROR_TOADDRESS, httpresp.GetMsg(httpresp.ERROR_TOADDRESS), nil)
		return
	}

	//6：验证商户余额
	//ok, _, err = api.TransferSecurityService.VerifyMchBalance(mainName, params.ContractAddress, params.Amount, mchId)
	//if err != nil {
	//	log.Errorf("订单：%s,验证商户余额异常,币种：%s,商户：%s, error: %s", params.OutOrderId, mainName, params.Sfrom, err.Error())
	//}
	//if !ok {
	//	dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单：%s,验证商户余额异常,币种：%s,商户：%s, error: %v", params.OutOrderId, mainName, params.Sfrom, err))
	//	httpresp.HttpRespCodeError(ctx, httpresp.MchAmountNotEnough, httpresp.GetMsg(httpresp.MchAmountNotEnough), nil)
	//	return
	//}
	//7：单笔出账限额，每小时出账限额，每日出账限额
	ok, err = api.TransferSecurityService.VerifyRisk(mainName, params.Amount, mchId)
	if err != nil {
		log.Errorf("风控验证异常,币种：%s,商户：%s, error: %s", mainName, params.Sfrom, err.Error())
	}
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.SingleQuotaLimit, httpresp.GetMsg(httpresp.SingleQuotaLimit), nil)
		return
	}
	//8 查看是否是配置中的币种
	_, ok = global.TransferModel[params.CoinName]
	if !ok {
		log.Errorf("配置文件无该币种配置，%s", params.CoinName)
		httpresp.HttpRespCodeError(ctx, httpresp.COIN_EMPTY, httpresp.GetMsg(httpresp.COIN_EMPTY), nil)
		return
	}
	//9.验证订单是否重复
	has, err := api.TransferSecurityService.IsDuplicateApplyOrder(params.OutOrderId, params.Sfrom)
	if has {
		if err != nil {
			log.Errorf("商户：%s ,查询重复异常，outOrderId :%s,error:%s", params.Sfrom, params.OutOrderId, err.Error())
		}
		log.Errorf("商户：%s ,重复订单，outOrderId :%s", params.Sfrom, params.OutOrderId)
		httpresp.HttpRespCodeError(ctx, httpresp.OUT_ORDER_ID, httpresp.GetMsg(httpresp.OUT_ORDER_ID), nil)
		return
	}

	callBackUrl := ""
	if strings.TrimSpace(params.CallBack) == "" {
		log.Infof("callBackUrl为空，从api_power表获取")
		pathArr := strings.Split(ctx.Request.URL.Path, "/")
		path := "/" + pathArr[len(pathArr)-1]
		if "/applyTransactionSecure" == path {
			path = "/applyTransaction"
		}
		callBackUrl = getApiCallBackUrl(mchId, params.CoinName, path)
		log.Infof("callBackUrl为空，从api_power表获取结果 %s", callBackUrl)
		if callBackUrl == "" {
			log.Errorf("商户：%s,缺少币种回调地址设置，币种 :%s,api :%s", params.Sfrom, params.CoinName, path)
			httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
			return
		}
	} else {
		callBackUrl = params.CallBack
	}

	log.Infof("商户回调地址为：%s", callBackUrl)
	if callBackUrl == "" {
		log.Errorf("商户回调地址为空：%s", callBackUrl)
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
		return
	}

	orderId := int64(0)
	//判断币种属于哪种模型，进行对应的存储
	if global.TransferModel[params.CoinName] == model.TransferModelUtxo {
		orderId, err = api.OrderService.SaveTransferByUtxo(*params, callBackUrl)
	} else {
		orderId, err = api.OrderService.SaveTransferByAccount(*params, callBackUrl)
	}
	if err != nil {
		log.Errorf("保存商户订单异常:%s", err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
		return
	}
	createAt := time.Now().Unix()
	//保存到第二数据库
	//coinType :=coinName
	//if tokenName!="" {
	//	coinType = tokenName
	//}
	//coin:=global.CoinDecimal[coinType]
	//log.Infof("coinType=%s,decimal=%d,amount=%s",coinType,coin.Decimal,amount.String())
	cryptData, err := util.CreateCheckApplyContent(util.RawContent{
		OutOrderId:       outOrderId,
		ToAddress:        toAddress,
		MainCoin:         coinName,
		Token:            tokenName,
		ToAmountFloatStr: amount.String(),
		CreateAt:         fmt.Sprintf("%d", createAt),
	})
	if err != nil {
		log.Errorf("保存加密订单错误cd 订单异常:%s", err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
		return
	}

	ca := &entity.CheckApply{
		ApplyId:  orderId,
		Content:  cryptData,
		CreateAt: createAt,
		UpdateAt: createAt,
	}
	_, err = ca.Add()
	if err != nil {
		log.Errorf("保存到第二数据库错误:%s", err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
		return
	}
	log.Infof("订单 %s 接收未完成", outOrderId)
	//成功
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), orderId)
}

//数据库设计是每个币种的每个api有不同的回调地址（有点过于复杂了）
func getApiCallBackUrl(mchId int, coinName, apiPath string) string {
	mchAuth := global.MchAuth[mchId]
	if mchAuth != nil {
		coins := mchAuth.Api[coinName]
		if coins != nil {
			pathAuths := coins.Auth[apiPath]
			if pathAuths != nil {
				return pathAuths.CallBackUrl
			}
		}
	}
	return ""
}

//CoinBalance 余额查询
func CoinBalance(c *gin.Context) {
	clientId := c.PostForm("api_key")
	mchVal := global.MchBaseInfo[clientId]
	if mchVal == nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}

	mchId := mchVal.AppId
	redisHelper, err := util.AllocRedisClient()
	defer redisHelper.Close()

	key := v3.GetCacheKeyV2(strconv.Itoa(mchId))

	cacheData := v3.GetFromCache(redisHelper, key)
	if cacheData != nil {
		log.Infof("GetMchAllBalanceV2 AppId=%d 从缓存获取数据成功", mchId)
		httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), *cacheData)
		return
	}

	log.Infof("GetMchAllBalanceV2 AppId=%d 从缓存获取数据失败，继续从DB获取", mchId)
	result, err := api.BalanceService.GetMchAllBalanceV2(mchId)
	if err != nil {
		log.Errorf("GetMchAllBalanceV2 error:%s", err.Error())
		httpresp.HttpRespErrorOnly(c)
		return
	}
	log.Infof("GetMchAllBalanceV2 AppId=%d 设置数据到缓存", mchId)
	//v3.SetToCache(redisHelper, key, v3.MchAllBalanceExpireSecV2, &result)

	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
}


