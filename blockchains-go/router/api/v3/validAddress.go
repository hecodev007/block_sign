package v3

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"strings"
)

func ValidAddress(ctx *gin.Context) {
	coinName := ctx.PostForm("coin")
	if coinName == "" {
		log.Error("valid address params异常: coin is null")
		httpresp.HttpRespCodeError(ctx, httpresp.PARAM_ERROR, httpresp.GetMsg(httpresp.PARAM_ERROR), nil)
		return
	}

	coinName = strings.ToLower(coinName)
	address := ctx.PostForm("address")

	if address == "" {
		log.Error("valid address params异常: address is null")
		httpresp.HttpRespCodeError(ctx, httpresp.PARAM_ERROR, httpresp.GetMsg(httpresp.PARAM_ERROR), nil)
		return
	}

	log.Infof("ValidAddress address=%s coin=%s", address, coinName)

	ok, err := api.TransferSecurityService.VerifyAddress(address, coinName)
	if err != nil || !ok {
		log.Errorf("valid address error: %v, isOK=[%v] ", err, ok)

		if err.Error() == "contractAddrNotAllow" {
			httpresp.HttpRespCodeError(ctx, httpresp.ContractAddrNotAllow, httpresp.GetMsg(httpresp.ContractAddrNotAllow), nil)
		} else {
			httpresp.HttpRespCodeError(ctx, httpresp.ValidAddressError, httpresp.GetMsg(httpresp.ValidAddressError), nil)
		}
		return
	}
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), ok)
}

func ValidInsideAddress(ctx *gin.Context) {

	address := ctx.PostForm("address")
	if address == "" {
		log.Error("valid address params异常: address is null")
		httpresp.HttpRespCodeError(ctx, httpresp.PARAM_ERROR, httpresp.GetMsg(httpresp.PARAM_ERROR), nil)
		return
	}
	log.Infof("ValidAddress address=%s ", address)
	ok, err := api.TransferSecurityService.IsInsideAddress(address)
	if err != nil || !ok {
		log.Errorf("valid address error: %v, isOK=[%v] ", err, ok)
		if err.Error() == "contractAddrNotAllow" {
			httpresp.HttpRespCodeError(ctx, httpresp.ContractAddrNotAllow, httpresp.GetMsg(httpresp.ContractAddrNotAllow), nil)
		} else {
			httpresp.HttpRespCodeError(ctx, httpresp.ValidAddressError, httpresp.GetMsg(httpresp.ValidAddressError), nil)
		}
		return
	}
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), ok)
}
