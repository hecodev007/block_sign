package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"io/ioutil"
	"strings"
)

func TransPush(ctx *gin.Context) {
	var (
		err      error
		data     []byte
		pushBase model.PushBaseBlockInfo
	)
	if data, err = ioutil.ReadAll(ctx.Request.Body); err != nil {
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "get post data error", nil)
		return
	}
	if len(data) == 0 {
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "post data is empty", nil)
		return
	}

	if err = json.Unmarshal(data, &pushBase); err != nil {
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "post data is not json", nil)
		return
	}

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "write mq fail", nil)
		return
	}
	defer redisHelper.Close()

	log.Infof("TransPush %s", string(data))

	log.Infof("pushBase.Type:%d", pushBase.Type)
	switch pushBase.Type {
	case model.PushTypeTX:
		redisHelper.LeftPush("ticket_list_new", string(data))
	case model.PushTypeConfir:
		if strings.ToLower(pushBase.CoinName) != "btc-stx" {
			redisHelper.LeftPush("confirm_list_new", string(data))
		}
	case model.PushTypeAccountTX:
		redisHelper.LeftPush("account_list_new", string(data))
	case model.PushTypeAccountConfir:
		redisHelper.LeftPush("confirm_list_new", string(data))

	case model.PushTypeEosTX:
		redisHelper.LeftPush("eos_push_list_new", string(data))
		break
	case model.BtmPushTypeTX:
		redisHelper.LeftPush("btm_push_list_new", string(data))
		break
	case model.BtmPushTypeConfir:
		redisHelper.LeftPush("confirm_list_new", string(data))
		break
	default:
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "unknow push type", nil)
		return
	}

	httpresp.HttpRespCodeOkOnly(ctx)
}

func TransPushTest(ctx *gin.Context) {
	address := ctx.Query("address")
	coin := ctx.Query("coinname")
	res := &entity.FcGenerateAddressList{}
	result, err := dao.TransPushGet(res, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, coin)
	if !result || err != nil {
		log.Debug(err)
		res.PlatformId = 0
		res.Type = 0
	}
	log.Debug(res)
	httpresp.HttpRespCodeOkOnly(ctx)
}

//根据订单号主动推送订单
func TransPushOrder(ctx *gin.Context) {
	type body struct {
		OutOrderId string `json:"outOrderId"`
	}
	params := new(body)
	ctx.BindJSON(params)
	if params.OutOrderId == "" {
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "Miss outOrderId", nil)
		return
	}
	log.Infof("TransPushOrder OutOrderId=%s", params.OutOrderId)
	err := api.OrderService.NotifyToMchByOutOrderId(params.OutOrderId)
	if err != nil {
		log.Errorf("outOrderId：%s,推送商户失败,err:%s", params.OutOrderId, err.Error())
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "回调失败", err.Error())
		return
	}
	httpresp.HttpRespCodeOkOnly(ctx)
}
