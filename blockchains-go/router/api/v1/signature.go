package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/service"
	"io/ioutil"
	"strings"
)

func Signature(ctx *gin.Context) {
	var (
		err      error
		data     []byte
		signData map[string]string
		mchName  string
		ok       bool
	)
	if data, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "get post data error", nil)
		return
	}
	if len(data) == 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is empty", nil)
		return
	}

	if signData, err = model.DecodeSignData(data); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is not json", nil)
		return
	}
	//trim
	for k, v := range signData {
		signData[k] = strings.Trim(v, " ")
	}

	mchName = signData["sfrom"]
	if len(mchName) == 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is not include 'sfrom'", nil)
		return
	}

	if len(signData[service.SignKey]) == 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is not include '"+service.SignKey+"'", nil)
		return
	}

	if ok, err = service.VerifySign(mchName, signData); !ok {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("signature no pass: %w", err).Error(), nil)
		return
	}
	httpresp.HttpRespOkOnly(ctx)
}

func ApplyAddress(ctx *gin.Context) {
	var (
		err          error
		applyAddrReq model.ApplyAddrReq
		as           []*entity.FcGenerateAddressList
		data         []byte
		retData      []string
	)

	//符合BindJSON条件, 必须contentType是json
	//if err = ctx.BindJSON(&applyAddrReq);err != nil {
	//	httpresp.HttpRespError(ctx,httpresp.FAIL,"post data is not ApplyAddr json",nil)
	//	return
	//}

	if data, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "get post data error", nil)
		return
	}
	if len(data) == 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is empty", nil)
		return
	}

	if applyAddrReq, err = model.DecodeApplyAddrData(data); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, "post data is not ApplyAddr json", nil)
		return
	}
	if as, err = service.AssignMchAddrs(applyAddrReq); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("ApplyAddress AssignMchAddrs error: %w", err).Error(), nil)
		return
	}
	retData = make([]string, 0, len(as))
	for _, a := range as {
		retData = append(retData, a.Address)
	}
	httpresp.HttpRespOK(ctx, "", retData)
}
