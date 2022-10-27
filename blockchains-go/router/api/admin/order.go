package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
)

//修复指定订单
func RepairOrder(ctx *gin.Context) {
	type body struct {
		OutOrderId string `json:"outOrderId"`
	}
	params := new(body)
	ctx.BindJSON(params)
	if params.OutOrderId == "" {
		httpresp.HttpRespCodeError(ctx, httpresp.FAIL, "Miss outOrderId", nil)
		return
	}
	applyOrder, err := api.OrderService.GetApplyOrder(params.OutOrderId)
	if err != nil {
		log.Errorf("没有这个订单：%s", params.OutOrderId)
		return
	}

	if applyOrder.Status == int(entity.ApplyStatus_TransferOk) {
		log.Infof("apply订单状态异常：%s,状态%s", params.OutOrderId, applyOrder.Status)
		httpresp.HttpRespErrorOnly(ctx)
		return
	}

	//查询成功记录
	order, err := api.WalletOrderService.GetSuccessColdOrder(params.OutOrderId)
	if err != nil || order.Status != status.BroadcastStatus.Int() {
		log.Errorf("查询order订单异常：%s", params.OutOrderId)
		return
	}
	if order.Status != status.BroadcastStatus.Int() {
		log.Errorf("查询order订单状态异常：%s", params.OutOrderId)
		return
	}
	err = api.OrderService.SendApplyTransferSuccess(applyOrder.Id)
	if err != nil {
		log.Errorf("修改状态异常：%s,error:%s", params.OutOrderId, err.Error())
		return
	}
	//推送
	api.OrderService.NotifyToMchByOutOrderId(params.OutOrderId)
	httpresp.HttpRespCodeOkOnly(ctx)
}
