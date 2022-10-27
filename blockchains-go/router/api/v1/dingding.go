package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"strings"
)

//hoo的token
const HoosafeToken = "Rylink2020BlockChainsGo"

//const HooadminUsers_ = "$:LWCP_v1:$7T+wAtaQ0scPArIXzp39Ng=="

const adminUsert = "admin"
const guestUsert = "guest"

var coins = []string{"klay", "btc", "eth"}

//全部权限
var AdminUsers map[string]string = map[string]string{
	"$:LWCP_v1:$7T+wAtaQ0scPArIXzp39Ng==": guestUsert, //lijiayi
	"$:LWCP_v1:$r1zfb9PtSpuH/4dYeiHu1A==": adminUsert, //zhuwenjian
}

type DdingStruct struct {
	ChatbotUserId  string      `json:"chatbotUserId"`  //机器人id
	ConversationId string      `json:"conversationId"` //
	SenderId       string      `json:"senderId"`       //人员ID
	SenderNick     string      `json:"senderNick"`     //昵称
	IsAdmin        bool        `json:"isAdmin"`        //是否是管理员
	CreateAt       int64       `json:"createAt"`       //创建时间
	Msgtype        string      `json:"msgtype"`        //text
	Text           MsgtypeText //text类型内容
}
type MsgtypeText struct {
	Content string `json:"content"`
}

//钉钉的Outgoing机制
//fix 短时间内发送会有状态覆盖问题，后续解决,并且逻辑越来越长，需要分离
func DingOutgoing(ctx *gin.Context) {
	if dingding.ReviewDingBot == nil {
		log.Errorf("初始化钉钉机器人失败")
		return
	}
	log.Infof("ip:%s", ctx.ClientIP())
	log.Infof("ding head:%s", ctx.GetHeader("token"))
	token := ctx.GetHeader("token")
	if HoosafeToken != token {
		httpresp.HttpRespCodeOkOnly(ctx)
	}
	data, _ := ctx.GetRawData()
	log.Infof("ding body:%s", string(data))

	contextReq := new(DdingStruct)
	json.Unmarshal(data, contextReq)

	if _, ok := AdminUsers[contextReq.SenderId]; !ok {
		//用户组没有这个用户
		httpresp.HttpRespErrorOnly(ctx)
		return
	}

	userType := AdminUsers[contextReq.SenderId]

	if contextReq.Msgtype == "text" {
		contextReq.Text.Content = strings.TrimLeft(contextReq.Text.Content, " ")
		log.Infof("接收内容为：%s", contextReq.Text.Content)
		if strings.HasPrefix(contextReq.Text.Content, "重推") {
			if userType != adminUsert {
				dingding.ReviewDingBot.NotifyStr("您暂时没有权限")
				return
			}
			//后缀跟的是订单号
			outOrderId := strings.Replace(contextReq.Text.Content, "重推", "", -1)
			outOrderId = strings.TrimSpace(outOrderId)
			//查询订单信息
			applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
			if applyOrder == nil {
				if err != nil {
					if err.Error() != "Not Fount!" {
						log.Errorf("订单：%s,查询订单信息异常", err.Error())
					}
				}
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：'%s'，查询异常", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			has := false
			for _, v := range coins {
				if v == strings.ToLower(applyOrder.CoinName) {
					has = true
				}
			}
			if !has {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("目前不支持该币种：%s", applyOrder.CoinName))
				return
			}

			//查询是否允许重推
			err = api.OrderService.IsAllowRepush(outOrderId)
			if err != nil {
				dingding.ReviewDingBot.NotifyStr(err.Error())
				httpresp.HttpRespErrorOnly(ctx)
				return
			}
			//设置重新出账
			err = api.OrderService.SendApplyReviewOk(applyOrder.Id)
			if err != nil {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，异常:%s", outOrderId, err.Error()))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}
			redisHelper, err := util.AllocRedisClient()
			if err != nil {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，异常:%s", outOrderId, err.Error()))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}
			defer redisHelper.Close()
			interceptKey := fmt.Sprintf("%d_%s", applyOrder.AppId, applyOrder.OutOrderid)
			//清除rediskey
			_ = redisHelper.Del(interceptKey)
			dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，成功", outOrderId))
		} else if strings.HasPrefix(contextReq.Text.Content, "审核") {
			outOrderId := strings.Replace(contextReq.Text.Content, "审核", "", -1)
			outOrderId = strings.TrimSpace(outOrderId)
			applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
			if applyOrder == nil {
				if err != nil {
					if err.Error() != "Not Fount!" {
						log.Errorf("订单：%s,查询订单信息异常", err.Error())
					}
				}
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，失败，没有该订单", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			has := false
			for _, v := range coins {
				if v == strings.ToLower(applyOrder.CoinName) {
					has = true
				}
			}
			if !has {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("目前不支持该币种：%s", applyOrder.CoinName))
				return
			}

			if applyOrder.Status != int(entity.ApplyStatus_Auditing) {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核失败，订单：%s，非待审核状态", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			err = api.OrderService.SendApplyReviewOk(applyOrder.Id)
			if err != nil {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，异常:%s", outOrderId, err.Error()))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}
			dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，成功", outOrderId))
		} else if strings.HasPrefix(contextReq.Text.Content, "取消") {
			outOrderId := strings.Replace(contextReq.Text.Content, "取消", "", -1)
			outOrderId = strings.TrimSpace(outOrderId)
			applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
			if applyOrder == nil {
				if err != nil {
					if err.Error() != "Not Fount!" {
						log.Errorf("订单：%s,查询订单信息异常", err.Error())
					}
				}
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消订单：%s，失败，没有该订单", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			has := false
			for _, v := range coins {
				if v == strings.ToLower(applyOrder.CoinName) {
					has = true
				}
			}
			if !has {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("目前不支持该币种：%s", applyOrder.CoinName))
				return
			}

			if applyOrder.Status != int(entity.ApplyStatus_Auditing) {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消失败，订单：%s，非待审核状态", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			err = api.OrderService.SendAuditFail(applyOrder.Id)
			if err != nil {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消订单：%s，异常:%s", outOrderId, err.Error()))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}
			dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消订单：%s，成功", outOrderId))
		} else if strings.HasPrefix(contextReq.Text.Content, "检查") {
			outOrderId := strings.Replace(contextReq.Text.Content, "检查", "", -1)
			outOrderId = strings.TrimSpace(outOrderId)
			applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
			if applyOrder == nil {
				if err != nil {
					if err.Error() != "Not Fount!" {
						log.Errorf("订单：%s,查询订单信息异常", err.Error())
					}
				}
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("检查订单：%s，失败，没有该订单", outOrderId))
				httpresp.HttpRespErrorOnly(ctx)
				return
			}

			has := false
			for _, v := range coins {
				if v == strings.ToLower(applyOrder.CoinName) {
					has = true
				}
			}
			if !has {
				dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("目前不支持该币种：%s", applyOrder.CoinName))
				return
			}
			//检查币种类型
			walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)

			switch walletType {
			case status.WalletType_Cold:
				coldOrders, err := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
				if err != nil {
					dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("c:查询签名订单异常 订单：%s，币种名：%s，err:%s", outOrderId, applyOrder.CoinName, err.Error()))
					return
				}
				if len(coldOrders) == 0 {
					dingding.ReviewDingBot.NotifyStr(
						fmt.Sprintf("订单：%s，币种名：%s,冷钱包无相关记录，注意检查是否需要重发，apply状态：%d",
							outOrderId, applyOrder.CoinName, applyOrder.Status))
					return
				}
				for _, v := range coldOrders {
					if v.Status < status.BroadcastStatus.Int() {
						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单正在执行，apply状态：%d",
								outOrderId, applyOrder.CoinName, applyOrder.Status))
						return
					}
				}
				//时间倒序，那么取第一条判断即可
				if coldOrders[0].Status == status.BroadcastStatus.Int() {
					if applyOrder.Status != int(entity.ApplyStatus_TransferOk) {
						log.Errorf(fmt.Sprintf("自动修复：订单：%s，币种名：%s,订单已经完成，apply状态异常：%d",
							outOrderId, applyOrder.CoinName, applyOrder.Status))
						api.OrderService.SendApplyTransferSuccess(applyOrder.Id)

						//从新下发回调
						api.OrderService.NotifyToMchByOutOrderId(applyOrder.OutOrderid)

						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单自动修复完成",
								outOrderId, applyOrder.CoinName))
					} else {
						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单已经完成",
								outOrderId, applyOrder.CoinName))
					}
					return
				} else if coldOrders[0].Status > status.BroadcastStatus.Int() {
					dingding.ReviewDingBot.NotifyStr(
						fmt.Sprintf("订单：%s，币种名：%s,签名端异常，状态：%d,错误内容：%s",
							outOrderId, applyOrder.CoinName, coldOrders[0].Status, coldOrders[0].ErrorMsg))
					return
				}

			case status.WalletType_Hot:
				hotOrders, err := api.WalletOrderService.GetHotOrder(applyOrder.OutOrderid)
				if err != nil {
					dingding.ReviewDingBot.NotifyStr(
						fmt.Sprintf("h:查询签名订单异常 订单：%s，币种名：%s，err:%s",
							outOrderId, applyOrder.CoinName, err.Error()))
					return
				}
				if len(hotOrders) == 0 {
					dingding.ReviewDingBot.NotifyStr(
						fmt.Sprintf("订单：%s，币种名：%s,热钱包无相关记录，注意检查是否需要重发,apply状态：%d",
							outOrderId, applyOrder.CoinName, applyOrder.Status))
					return
				}
				for _, v := range hotOrders {
					if v.Status < status.BroadcastStatus.Int() {
						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单正在执行，apply状态：%d",
								outOrderId, applyOrder.CoinName, applyOrder.Status))
						return
					}
				}
				//时间倒序，那么取第一条判断即可
				if hotOrders[0].Status == status.BroadcastStatus.Int() {
					if applyOrder.Status != int(entity.ApplyStatus_TransferOk) {
						log.Errorf(fmt.Sprintf("自动修复：订单：%s，币种名：%s,订单已经完成，apply状态异常：%d",
							outOrderId, applyOrder.CoinName, applyOrder.Status))
						api.OrderService.SendApplyTransferSuccess(applyOrder.Id)

						//从新下发回调
						api.OrderService.NotifyToMchByOutOrderId(applyOrder.OutOrderid)

						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单自动修复完成",
								outOrderId, applyOrder.CoinName))
					} else {
						dingding.ReviewDingBot.NotifyStr(
							fmt.Sprintf("订单：%s，币种名：%s,订单已经完成",
								outOrderId, applyOrder.CoinName))
					}
					return
				} else if hotOrders[0].Status > status.BroadcastStatus.Int() {
					dingding.ReviewDingBot.NotifyStr(
						fmt.Sprintf("订单：%s，币种名：%s,签名端异常，状态：%d,错误内容：%s",
							outOrderId, applyOrder.CoinName, hotOrders[0].Status, hotOrders[0].ErrorMsg))
					return
				}

			default:
				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("未识别的币种类型，订单：%s，币种名：%s",
						outOrderId, applyOrder.CoinName))
				return
			}
		}

	}
	httpresp.HttpRespCodeOkOnly(ctx)
}
