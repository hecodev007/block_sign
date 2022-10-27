package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"strings"
)

//统一交易接口
func Transfer(ctx *gin.Context) {
	params := model.TransferParams{}
	pathArr := strings.Split(ctx.Request.URL.Path, "/")
	path := "/" + pathArr[len(pathArr)-1]

	ctx.BindJSON(&params)
	//转换为小写
	params.CoinName = strings.ToLower(params.CoinName)
	coinName := params.CoinName
	if params.TokenName != "" {
		coinName = strings.ToLower(params.TokenName)
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
	mchId := mch.Id
	//1：验证币种是否支持
	ok, _ := api.TransferSecurityService.VerifyCoin(coinName, mchId)
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.COIN_ERROR, httpresp.GetMsg(httpresp.COIN_ERROR), nil)
		return
	}

	//2：验证币种关闭还是开放
	ok, _ = api.TransferSecurityService.VerifyCoinPermission(coinName, mchId)
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
	ok, _ = api.TransferSecurityService.VerifyCoinDecimal(coinName, params.Amount)
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
		httpresp.HttpRespCodeError(ctx, httpresp.ERROR_TOADDRESS, httpresp.GetMsg(httpresp.ERROR_TOADDRESS), nil)
		return
	}

	//6：验证商户余额
	ok, _, err = api.TransferSecurityService.VerifyMchBalance(params.CoinName, params.ContractAddress, params.Amount, mchId)
	if err != nil {
		log.Errorf("验证商户余额异常,币种：%s,商户：%s, error: %s", coinName, params.Sfrom, err.Error())
	}
	if !ok {
		httpresp.HttpRespCodeError(ctx, httpresp.MchAmountNotEnough, httpresp.GetMsg(httpresp.MchAmountNotEnough), nil)
		return
	}

	//7：单笔出账限额，每小时出账限额，每日出账限额
	ok, err = api.TransferSecurityService.VerifyRisk(coinName, params.Amount, mchId)
	if err != nil {
		log.Errorf("风控验证异常,币种：%s,商户：%s, error: %s", params.CoinName, params.Sfrom, err.Error())
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
	if params.CallBack == "" {
		callBackUrl := getApiCallBackUrl(mchId, params.CoinName, path)
		if callBackUrl == "" {
			log.Errorf("商户：%s,缺少币种回调地址设置，币种 :%s,api :%s", params.Sfrom, params.CoinName, path)
			httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
			return
		}
	} else {
		callBackUrl = params.CallBack
	}

	applyId := int64(0)
	//判断币种属于哪种模型，进行对应的存储
	if global.TransferModel[params.CoinName] == model.TransferModelUtxo {
		applyId, err = api.OrderService.SaveTransferByUtxo(params, callBackUrl)
	} else {
		applyId, err = api.OrderService.SaveTransferByAccount(params, callBackUrl)
	}
	if err != nil {
		log.Errorf("保存商户订单异常:%s", err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, httpresp.GetMsg(httpresp.FAIL), nil)
		return
	}
	//成功
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), applyId)
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
