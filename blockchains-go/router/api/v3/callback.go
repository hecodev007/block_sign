package v3

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
)

func repairOrderTx() {

}

func WalletCallBack(ctx *gin.Context) {
	bodyByte, _ := ioutil.ReadAll(ctx.Request.Body)
	log.Infof("接收wallet回调的内容是：%s", string(bodyByte))
	info := new(model.WalltCallBack)
	json.Unmarshal(bodyByte, info)
	if info.OuterOrderNo == "" || info.OrderNo == "" || info.Status < status.BroadcastStatus.Int() || info.MchName == "" {
		log.Error("必要参数异常")
		httpresp.HttpRespCodeErrOnly(ctx)
		return
	}
	//处理订单
	applyOrder, err := api.OrderService.GetApplyOrder(info.OuterOrderNo)
	if err != nil {
		log.Errorf("没有这个订单：%s", info.OuterOrderNo)
		httpresp.HttpRespCodeErrOnly(ctx)
		return
	}

	if applyOrder.Status != int(entity.ApplyStatus_CreateOk) {
		if applyOrder.Status == int(entity.ApplyStatus_Merge) {
			log.Infof("归集订单忽略 :%s", info.OuterOrderNo)
		} else if applyOrder.Status == int(entity.ApplyStatus_Fee) {
			log.Infof("打手续费订单忽略 :%s", info.OuterOrderNo)
		} else {
			log.Errorf("订单状态异常：%s,状态%d", info.OuterOrderNo, applyOrder.Status)
		}
		httpresp.HttpRespCodeOkOnly(ctx)
		return
	}
	mch, err := api.MchService.GetAppId(info.MchName)
	if err != nil {
		log.Errorf("没有这个商户：%s", info.MchName)
		httpresp.HttpRespCodeErrOnly(ctx)
		return
	}
	if applyOrder.AppId != mch.Id {
		log.Errorf("商户校验异常：%s", info.MchName)
		httpresp.HttpRespCodeErrOnly(ctx)
		return
	}
	//解冻
	if info.Status > status.BroadcastStatus.Int() {
		if strings.ToLower(applyOrder.CoinName) == "ckb" {
			orders, err := api.WalletOrderService.GetColdOrder(info.OuterOrderNo)
			if err == nil && len(orders) > 0 {
				order := orders[0]
				createData := []byte(order.CreateData)
				if len(createData) >= 0 {
					var cd model.CkbCreateData
					err = json.Unmarshal(createData, &cd)
					if err == nil && len(cd.Inputs) > 0 {
						var amount = decimal.NewFromInt(0)
						for _, in := range cd.Inputs {
							inAmount, _ := decimal.NewFromString(in.Amount)
							amount = amount.Add(inAmount)
							err = dao.FcTransPushUnFreezeUtxo("ckb", in.Txid, in.Index, in.Address)
							if err != nil {
								log.Errorf("Ckb解冻失败，更新fc_trans_push is_spent error：%v", err)
							}
						}
						log.Infof("ckb需要解冻的金额为： %s", amount.String())
						err = dao.FcAddressAmountUpdatePendingAmount("ckb", order.FromAddress, amount)
						if err != nil {
							log.Errorf("Ckb解冻失败，更新pending amount error：%v", err)
						} else {
							log.Infof("ckb解冻成功，解冻地址为：%s，解冻金额为： %s", order.FromAddress, amount.String())
						}
					}
				}
			}
		}
		if strings.ToLower(applyOrder.CoinName) == "btm" {
			orders, err := api.WalletOrderService.GetColdOrder(info.OuterOrderNo)
			if err == nil && len(orders) > 0 {
				order := orders[0]
				createData := []byte(order.CreateData)
				if len(createData) > 0 {
					var cd model.BtmCreateData
					err = json.Unmarshal(createData, &cd)
					if err == nil {
						var amount = decimal.NewFromInt(0)
						for _, in := range cd.Sources {
							inAmount := decimal.NewFromInt(in.Amount)
							amount = amount.Add(inAmount)
							err = dao.FcTransPushUnFreezeUtxo("btm", in.OutputId, in.SourcePos, in.SourceId)
							if err != nil {
								log.Errorf("btm解冻失败，更新fc_trans_push is_spent error：%v", err)
							}
						}
						//转换为float类型
						amount = amount.Shift(-8)
						log.Infof("btm需要解冻的金额为： %s", amount.String())
						err = dao.FcAddressAmountUpdatePendingAmount("btm", order.FromAddress, amount)
						if err != nil {
							log.Errorf("btm解冻失败，更新pending amount error：%v", err)
						} else {
							log.Infof("btm解冻成功，解冻地址为：%s，解冻金额为： %s", order.FromAddress, amount.String())
						}
					}
				}
			}
		}
		//解冻热钱包
		if in_array(strings.ToLower(applyOrder.CoinName), []interface{}{"stx"}) {
			res, err := dao.FcOrderHotFindByOutNo(info.OuterOrderNo)
			if err == nil && len(res) == 1 {
				order := res[0]
				sendAmount := decimal.NewFromInt(order.Amount)
				//获取精度
				coin, errC := dao.FcCoinSetGetCoinId(strings.ToLower(applyOrder.CoinName), "")
				if errC == nil {
					amount := sendAmount.Shift(-int32(coin.Decimal))
					//查询 pending_amount
					err = dao.FcAddressAmountUpdatePendingAmount(strings.ToLower(applyOrder.CoinName), order.FromAddress, amount)
					if err == nil {
						log.Infof("%s 解冻成功，解冻地址为：%s，解冻金额为： %s", applyOrder.CoinName, order.FromAddress, amount.String())
					} else {
						log.Infof("%s 解冻失败，解冻地址为：%s，解冻金额为： %s，Err= %v", applyOrder.CoinName, order.FromAddress, amount.String(), err)
					}
				} else {
					log.Errorf("%s 解冻，查询coin set error，Err=%v", applyOrder.CoinName, errC)
				}
			} else {
				log.Errorf("%s find order hot error,err=%v,nums=%d", applyOrder.CoinName, err, len(res))
			}
		}
	}

	//入账处理
	log.Infof("入账状态：%d", info.Status)
	if info.Status == status.BroadcastStatus.Int() {
		log.Info("入账准备")
		//查询成功记录
		order, err := api.WalletOrderService.GetSuccessColdOrder(info.OuterOrderNo)
		if err != nil {
			log.Errorf("入账失败，查询订单异常：%s", info.OuterOrderNo)
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("入账失败，查询订单：%s,异常：%s", info.OuterOrderNo, err.Error()))
			httpresp.HttpRespCodeErrOnly(ctx)
			return
		}
		if order.Status != status.BroadcastStatus.Int() {
			log.Errorf("入账失败，查询订单状态异常：%s", info.OuterOrderNo)
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("入账失败，订单：%s，状态异常：%d", info.OuterOrderNo, order.Status))
			httpresp.HttpRespCodeErrOnly(ctx)
			return
		}

		//修改成功，通知回调
		log.Info("修改成功")
		err = api.OrderService.SendApplyTransferSuccess(applyOrder.Id)
		if err != nil {
			log.Errorf("入账失败，订单：%s", info.OuterOrderNo)
			//钉钉通知异常
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("入账失败，订单：%s", info.OuterOrderNo))
			httpresp.HttpRespCodeErrOnly(ctx)
			return
		}
		dao.UpdatePriorityCompletedIfExist(applyOrder.Id)
		//通知商户
		api.OrderService.NotifyToMch(applyOrder)
	} else {
		successOrder, _ := api.WalletOrderService.GetSuccessColdOrder(info.OuterOrderNo)
		if successOrder != nil {
			//避免sql语句错误，再判断一次
			if successOrder.Status == status.BroadcastStatus.Int() {
				//日志打印
				log.Infof("订单：%s,已经广播成功，不允许重复推送异常状态：%d,请ai检查是否需要修复", info.OuterOrderNo, info.Status)
				//钉钉发送异常
				dingding.ErrTransferDingBot.NotifyStr(
					fmt.Sprintf("订单：%s,已经广播成功，不允许重复推送异常状态: %d,请ai检查是否需要修复",
						info.OuterOrderNo, info.Status))
				httpresp.HttpRespCodeErrOnly(ctx)
				return
			}
		}

		coldOrder, err := api.WalletOrderService.GetColdOrderByOrderId(info.OrderNo)
		errStr := info.Message
		if err != nil {
			//errStr = err.Error()
			//log.Errorf("查新订单异常,但是会继续向下执行,orderNo：%s",info.OrderNo)
			////不能中断。报警处理。
			//dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("查新订单异常,orderNo：%s",info.OrderNo))
		}
		if coldOrder != nil {
			errStr = coldOrder.ErrorMsg
		}
		if info.Status == status.BroadcastErrorStatus.Int() || info.Status == status.UnknowErrorStatus.Int() {
			api.OrderService.SendApplyFail(applyOrder.Id)
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("交易失败失败，订单：%s，状态：%d，错误：%s", info.OuterOrderNo, info.Status, errStr))
			httpresp.HttpRespOkOnly(ctx)
			return
		} else if info.Status == status.CreateErrorStatus.Int() ||
			info.Status == status.SignErrorStatus.Int() ||
			info.Status == status.PendingTimeoutStatus.Int() {
			log.Infof("交易失败，订单：%s", info.OuterOrderNo)
			err = api.OrderService.SendApplyFail(applyOrder.Id)
			if err != nil {
				dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单=[%s],币种=[%s],出账失败，error:%s", applyOrder.OutOrderid, applyOrder.CoinName, err.Error()))
			} else {
				dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单=[%s],币种=[%s],出账失败，检查是否可以重推", applyOrder.OutOrderid, applyOrder.CoinName))
			}

			//if applyOrder.ErrorNum >= 3 {
			//	//直接修改为失败
			//	api.OrderService.SendApplyFail(applyOrder.Id)
			//	dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单=[%s],币种=[%s],交易重试上限,出账失败", applyOrder.OutOrderid, applyOrder.CoinName))
			//} else {
			//	//重推
			//	//修改状态为失败重试状态
			//	api.OrderService.SendApplyRetry(applyOrder.Id)
			//
			//}

		}
	}
	httpresp.HttpRespCodeOkOnly(ctx)
}

func in_array(need interface{}, needArr []interface{}) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}
