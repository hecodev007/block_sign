package custody

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"strconv"
	"strings"
	"time"
)

//BatchCreateAddress 商户钱包地址创建（批量创建地址）
func BatchCreateAddress(ctx *gin.Context) {
	var (
		err          error
		applyAddrReq model.ApplyAddrReq
		as           []*entity.FcGenerateAddressList
		retData      []string
	)
	clientId := ctx.PostForm("api_key")
	outOrderId := ctx.PostForm("outOrderId")
	coinName := ctx.PostForm("coinName")
	coinName = strings.ToLower(coinName)
	num, _ := strconv.ParseInt(ctx.PostForm("num"), 10, 64)
	applyAddrReq = model.ApplyAddrReq{
		OutOrderId: outOrderId,
		CoinName:   coinName,
		Num:        num,
		ApiSignParams: util.ApiSignParams{
			ClientId: clientId,
		},
	}
	if clientId == "" || outOrderId == "" {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求参数错误 error: %s", outOrderId).Error(), nil)
		return
	}
	if _, ok := global.CoinDecimal[coinName]; !ok {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求币种错误: %s", coinName).Error(), nil)
		return
	}

	if num <= 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求币种数量错误 %d", num).Error(), nil)
		return
	}
	if as, err = service.AssignMchAddrsV2(applyAddrReq); err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求获取地址异常: %s", err.Error()).Error(), nil)
		return
	}
	retData = make([]string, 0, len(as))
	addrs := make([]string, 0)
	for _, a := range as {
		if coinName == "bsv" || coinName == "bch" {
			retData = append(retData, a.CompatibleAddress)
			addrs = append(addrs,a.CompatibleAddress)
		} else {
			retData = append(retData, a.Address)
			addrs = append(addrs,a.Address)
		}
	}
	if _, ok := api.RegisterService[strings.ToLower(coinName)]; ok {
		//地址注册
		resp, err1 := api.RegisterService[strings.ToLower(coinName)].RegisterToNode(retData)
		if err1 != nil {
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("注册地址到节点异常：%s", err1.Error()))
		} else {
			log.Infof("%s,注册地址结果：%s", coinName, string(resp))
		}
	}
	backData := make(map[string]interface{})
	backData["list"] = retData
	syncToAddrMgr(coinName, addrs)
	httpresp.HttpRespOK(ctx, "", backData)
}

//BatchCreateLotCoinAddress 商户多币种创建地址
func BatchCreateLotCoinAddress(ctx *gin.Context) {
	var err error
	clientId := ctx.PostForm("api_key")
	outOrderId := ctx.PostForm("outOrderId")
	coinName := ctx.PostForm("coinName")
	coinName = strings.ToLower(coinName)
	num, _ := strconv.ParseInt(ctx.PostForm("num"), 10, 64)

	if clientId == "" || outOrderId == "" {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求参数错误 error: %s", outOrderId).Error(), nil)
		return
	}
	if _, ok := global.CoinDecimal[coinName]; !ok {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求币种错误: %s", coinName).Error(), nil)
		return
	}

	if num <= 0 {
		httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求币种数量错误 %d", num).Error(), nil)
		return
	}
	retData := make(map[string]string)
	coinArr := strings.Split(coinName, ",")
	addrs := make([]string, 0)
	for _, coin := range coinArr {
		applyAddrReq := model.ApplyAddrReq{
			OutOrderId: outOrderId,
			CoinName:   coin,
			Num:        num,
			ApiSignParams: util.ApiSignParams{
				ClientId: clientId,
			},
		}
		var as []*entity.FcGenerateAddressList
		//	retData      []string
		if as, err = service.AssignMchAddrsV2(applyAddrReq); err != nil {
			httpresp.HttpRespError(ctx, httpresp.FAIL, fmt.Errorf("请求获取地址异常: %s", err.Error()).Error(), nil)
			return
		}

		for _, a := range as {
			if coin == "bsv" || coin == "bch" {
				retData[coin] = a.CompatibleAddress
				addrs = append(addrs, a.CompatibleAddress)
			} else {
				retData[coin] = a.Address
				addrs = append(addrs, a.Address)
			}

		}

		if _, ok := api.RegisterService[strings.ToLower(coin)]; ok {
			//地址注册
			resp, err1 := api.RegisterService[strings.ToLower(coin)].RegisterToNode([]string{coin})
			if err1 != nil {
				dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("注册地址到节点异常：%s", err1.Error()))
			} else {
				log.Infof("%s,注册地址结果：%s", coin, string(resp))
			}
		}
	}

	backData := make(map[string]interface{})
	backData["list"] = retData
	syncToAddrMgr(coinName, addrs)
	httpresp.HttpRespOK(ctx, "", backData)
}

func BackAddress(ctx *gin.Context) {
	var err error
	var rsp model.AddressRsp
	err = ctx.BindJSON(&rsp)
	if err != nil || rsp.Mch == "" {
		log.Errorf("bind err = %v", err)
		httpresp.HttpRespError(ctx, httpresp.FAIL, httpresp.MsgFlags[httpresp.PARAM_FORMAT_ERROR], nil)
		return
	}
	err = service.SaveBackAddress(rsp)
	if err != nil {
		httpresp.HttpRespError(ctx, httpresp.FAIL, err.Error(), nil)
		return
	}
	httpresp.HttpRespOK(ctx, "", nil)
}

func syncToAddrMgr(coinCode string, addresses []string) {
	log.Infof("custody (新)拉取地址同步到addrmanagement codeCode=%s 地址=%v", coinCode, addresses)
	models := make([]entity.Addresses, 0)
	for _, a := range addresses {
		models = append(models, entity.Addresses{
			CreatedAt: time.Now(),
			Address:   a,
			CoinType:  coinCode,
			Status:    "used",
			ComeFrom:  "merchant",
			UserId:    5, // 对应 user表id 历史遗留问题，直接使用5
		})
	}
	rows, err := dao.AmAddBatchAddresses(models)
	if err != nil {
		log.Infof("custody 拉取地址同步到addrmanagement 插入到数据失败 %v", err)
		return
	}
	log.Infof("custody 拉取地址同步到addrmanagement 插入到数据受影响行数 %d", rows)
}
