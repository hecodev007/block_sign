package v3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/assist"
	dingModel "github.com/group-coldwallet/blockchains-go/model/dingding"
	"github.com/group-coldwallet/blockchains-go/model/merge"
	"github.com/group-coldwallet/blockchains-go/model/recycle"
	"github.com/group-coldwallet/blockchains-go/model/repush"
	"github.com/group-coldwallet/blockchains-go/model/reset"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/token"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/runtime/job"
	"github.com/group-coldwallet/blockchains-go/service"
	recycle2 "github.com/group-coldwallet/blockchains-go/service/recycle"
	transfer2 "github.com/group-coldwallet/blockchains-go/service/transfer"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

var (
	rePushMtx sync.RWMutex
	tm        *job.TxManager
)

// hoo的token

// 钉钉的Outgoing机制
// fix 短时间内发送会有状态覆盖问题，后续解决,并且逻辑越来越长，需要分离
func DingOutgoing(ctx *gin.Context, txMgr *job.TxManager) {
	var (
		err error
		//token      string                  // 机器人token
		contextReq *dingModel.DdingStruct  // 钉钉发送过来的数据结构
		content    string                  // 接收的内容
		dingRole   *dingModel.DingRoleAuth // 用户权限
	)
	tm = txMgr
	contextReq = new(dingModel.DdingStruct)
	dingRole = new(dingModel.DingRoleAuth)

	if dingding.ReviewDingBot == nil {
		log.Errorf("初始化钉钉机器人失败")
		return
	}
	log.Infof("ip:%s", ctx.ClientIP())
	log.Infof("ding head:%s", ctx.GetHeader("token"))
	//token = ctx.GetHeader("token")

	// 验证token
	//if conf.Cfg.IMBot.SecretToken != token {
	//	httpresp.HttpRespCodeOkOnly(ctx)
	//	return
	//}
	// 解析发送过来的内容
	data, _ := ctx.GetRawData()
	log.Infof("ding body:%s", string(data))
	json.Unmarshal(data, contextReq)

	if contextReq == nil {
		httpresp.HttpRespErrorOnly(ctx)
		return
	}
	if contextReq.Msgtype == "text" {
		content = strings.TrimSpace(contextReq.Text.Content)
	}
	if content == "" {
		httpresp.HttpRespErrorOnly(ctx)
		return
	}
	log.Infof("接收内容为：%s", contextReq.Text.Content)

	// 检查是否存在用户
	if _, ok := dingding.DingUsers[contextReq.SenderId]; !ok {
		contextReq.SenderId = "guest"
		// 用户组没有这个用户
		//httpresp.HttpRespErrorOnly(ctx)
		//return
	}

	// 获取用户角色
	dingRoleKey := dingding.DingUsers[contextReq.SenderId].RoleName

	// 判断用户是否拥有权限
	dingRole = dingding.DingRoles[dingRoleKey]
	if dingRole == nil {
		dingding.ReviewDingBot.NotifyStr("缺少设置权限")
		return
	}
	log.Infof("%s 拥有权限：%v", dingRoleKey, dingRole)
	hasRole, dingCommand := dingRole.HaveCommand(content)
	if !hasRole {
		dingding.ReviewDingBot.NotifyStr("命令正在开发中...")
		return
	}

	specifiedMessage := ""
	switch dingCommand {
	case dingModel.DING_REPLACE_FAILURE_TXS:
		err = replaceFailureTxs(content)
	case dingModel.DING_CANCEL_TXS:
		err = cancelTx(content)
	case dingModel.DING_FORCE_CANCEL_TXS:
		err = forceCancelTx(content)
	case dingModel.DING_OUT_COLLECT:
		err = outCollect(content)
	case dingModel.DING_VERIFY:
		err = verifyOrder(content, dingRole)
	case dingModel.DING_REPUSH_ORDER:
		err = rePushOrder(content, dingRole)
	case dingModel.DING_CANCEL_ORDER:
		err = cancelOrder(content, dingRole)
	case dingModel.DING_CHECK_ORDER:
		err = checkOrder(content, dingRole)
	case dingModel.DING_ROLLBACK_ORDER:
		err = rollbackOrder(content, dingRole)
	//case dingModel.DING_ABANDONED_ORDER:
	//	err = abandonedOrder(content, dingRole)
	//	if err == nil {
	//		dingding.ReviewDingBot.NotifyStr("已执行废弃需求")
	//	}
	case dingModel.DING_DISCARD_REPUSH_ORDER:
		err = DiscardAndRePush(content, dingRole)
	case dingModel.DING_DISCARD_ROLLBACK_ORDER:
		err = DiscardAndRollback(content, dingRole)
	case dingModel.DING_MERGE_ORDER:
		err = mergeColdAddress(content)
	case dingModel.DING_RECYCLE_COIN:
		err = recycleCoin(content)
	case dingModel.DING_RESET_ETH:
		err = resetEthTx(content)
	case dingModel.DING_ETH_GAS:
		err = resetEthGas(content)
	case dingModel.DING_ETH_CLOSE_ALLCOLLECT:
		err = resetEthAllCollect(content)
	case dingModel.DING_ETH_CLOSE_COLLECT:
		err = resetEthCloseCollect(content)
	case dingModel.DING_ETH_OPEN_COLLECT:
		err = resetEthOpenCollect(content)
	case dingModel.DING_DOT_RECYCLE:
		err = recycleDot(content)
	case dingModel.DING_DHX_RECYCLE:
		err = recycleDhx(content)
	case dingModel.DING_BTM_RECYCLE:
		//	归集btm
		//err = recycleBtm(content)
	case dingModel.DING_CKB_RECYCLE:
		//	归集ckb
		err = recycleCKB(content)
	// 2020-09-28 新添加四个命令
	case dingModel.DING_ETH_FEE:
		err = ethTransferFee(content)
	case dingModel.DING_ETH_LIST_AMOUNT:
		err = ethListAmount(content)
	case dingModel.DING_ETH_COLLECT_TOKEN:
		err = ethCollectToken(content)
	case dingModel.DING_ETH_INTERNAL:
		err = ethInternal(content)
	case dingModel.DING_COLLECT:
		err = collect(content)
	case dingModel.DING_ORDER_COLLECT:
		err = orderCollect(content)
	case dingModel.DING_ETH_RESET_NONCE:
		err = ethResetNonce(content)
	case dingModel.DING_CHAIN_REPUSH:
		err = RepushChainData(content, contextReq.SenderNick)
	case dingModel.DING_CHAIN_FORCE_REPUSH:
		err = forceRepushChainData(content, contextReq.SenderNick)
	case dingModel.DING_FAIL_CHAIN:
		err = rollBackChainData(content, dingRole)
	case dingModel.DING_FIX_BALANCE:
		err = fixBlance(content, dingRole)
	case dingModel.DING_DEL_KEY:
		err = delRediskey(content, dingRole)
	case dingModel.DING_REFRESH_KEY:
		err = refreshkey(content)
	case dingModel.DING_XRP_SUPPLEMENTAL:
		err = xrpSupplemental(content)
	case dingModel.DING_BTC_MERGE:
		err = btcMergeCoin(content)
		specifiedMessage = "请求已收到，To地址为：3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw，请到BTC浏览器查看合并进度"
	case dingModel.DING_BTC_RECYCLE:
		err = btcRecycleCoin(content)
		specifiedMessage = "请求已收到，To地址为：3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw，请到BTC浏览器查看回收进度"
	case dingModel.DING_MAIN_CHAIN_REFRESH:
		err = mainChainRefresh(content)

	case dingModel.DING_FIX_ADDR_AMOUNT_ALL0C:
		err = FixCollectAddressAsset(content)
	case dingModel.DING_SAME_WAY_BACK:
		err = setSameWayBack(content)
	case dingModel.DING_ORDER_TX_LINK:
		err = OrderTxLink(content)
	case dingModel.DING_SET_PRIORITY_ORDER:
		err = SetPriorityOrder(content)
	case dingModel.DING_CANCEL_PRIORITY_ORDER:
		err = CancelPriorityOrder(content)
	// case dingModel.DING_COIN_FEE:
	//	err = transferFee(content)
	// case dingModel.DING_COIN_COLLECT_TOKEN:
	//	err = collectToken(content)
	// case dingModel.DING_FIND_ADDRESS_FEE:
	//	err = findAddressFee(content)
	default:
		httpresp.HttpRespErrorOnly(ctx)
		return
	}
	if err != nil {
		dingding.ReviewDingBot.NotifyStr(err.Error())
		httpresp.HttpRespErrorOnly(ctx)
		return
	}
	if specifiedMessage != "" {
		httpresp.HttpRespCodeOkOnlyWithMsg(ctx, specifiedMessage)
	} else {
		httpresp.HttpRespCodeOkOnly(ctx)
	}
}

func outCollect(content string) error {
	var (
		jsonDataStr string
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_OUT_COLLECT.ToString(), "", -1)
	params := new(merge.OutCollectParams)
	err := json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("钉钉出账归集: %+v", params)
	redisHelper, _ := util.AllocRedisClient()
	defer redisHelper.Close()
	tag := ""
	if params.Status == 1 {
		tag = "开启"
		redisHelper.Set("outcollect", "1", time.Hour*24*365*10)
	} else {
		tag = "关闭"
		redisHelper.Del("outcollect")
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("出账归集已%s", tag))
	return nil
}

// 审核订单
func verifyOrder(content string, auth *dingModel.DingRoleAuth) error {
	outOrderId := strings.Replace(content, dingModel.DING_VERIFY.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	applyOrder, err := api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，失败，没有该订单", outOrderId))
		return fmt.Errorf("审核订单：%s，失败，没有该订单, 错误信息 :%s", outOrderId, err.Error())
	}

	if !auth.HaveCoin(applyOrder.CoinName) {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("目前不支持该币种：%s", applyOrder.CoinName))
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}

	if applyOrder.Status != int(entity.ApplyStatus_Auditing) {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核失败，订单：%s，非待审核状态", outOrderId))
		return fmt.Errorf("审核失败，订单：%s，非待审核状态", outOrderId)
	}

	pushMsg := ""
	if applyOrder.TxType == entity.MultiAddrTx {
		orderId := fmt.Sprintf("%s_%d", applyOrder.OrderId, time.Now().Unix())
		if err = dao.FcTransfersApplyUpdateOrderIdAndStatus(applyOrder.Id, int(entity.ApplyStatus_Creating), orderId, 2); err != nil {
			return fmt.Errorf("审核订单:%s 失败 %v", applyOrder.OutOrderid, err)
		}
		if err = order.PushToWaitingList(applyOrder.OutOrderid); err != nil {
			pushMsg = fmt.Sprintf("，将订单推入待处理列表失败 %v", err)
		}
	} else {
		err = api.OrderService.SendApplyReviewOk(applyOrder.Id)
		if err != nil {
			// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，异常:%s", outOrderId, err.Error()))
			return fmt.Errorf("审核订单：%s，异常:%s", outOrderId, err.Error())
		}
	}

	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，成功 %s", outOrderId, pushMsg))

	fourMin := int64(60 * 4)
	if time.Now().Unix() >= applyOrder.Createtime+fourMin {
		log.Infof("订单审核已超时 %s", outOrderId)
		record, err := dao.FcReportRecordGetByOuterOrderNo(outOrderId)
		if err != nil {
			log.Infof("订单审核已超时 FcReportRecordGetByOuterOrderNo(%s) 失败 %v", outOrderId, err)
			return nil
		}
		reason := fmt.Sprintf("大额二次审核时长:%d", time.Now().Unix()-applyOrder.Createtime)
		if record == nil {
			rr := &entity.FcReportRecord{
				Chain:        applyOrder.CoinName,
				CoinCode:     applyOrder.CoinName,
				TxId:         "",
				ReportType:   entity.AuditTimeout,
				OuterOrderId: outOrderId,
				Remark:       reason,
				CreateTime:   time.Now().Unix(),
			}
			rr.Insert()
		} else {
			dao.FcReportRecordUpdate(record.Id, reason)
		}
	}

	return nil
}
func rePushOrder(content string, auth *dingModel.DingRoleAuth) error {
	outOrderId := strings.Replace(content, dingModel.DING_REPUSH_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	if err := rePushTryLock(outOrderId); err != nil {
		return err
	}
	return rePushOrderCore(outOrderId, auth)
}

func rePushTryLock(outOrderId string) error {
	defer rePushMtx.Unlock()
	rePushMtx.Lock()
	cache, err := redis.Client.Get(redis.GetRePushCacheKey(outOrderId))
	if err != nil {
		return err
	}
	if cache != "" {
		return fmt.Errorf("订单/交易 %s 重推操作正在执行，请稍后重试", outOrderId)
	}
	if err = redis.Client.Set(redis.GetRePushCacheKey(outOrderId), outOrderId, time.Minute*5); err != nil {
		return err
	}
	log.Infof("订单/交易 %s redis锁定成功", outOrderId)
	return nil
}

func callWalletServer(url string, data interface{}) error {
	log.Infof("准备调用walletServer %s，请求参数=%v", url, data)
	resp, err := util.PostJsonByAuth(url, conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, data)
	if err != nil {
		return fmt.Errorf("调用walletServer出错: %v", err)
	}

	log.Infof("%s 调用walletServer返回数据 %s", url, string(resp))
	result := transfer.DecodeWalletServerRespOrder(resp)
	if !result.Success() {
		return fmt.Errorf("call walletServer failure:%s", result.Message)
	}
	return nil
}

func rePushForMultiFromCore(url string, tx *entity.FcOrderTxs) error {
	req := struct {
		SeqNo string `json:"seqNo"`
	}{SeqNo: tx.SeqNo}
	if err := callWalletServer(url, req); err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推失败：%s, %s", tx.SeqNo, err.Error()))
		return err
	} else {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推成功：%s", tx.SeqNo))
		return nil
	}
}

func cancelTxForMultiFromCore(url string, seqNo string) {
	req := struct {
		SeqNo string `json:"seqNo"`
	}{SeqNo: seqNo}
	if err := callWalletServer(url, req); err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消交易失败：%s, %s", seqNo, err.Error()))
	} else {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("强制取消交易成功：%s", seqNo))
	}
}

func rePushForceForMultiFrom(tx *entity.FcOrderTxs) error {
	callUrl := fmt.Sprintf("%s/v2/forcerepush", conf.Cfg.Walletserver.Url)
	return rePushForMultiFromCore(callUrl, tx)
}

func rePushForMultiFromOrderTx(tx *entity.FcOrderTxs) error {
	callUrl := fmt.Sprintf("%s/v2/repush", conf.Cfg.Walletserver.Url)
	return rePushForMultiFromCore(callUrl, tx)
}

func rePushForMultiFrom(applyOrder *entity.FcTransfersApply) error {
	// 只有43状态下才能重推
	if applyOrder.Status != int(entity.ApplyStatus_Creating) {
		return fmt.Errorf("多地址出账订单，当前状态%d，不可重推", applyOrder.Status)
	}
	listRange, err := redis.Client.ListRange(redis.CacheKeyWaitingOrderList, 0, 10000)
	if err != nil {
		return err
	}
	for _, o := range listRange {
		if o == applyOrder.OutOrderid {
			return fmt.Errorf("订单%s在待处理列表已存在", applyOrder.OutOrderid)
		}
	}

	// 多地址出账
	// 生成订单时失败，直接重推
	dao.FcTransfersApplyUpdateRemark(int64(applyOrder.Id), " ")
	if err = order.PushToWaitingList(applyOrder.OutOrderid); err != nil {
		return fmt.Errorf("，将订单推入待处理列表失败 %v", err)
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推成功：%s", applyOrder.OutOrderid))
	return nil
}

func chainFailureForMultiFrom(tx *entity.FcOrderTxs) {
	req := struct {
		SeqNoList []string `json:"seqNoList"`
	}{SeqNoList: []string{tx.SeqNo}}
	callUrl := fmt.Sprintf("%s/v2/chainfailure", conf.Cfg.Walletserver.Url)
	if err := callWalletServer(callUrl, req); err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("设置为链上失败出错：%s, %s", tx.SeqNo, err.Error()))
	} else {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("设置为链上失败操作已完成：%s", tx.SeqNo))
	}
}

// 重推订单
func rePushOrderCore(outOrderId string, auth *dingModel.DingRoleAuth) error {
	defer func() {
		redis.Client.Del(redis.GetRePushCacheKey(outOrderId))
	}()
	log.Infof("准备处理重推订单 %s", outOrderId)
	var (
		err        error
		applyOrder *entity.FcTransfersApply
	)
	outOrderId = strings.TrimSpace(outOrderId)
	// 查询订单信息
	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		tx, errTx := dao.GetOrderTxBySeqNo(outOrderId)
		if errTx != nil {
			return fmt.Errorf("订单：%s,查询订单交易信息异常,error=[%s]", outOrderId, errTx.Error())
		}
		log.Infof("查询到 tx 数据 %v", tx)
		if tx != nil {
			rePushForMultiFromOrderTx(tx)
			return nil
		}
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}

	if applyOrder.TxType == entity.MultiAddrTx {
		txs, err := dao.FindOrderTxsByOuterOrderNo(outOrderId)
		if err != nil {
			return fmt.Errorf("FindOrderTxsByOuterOrderNo 订单:%s 出错 %v", outOrderId, err)
		}
		//isSuc := true
		if len(txs) == 0 {
			// 表示 applyOrder 还没有生成订单
			rePushForMultiFrom(applyOrder)
		} else {
			// 43和49都可重推
			if applyOrder.Status != int(entity.ApplyStatus_Creating) && applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
				return fmt.Errorf("多地址出账订单，当前状态%d，不可重推", applyOrder.Status)
			}
			// 把订单下面可以重推的交易全部重推
			for _, tx := range txs {
				if tx.CanRePush() {
					rePushForMultiFromOrderTx(&tx)
				}
			}
		}
		return nil
	}

	// 查询是否允许重推
	err = api.OrderService.IsAllowRepush(outOrderId)
	if err != nil {
		return err
	}

	// 设置重新出账,重试一次，设置错误次数为3即可
	err = api.OrderService.SendApplyRetryOnce(applyOrder.Id, global.RetryNum-1)
	if err != nil {
		return fmt.Errorf("重推订单：%s，异常:%s", outOrderId, err.Error())
	}
	log.Infof("%s 执行SendApplyRetryOnce完成", outOrderId)
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，异常:%s", outOrderId, err.Error()))
		return err
	}
	defer redisHelper.Close()
	interceptKey := fmt.Sprintf("%d_%s", applyOrder.AppId, applyOrder.OutOrderid)
	// 清除rediskey
	_ = redisHelper.Del(interceptKey)

	cltMsg := ""
	if order.NewCollectEnable() && order.IsNewCollectVersion(applyOrder.CoinName) {
		if err = tryCollect(applyOrder); err != nil {
			cltMsg = "，出账归集失败:" + err.Error()
		}
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重推订单：%s，成功%s", outOrderId, cltMsg))
	return nil
}

// 取消订单
func cancelOrder(content string, auth *dingModel.DingRoleAuth) error {
	var (
		err        error
		outOrderId string
		applyOrder *entity.FcTransfersApply
	)
	// 后缀跟的是订单号
	outOrderId = strings.Replace(content, dingModel.DING_CANCEL_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}
	if applyOrder.Status != int(entity.ApplyStatus_Auditing) || applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		return fmt.Errorf("取消失败，订单：%s，非待审核状态", outOrderId)
	}
	err = api.OrderService.SendAuditFail(applyOrder.Id)
	if err != nil {
		return fmt.Errorf("取消订单：%s，异常:%s", outOrderId, err.Error())
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("取消订单：%s，成功", outOrderId))
	return nil
}

func checkOrderForMultiFrom(applyOrder *entity.FcTransfersApply) error {
	txs, err := dao.FindOrderTxsByOuterOrderNo(applyOrder.OutOrderid)
	if err != nil {
		return err
	}
	orderSta := 0
	orderTotalAmount := ""
	walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)
	if walletType == status.WalletType_Cold {
		coldOrders, err := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
		if err != nil {
			return err
		}
		if len(coldOrders) == 0 {
			if applyOrder.Remark != "" {
				return fmt.Errorf("订单apply状态:%d %s", applyOrder.Status, applyOrder.Remark)
			}

			return fmt.Errorf("订单apply状态:%d (还没有生成fc_order)，请尝试重推", applyOrder.Status)
		}
		if len(coldOrders) > 1 {
			return fmt.Errorf("订单不唯一")
		}
		orderSta = coldOrders[0].Status
		orderTotalAmount = coldOrders[0].TotalAmount
	} else {
		return fmt.Errorf("orderHot 暂不支持多地址出账")
	}

	sucMsg := ""
	if orderSta == status.BroadcastStatus.Int() {
		sucMsg = "（订单已完成）"
	}

	orderAmtDecimal, _ := decimal.NewFromString(orderTotalAmount)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("订单：%s，金额：%s，Apply状态：%d, 订单状态：%d %s\n", applyOrder.OutOrderid, orderAmtDecimal, applyOrder.Status, orderSta, sucMsg))
	for _, tx := range txs {
		txAmtDecimal, _ := decimal.NewFromString(tx.Amount)
		sb.WriteString("----------------------------------------------------------------------------------\n")
		sb.WriteString(fmt.Sprintf("交易：%s\n", tx.SeqNo))
		sb.WriteString(fmt.Sprintf("出账地址：%s\n", tx.FromAddress))
		sb.WriteString(fmt.Sprintf("金额：%s\n", txAmtDecimal))
		sb.WriteString(fmt.Sprintf("状态：%d（%s）\n", tx.Status, entity.OtxStatusMap[tx.Status]))
		if !tx.IsCompleted() && tx.ErrMsg != "" && tx.ErrMsg != "REPUSH" {
			sb.WriteString(fmt.Sprintf("错误信息：%s\n", tx.ErrMsg))
		}
		sb.WriteString("----------------------------------------------------------------------------------\n")
	}
	dingding.ReviewDingBot.NotifyStr(sb.String())
	return nil
}

// 检查订单
func checkOrder(content string, auth *dingModel.DingRoleAuth) error {
	var (
		err        error
		outOrderId string
		applyOrder *entity.FcTransfersApply
	)
	// 后缀跟的是订单号
	outOrderId = strings.Replace(content, dingModel.DING_CHECK_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}

	if applyOrder.TxType == entity.MultiAddrTx {
		// 多地址出账版本
		return checkOrderForMultiFrom(applyOrder)
	}

	// 检查币种类型
	walletType := global.WalletType(applyOrder.CoinName, applyOrder.AppId)
	switch walletType {
	case status.WalletType_Cold:
		coldOrders, err := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
		if err != nil {
			return fmt.Errorf("c:查询签名订单异常 订单：%s，币种名：%s，err:%s", outOrderId, applyOrder.CoinName, err.Error())
		}
		if len(coldOrders) == 0 {
			return fmt.Errorf("订单：%s，币种名：%s,冷钱包无相关记录，注意检查是否需要重发，apply状态：%d",
				outOrderId, applyOrder.CoinName, applyOrder.Status)
		}
		for _, v := range coldOrders {
			if v.Status < status.BroadcastStatus.Int() {
				return fmt.Errorf("订单：%s，币种名：%s,订单正在执行，apply状态：%d",
					outOrderId, applyOrder.CoinName, applyOrder.Status)
			}
		}
		// 时间倒序，那么取第一条判断即可
		if coldOrders[0].Status == status.BroadcastStatus.Int() {
			if applyOrder.Status != int(entity.ApplyStatus_TransferOk) {
				log.Errorf(fmt.Sprintf("自动修复：订单：%s，币种名：%s,订单已经完成，apply状态异常：%d",
					outOrderId, applyOrder.CoinName, applyOrder.Status))
				api.OrderService.SendApplyTransferSuccess(applyOrder.Id)

				// 从新下发回调
				api.OrderService.NotifyToMchByOutOrderId(applyOrder.OutOrderid)

				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("订单：%s，币种名：%s,订单自动修复完成",
						outOrderId, applyOrder.CoinName))
			} else {
				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("订单：%s，币种名：%s,订单已经完成",
						outOrderId, applyOrder.CoinName))
			}
			return nil
		} else if coldOrders[0].Status > status.BroadcastStatus.Int() {
			return fmt.Errorf("订单：%s，币种名：%s,签名端异常，状态：%d,错误内容：%s",
				outOrderId, applyOrder.CoinName, coldOrders[0].Status, coldOrders[0].ErrorMsg)
		}

	case status.WalletType_Hot:
		hotOrders, err := api.WalletOrderService.GetHotOrder(applyOrder.OutOrderid)
		if err != nil {

			return fmt.Errorf(fmt.Sprintf("h:查询签名订单异常 订单：%s，币种名：%s，err:%s",
				outOrderId, applyOrder.CoinName, err.Error()))
		}
		if len(hotOrders) == 0 {
			return fmt.Errorf("订单：%s，币种名：%s,热钱包无相关记录，注意检查是否需要重发,apply状态：%d",
				outOrderId, applyOrder.CoinName, applyOrder.Status)
		}
		for _, v := range hotOrders {
			if v.Status < status.BroadcastStatus.Int() {
				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("订单：%s，币种名：%s,订单正在执行，apply状态：%d",
						outOrderId, applyOrder.CoinName, applyOrder.Status))
				return nil
			}
		}
		// 时间倒序，那么取第一条判断即可
		if hotOrders[0].Status == status.BroadcastStatus.Int() {
			if applyOrder.Status != int(entity.ApplyStatus_TransferOk) {
				log.Errorf(fmt.Sprintf("自动修复：订单：%s，币种名：%s,订单已经完成，apply状态异常：%d",
					outOrderId, applyOrder.CoinName, applyOrder.Status))
				api.OrderService.SendApplyTransferSuccess(applyOrder.Id)

				// 从新下发回调
				api.OrderService.NotifyToMchByOutOrderId(applyOrder.OutOrderid)

				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("订单：%s，币种名：%s,订单自动修复完成",
						outOrderId, applyOrder.CoinName))
			} else {
				dingding.ReviewDingBot.NotifyStr(
					fmt.Sprintf("订单：%s，币种名：%s,订单已经完成",
						outOrderId, applyOrder.CoinName))
			}
			return nil
		} else if hotOrders[0].Status > status.BroadcastStatus.Int() {
			return fmt.Errorf("订单：%s，币种名：%s,签名端异常，状态：%d,错误内容：%s",
				outOrderId, applyOrder.CoinName, hotOrders[0].Status, hotOrders[0].ErrorMsg)
		}
	default:
		return fmt.Errorf("未识别的币种类型，订单：%s，币种名：%s",
			outOrderId, applyOrder.CoinName)
	}
	return nil
}

//
//func abandonedOrder(content string, auth *dingModel.DingRoleAuth) error {
//	var (
//		err        error
//		outOrderId string
//		applyOrder *entity.FcTransfersApply
//	)
//	// 后缀跟的是订单号
//	outOrderId = strings.Replace(content, dingModel.DING_ABANDONED_ORDER.ToString(), "", -1)
//	outOrderId = strings.TrimSpace(outOrderId)
//	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
//	if err != nil {
//		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
//	}
//	if !auth.HaveCoin(applyOrder.CoinName) {
//		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
//	}
//	if applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
//		return fmt.Errorf("订单：%s，apply 状态不允许废弃", outOrderId)
//	}
//
//	// 检查币种类型
//	walletType := global.WalletType[applyOrder.CoinName]
//	switch walletType {
//	case status.WalletType_Cold:
//		orders, _ := api.WalletOrderService.GetColdOrder(applyOrder.OutOrderid)
//		if len(orders) == 0 {
//			return fmt.Errorf("订单：%s，无相关执行记录", outOrderId)
//		}
//
//		for _, v := range orders {
//			if v.Status == status.BroadcastStatus.Int() {
//				return fmt.Errorf("订单：%s，已经广播，不允许废弃", outOrderId)
//			}
//		}
//
//		var orderId string
//		for _, v := range orders {
//			if v.Status == status.BroadcastErrorStatus.Int() {
//				orderId = v.OrderNo
//				break
//			}
//		}
//		if orderId == "" {
//			return fmt.Errorf("订单：%s，不存在异常，不允许废弃", outOrderId)
//		}
//		return api.WalletOrderService.UpdateColdOrderState(orderId, status.AbandonedTransaction.Int())
//	case status.WalletType_Hot:
//		orders, _ := api.WalletOrderService.GetHotOrder(applyOrder.OutOrderid)
//		if len(orders) == 0 {
//			return fmt.Errorf("订单：%s，无相关执行记录", outOrderId)
//		}
//
//		for _, v := range orders {
//			if v.Status == status.BroadcastStatus.Int() {
//				return fmt.Errorf("订单：%s，已经广播，不允许废弃", outOrderId)
//			}
//		}
//
//		var orderId string
//		for _, v := range orders {
//			if v.Status == status.BroadcastErrorStatus.Int() {
//				orderId = v.OrderNo
//				break
//			}
//		}
//		if orderId == "" {
//			return fmt.Errorf("订单：%s，不存在异常，不允许废弃", outOrderId)
//		}
//		return api.WalletOrderService.UpdateHotOrderState(orderId, status.AbandonedTransaction.Int())
//	default:
//		return errors.New("error wallettype")
//	}
//	return nil
//
//}

// 合并冷地址地址金额，应付多地址出账时的问题，目前针对账户模型
func mergeColdAddress(content string) error {
	var (
		jsonDataStr string
		msg         string
		err         error
		has         bool // 是否允许执行
	)
	// 目前支持币种
	coins := []string{"btm", "satcoin", "eac", "bsc", "heco", "bnb", "cds", "ar", "ksm", "bnc", "hnt", "crab", "vet", "celo", "mtr", "fio", "dot", "azero", "sgb-sgb", "kar", "dhx", "nodle"}

	jsonDataStr = strings.Replace(content, dingModel.DING_MERGE_ORDER.ToString(), "", -1)
	params := new(merge.MergeParams)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)

	for _, coin := range coins {
		if strings.ToLower(params.Coin) == coin {
			has = true
			break
		}
	}
	if !has {
		return errors.New("暂不允许该币种合并")
	}
	if _, ok := api.MergeService[strings.ToLower(params.Coin)]; !ok {
		return errors.New("缺少币种初始化配置")
	}
	msg, err = api.MergeService[strings.ToLower(params.Coin)].MergeCoin(*params)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(msg)
	return nil

}

func rollbackOrder(content string, auth *dingModel.DingRoleAuth) error {
	var (
		err        error
		outOrderId string
		applyOrder *entity.FcTransfersApply
	)
	// 后缀跟的是订单号
	outOrderId = strings.Replace(content, dingModel.DING_ROLLBACK_ORDER.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)
	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}
	if applyOrder.Status != int(entity.ApplyStatus_Ignore) {
		return fmt.Errorf("订单：%s,已经回滚", outOrderId)
	}

	if applyOrder.Status != int(entity.ApplyStatus_Auditing) ||
		applyOrder.Status != int(entity.ApplyStatus_CreateFail) {
		return fmt.Errorf("订单：%s，apply 状态不允许回滚", outOrderId)
	}
	err = api.OrderService.SendApplyRollback(applyOrder.Id)
	if err != nil {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，异常:%s", outOrderId, err.Error()))
		return fmt.Errorf("审核订单：%s，异常:%s", outOrderId, err.Error())
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，成功", outOrderId))
	return nil

}

func btcMergeCoin(content string) error {
	var (
		jsonDataStr string
	)
	cacheKey := "btcmerge"
	v, err := redis.Client.Get(cacheKey)
	if err != nil {
		return err
	}
	if v != "" {
		return errors.New("上次一BTC合并操作尚未完成，不可频繁操作")
	}

	jsonDataStr = strings.Replace(content, dingModel.DING_BTC_MERGE.ToString(), "", -1)
	params := new(merge.BtcMergeParams)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}

	if params.AppId == 0 {
		return fmt.Errorf("参数异常：%s", jsonDataStr)
	}
	redis.Client.Set(cacheKey, "btc", time.Minute*20)
	go func() {
		job.BtcMergeProcess(int64(params.AppId))
		redis.Client.Del(cacheKey)
	}()
	return nil
}

func btcRecycleCoin(content string) error {
	var (
		jsonDataStr string
	)
	cacheKey := "btcrecycle"
	v, err := redis.Client.Get(cacheKey)
	if err != nil {
		return err
	}
	if v != "" {
		return errors.New("上次一BTC零散回收操作尚未完成，不可频繁操作")
	}

	jsonDataStr = strings.Replace(content, dingModel.DING_BTC_RECYCLE.ToString(), "", -1)
	params := new(recycle.BtcRecycleParams)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}

	if params.AppId == 0 {
		return fmt.Errorf("参数异常：%s", jsonDataStr)
	}
	redis.Client.Set(cacheKey, "btc", time.Minute*20)
	go func() {
		job.BtcCollectProcess(int64(params.AppId))
		redis.Client.Del(cacheKey)
	}()
	return nil
}

// 合并冷地址地址金额，应付多地址出账时的问题，目前针对账户模型
func recycleCoin(content string) error {
	var (
		jsonDataStr string
		msg         string
		err         error
		has         bool // 是否允许执行
		feeFloat    decimal.Decimal
	)
	// 目前支持币种
	coins := []string{"bsv", "ltc", "zec", "dcr", "ghost", "bch", "hc", "doge",
		"avax", "dash", "biw", "atom", "oneo", "bcha", "xec", "ada", "zen", "satcoin", "eac", "btm"}

	jsonDataStr = strings.Replace(content, dingModel.DING_RECYCLE_COIN.ToString(), "", -1)
	params := new(recycle.RecycleParams)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}

	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)

	if params.Coin == "" || params.AppId == 0 || params.Model > 1 || params.Model < 0 {
		return fmt.Errorf("参数异常：%s", string(jsonDataStr))
	}

	feeFloat, err = decimal.NewFromString(params.FeeFloat)
	if err != nil {
		return fmt.Errorf("fee参数异常：%s", err.Error())
	}

	for _, coin := range coins {
		if strings.ToLower(params.Coin) == coin {
			has = true
		}
	}
	if !has {
		return errors.New("暂不允许该币种零散回收")
	}
	if _, ok := api.RecycleService[strings.ToLower(params.Coin)]; !ok {
		return errors.New("缺少币种初始化配置")
	}

	mchinfo, err := api.MchService.GetMchName(params.AppId)
	if err != nil {
		return errors.New("缺少商户信息")
	}

	// 获取币种的冷地址
	var toAddress string
	toAddrs, err := service.GetColdAddre(params.AppId, params.Coin)
	if err != nil || len(toAddrs) == 0 {
		return fmt.Errorf("无法获取币种%s冷地址", params.Coin)
	}
	toAddress = toAddrs[0].Address

	paramsOrder := model.TransferParams{
		Sfrom:           mchinfo.Platform,
		CallBack:        "",
		OutOrderId:      fmt.Sprintf("%s-utxo-merge-%d-%d", params.Coin, params.Model, time.Now().Unix()),
		CoinName:        strings.ToLower(params.Coin),
		Amount:          decimal.Decimal{},
		ToAddress:       toAddress,
		TokenName:       "",
		ContractAddress: "",
		Memo:            "",
		Fee:             decimal.Decimal{},
		IsForce:         false,
	}

	coinSet := global.CoinDecimal[strings.ToLower(params.Coin)]
	if coinSet == nil {
		return fmt.Errorf("合并，DB缺少币种[%s]设置", params.Coin)
	}

	ta := &entity.FcTransfersApply{
		Username:   "api",
		OrderId:    util.GetUUID(),
		Applicant:  paramsOrder.Sfrom,
		AppId:      mchinfo.Id,
		CallBack:   "",
		OutOrderid: paramsOrder.OutOrderId,
		CoinName:   params.Coin,
		Type:       "hb",
		Status:     int(entity.ApplyStatus_Merge),
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     toAddress,
		AddressFlag: "to",
		ToAmount:    "0",
		Lastmodify:  util.GetChinaTimeNow(),
	}

	orderId, err := dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo})
	if err != nil {
		return fmt.Errorf("%s币种零散归集，创建订单异常", params.Coin)
	}

	reqHead := &transfer.OrderRequestHead{
		ApplyId:      orderId,
		ApplyCoinId:  int64(coinSet.Id),
		OuterOrderNo: paramsOrder.OutOrderId,
		OrderNo:      util.GetUUID(),
		MchId:        int64(mchinfo.Id),
		MchName:      mchinfo.Platform,
		CoinName:     params.Coin,
		Worker:       "",
	}
	// neo币种暂时特殊处理
	if params.FromAddress != "" && strings.ToLower(params.Coin) == "oneo" {
		// 查询from地址是否属于这个商户
		addrData, err := dao.FcGenerateAddressGet(params.FromAddress)
		if err != nil || addrData.PlatformId != params.AppId {
			return fmt.Errorf("无法在商户下面找到该地址信息，：%s", params.FromAddress)
		}
		reqHead.RecycleAddress = params.FromAddress
	}

	msg, err = api.RecycleService[strings.ToLower(params.Coin)].RecycleCoin(reqHead, toAddress, feeFloat, params.Model)
	if err != nil {
		return err
	} else {
		dingding.ReviewDingBot.NotifyStr(msg + ",地址:" + toAddress)
	}
	return nil

}

/*
func: reset eth tx
auth: flynn
date: 2020-07-03
*/
func resetEthTx(content string) error {
	var (
		jsonDataStr string
		err         error
		order       *entity.FcOrder
		apply       *entity.FcTransfersApply
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_RESET_ETH.ToString(), "", -1)
	params := new(reset.ResetEthReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.OrderId == "" || params.Txid == "" {
		return fmt.Errorf("参数异常：%s", string(jsonDataStr))
	}
	//
	order, err = dao.FcOrderGetByOutOrderNoAndTxid(params.OrderId, params.Txid, int(status.WaitResetTransaction))
	if err != nil || order == nil {
		order, err = dao.FcOrderGetByOutOrderNoAndTxid(params.OrderId, params.Txid, int(status.BroadcastStatus))
		if err != nil {
			return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", params.OrderId, err.Error())
		}
		// 将 status设置为12
		err = dao.FcOrderUpdateState2(params.OrderId, params.Txid, int(status.WaitResetTransaction))
		if err != nil {
			return fmt.Errorf("订单：%s，更新,error=[%s]", params.OrderId, err.Error())
		}
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s：%s，第一阶段成功", dingModel.DING_RESET_ETH.ToString(), params.OrderId))
		return nil
	} else {
		// 查询apply表，更新状态： 30--> 49
		apply, err = dao.FcTransfersApplyByOutOrderNo(params.OrderId)
		if err != nil {
			return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", params.OrderId, err.Error())
		}
		if apply.Status != 30 {
			return fmt.Errorf("订单：%s,订单状态错误，订单状态=[%s]", params.OrderId, apply.Status)
		}
		err = dao.FcTransfersApplyUpdateByOutNOAddErr(params.OrderId, 49)
		if err != nil {
			return fmt.Errorf("订单：%s，更新,error=[%s]", params.OrderId, err.Error())
		}
		// 更新order表状态为10
		err = dao.FcOrderUpdateState2(params.OrderId, params.Txid, int(status.AbandonedTransaction))
		if err != nil {
			return fmt.Errorf("订单：%s，更新,error=[%s]", params.OrderId, err.Error())
		}
		dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s：%s，第二阶段成功", dingModel.DING_RESET_ETH.ToString(), params.OrderId))
		return nil
	}
}

var (
	dotSrv *recycle2.DotRecycleService
	dhxSrv *recycle2.DhxRecycleService
	//btmSrv *recycle2.BtmRecycleService
	ckbSrv *recycle2.CkbRecycleService
)

func recycleDot(content string) error {
	var (
		jsonDataStr string
		err         error
		has         bool
	)

	jsonDataStr = strings.Replace(content, dingModel.DING_DOT_RECYCLE.ToString(), "", -1)
	params := new(recycle.RecycleDotReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.AppId == 0 {
		return errors.New("do not set app id")
	}
	// 目前支持币种
	coins := []string{"dot", "azero", "sgb-sgb", "kar", "nodle"}
	var coinName string
	for _, c := range coins {
		if strings.ToLower(params.Coin) == c {
			has = true
			coinName = c
			break
		}
	}
	if !has {
		return errors.New("暂不允许该币种零散回收")
	}
	if params.Num == 0 {
		params.Num = 10 //	默认回收10笔
	}
	mchinfo, err := api.MchService.GetMchName(params.AppId)
	if err != nil {
		return errors.New("缺少商户信息")
	}
	var to string
	if params.Address != "" {
		to = params.Address
	} else {
		// 获取币种的冷地址
		toAddrs, err := service.GetColdAddre(params.AppId, coinName)
		if err != nil || len(toAddrs) == 0 {
			return fmt.Errorf("无法获取币种%s冷地址", coinName)
		}
		to = toAddrs[rand.Intn(len(toAddrs))].Address
	}
	if dotSrv == nil {
		dotSrv = recycle2.NewDotRecycleService()
	}
	err = dotSrv.RecycleCoin(mchinfo, to, params.Num)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 回收成功，回收到冷地址 %s", dingModel.DING_DOT_RECYCLE.ToString(), to))
	return nil
}

func recycleDhx(content string) error {
	var (
		jsonDataStr string
		err         error
		has         bool
	)

	jsonDataStr = strings.Replace(content, dingModel.DING_DHX_RECYCLE.ToString(), "", -1)
	params := new(recycle.RecycleDotReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.AppId == 0 {
		return errors.New("do not set app id")
	}
	// 目前支持币种
	coins := []string{"dhx"}
	var coinName string
	for _, c := range coins {
		if strings.ToLower(params.Coin) == c {
			has = true
			coinName = c
			break
		}
	}
	if !has {
		return errors.New("暂不允许该币种零散回收")
	}
	if params.Num == 0 {
		params.Num = 10 //	默认回收10笔
	}
	mchinfo, err := api.MchService.GetMchName(params.AppId)
	if err != nil {
		return errors.New("缺少商户信息")
	}
	var to string
	if params.Address != "" {
		to = params.Address
	} else {
		// 获取币种的冷地址
		toAddrs, err := service.GetColdAddre(params.AppId, coinName)
		if err != nil || len(toAddrs) == 0 {
			return fmt.Errorf("无法获取币种%s冷地址", coinName)
		}
		to = toAddrs[rand.Intn(len(toAddrs))].Address
	}
	if dhxSrv == nil {
		dhxSrv = recycle2.NewDhxRecycleService()
	}
	err = dhxSrv.RecycleCoin(mchinfo, to, params.Num)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 回收成功，回收到冷地址 %s", dingModel.DING_DHX_RECYCLE.ToString(), to))
	return nil
}

func mainChainRefresh(content string) error {
	var (
		jsonDataStr string
		err         error
	)

	jsonDataStr = strings.Replace(content, dingModel.DING_MAIN_CHAIN_REFRESH.ToString(), "", -1)
	params := new(token.NewTokenReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)

	support := map[string]string{}
	support["trx"] = ""
	support["bsc"] = ""
	support["heco"] = ""
	support["eos"] = ""
	support["okt"] = ""
	support["bnb"] = ""
	support["hsc"] = ""
	support["eth"] = ""

	coinName := strings.ToLower(params.Coin)
	if _, ok := support[coinName]; !ok {
		return errors.New("不支持的链" + coinName)
	}

	needExecCmd := coinName != "eth"

	wg := sync.WaitGroup{}
	wg.Add(1)
	if needExecCmd {
		wg.Add(1)
	}
	go func() {
		// 刷新内存
		runtime.InitGlobalReload()
		wg.Done()
	}()

	if needExecCmd {
		go func() {
			cmd := exec.Command("/bin/sh", "-c", "ansible "+coinName+" -m command -a 'supervisorctl restart "+coinName+"sync'")
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			log.Info(cmd.Args)
			err := cmd.Run()
			if err != nil {
				log.Infof("主链刷新 Command finished with error %s %s\n", err.Error(), stderr.String())
				log.Info(stderr.String())
			}
			log.Info(out.String())
			wg.Done()
		}()
	}
	wg.Wait()
	dingding.ReviewDingBot.NotifyStr("主链刷新成功")
	return nil
}

// 重置gas
func resetEthGas(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_GAS.ToString(), "", -1)
	params := new(reset.ResetEthGasReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	// http转发
	url := conf.Cfg.Walletserver.Url + "/admin/admin/replace/eth"
	resp, err := util.PostJsonByAuth(url, "rylink", "hoo123!@#", params)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重置gas成功，resp：%s", string(resp)))
	return nil
}

// 关闭所有归集
func resetEthAllCollect(content string) error {
	var (
		err error
		pid int
	)
	if content != dingModel.DING_ETH_CLOSE_ALLCOLLECT.ToString() {
		return fmt.Errorf("不支持的命令：%s", content)
	}
	//
	pid = global.CoinDecimal["eth"].Id
	if pid == 0 {
		return errors.New("查询ethid失败")
	}
	err = api.CoinService.CloseAllCollect(pid)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr("关闭eth全部归集成功")
	return nil
}

// 关闭单个归集
func resetEthCloseCollect(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_CLOSE_COLLECT.ToString(), "", -1)
	params := new(reset.ResetEthCollectReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	err = api.CoinService.CloseCollect(params.Coin)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("关闭eth--%s归集成功", params.Coin))
	return nil
}

// 开启单个归集
func resetEthOpenCollect(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_OPEN_COLLECT.ToString(), "", -1)
	params := new(reset.ResetEthCollectReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)

	coinSet, _ := dao.FcCoinSetGetByName(params.Coin, 1)
	if coinSet != nil {
		log.Infof("开启归集 获取到的coinSet为：%+v", coinSet)
	}
	st, _ := decimal.NewFromString(coinSet.StaThreshold)
	fs, _ := decimal.NewFromString(params.Bof)
	if fs.Cmp(st) == -1 {
		return fmt.Errorf("币种：%s bof不能低于%s", params.Coin, st)
	}

	err = api.CoinService.OpenCollect(params.Coin, params.Bof)
	if err != nil {
		return err
	}

	redis.Client.Del(fmt.Sprintf("clt_center_coin_collect_conf_%s", strings.ToLower(params.Coin)))
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("开启eth--%s归集成功", params.Coin))

	return nil
}

/*
归集BTM
//*/
//func recycleBtm(content string) error {
//	var (
//		jsonDataStr string
//		err         error
//	)
//	jsonDataStr = strings.Replace(content, dingModel.DING_BTM_RECYCLE.ToString(), "", -1)
//	params := new(recycle.RecycleBtmReq)
//	err = json.Unmarshal([]byte(jsonDataStr), params)
//	if err != nil {
//		return err
//	}
//	log.Infof("jsonDataStr:%s", jsonDataStr)
//	log.Infof("jsonDataStr:%+v", params)
//	if params.AppId == 0 {
//		return errors.New("do not set app id")
//	}
//	mchinfo, err := api.MchService.GetMchName(params.AppId)
//	if err != nil {
//		return errors.New("缺少商户信息")
//	}
//	if btmSrv == nil {
//		btmSrv = recycle2.NewBtmRecycleService()
//	}
//	coin, err := btmSrv.RecycleCoin(mchinfo, to, params.Num)
//	if err != nil {
//		return err
//	}
//	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 回收成功!!!", dingModel.DING_BTM_RECYCLE.ToString()))
//	return nil
//}

/*
date: 2020-09-28
author: flynn
func: add 4 ding commond
*/

func ethTransferFee(content string) error {
	var (
		jsonDataStr string
		err         error
		mch         *entity.FcMch
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_FEE.ToString(), "", -1)
	params := new(transfer.EthTransferFee)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.MchId == 0 {
		return errors.New("缺少MchId")
	}
	if params.To == "" {
		return errors.New("缺少to地址")
	}
	mch, err = api.MchService.GetMchName(int(params.MchId))
	if err != nil || mch == nil {
		return fmt.Errorf("get mch info error,%v", err)
	}
	mchName := mch.Platform
	err = transfer2.EthTransferFee(params.MchId, params.To, mchName, params.FeeFloat)
	if err != nil {
		return err
	}
	// 通知成功
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 打手续费成功!!!", dingModel.DING_ETH_FEE.ToString()))
	return nil
}

func ethListAmount(content string) error {
	var (
		jsonDataStr string
		err         error
		mch         *entity.FcMch
		aa          []*entity.FcAddressAmount
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_LIST_AMOUNT.ToString(), "", -1)
	params := new(transfer.EthListAmount)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.MchId == 0 {
		return errors.New("缺少MchId")
	}
	if params.Coin == "" {
		return errors.New("缺少币种coin")
	}
	mch, err = api.MchService.GetMchName(int(params.MchId))
	if err != nil || mch == nil {
		return fmt.Errorf("get mch info error,%v", err)
	}
	limit := 5
	if params.Num > 0 {
		limit = params.Num
	}

	coinSets, err := entity.FcCoinSet{}.Find(builder.Eq{"name": params.Coin})
	coinSet := coinSets[0]
	if err != nil {
		log.Errorf("列举金额 从fcCoinSet获取数据失败:%s", err.Error())
		return errors.New("从coinSet获取数据失败")
	}

	// 查找amount
	aa, err = entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"coin_type": strings.ToLower(params.Coin), "app_id": params.MchId}.
		And(builder.In("type", []int{1, 2})), limit)

	if err != nil {
		return fmt.Errorf("find address amount error,%v", err)
	}
	if len(aa) == 0 {
		return errors.New("do not find amount")
	}

	var respStr string
	var collectStr string
	var collectType string
	respStr = "出账地址,非特殊情况不合并 \n 一般使用用户地址归集 \n"

	if 1 == coinSet.IsCollect {
		collectType = "是"
	} else if 0 == coinSet.IsCollect {
		collectType = "否"
	} else {
		collectType = "未知"
	}

	collectStr = "是否开启归集：" + collectType + "\n"
	collectStr = collectStr + "bof：" + coinSet.CollectThreshold + "\n"
	respStr = respStr + collectStr + "\n"

	for _, a := range aa {
		adType := "未知"
		switch a.Type {
		case 1:
			adType = "出账"
		case 2:
			adType = "用户"
		}
		respStr = respStr + fmt.Sprintf("%s地址：%s,金额：%s\n", adType, a.Address, a.Amount)
	}
	// 通知成功
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 列举amount成功\n%s", dingModel.DING_ETH_LIST_AMOUNT.ToString(), respStr))
	return nil
}

func orderCollect(content string) error {
	outerOrderNo := strings.Replace(content, dingModel.DING_ORDER_COLLECT.ToString(), "", -1)
	outerOrderNo = strings.TrimSpace(outerOrderNo)

	if err := rePushTryLock(outerOrderNo); err != nil {
		return err
	}
	if err := orderCollectProcess(outerOrderNo); err != nil {
		return err
	}
	return nil
}

func collect(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_COLLECT.ToString(), "", -1)
	params := new(transfer.Collect)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}

	chain := params.Coin
	//币种信息
	coinSet, err := dao.FcCoinSetGetByName(params.Coin, 1)
	if err != nil {
		return err
	}
	if coinSet.Pid > 0 {
		parent, err := dao.FcCoinSetGetCoinInfo(coinSet.Pid)
		if err != nil {
			return err
		}
		chain = parent.Name
	}

	if !order.IsNewCollectVersion(chain) {
		return errors.New("暂不支持此链")
	}

	list, err := dao.FcFindAddressAmountUserAddrList(params.MchId, params.Coin, params.From)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return errors.New("没有找到符合归集条件的地址，请检查地址和币种是否正确")
	}

	threshold, _ := decimal.NewFromString(coinSet.CollectThreshold)

	var sb strings.Builder
	var pickAddrs []string
	for _, addr := range list {
		amt, _ := decimal.NewFromString(addr.Amount)
		if amt.Cmp(threshold) >= 0 {
			pickAddrs = append(pickAddrs, addr.Address)
		} else {
			sb.WriteString(fmt.Sprintf("%s:未达到归集阈值\n", addr.Address))
		}
	}
	if len(pickAddrs) == 0 {
		return errors.New("没有找到符合归集条件的地址\n" + sb.String())
	}

	orderNo := fmt.Sprintf("INTERNAL_%s_%d", params.Coin, time.Now().Nanosecond())
	if err = order.CallAdminCollectCenter(orderNo, pickAddrs, coinSet.Token, params.Coin, threshold); err != nil {
		return err
	}
	msg := fmt.Sprintf("已发起归集请求\n符合条件的地址:%v", pickAddrs)
	if sb.String() != "" {
		msg += "\n忽略的地址\n" + sb.String()
	}
	dingding.ReviewDingBot.NotifyStr(msg)
	return nil
}

func ethCollectToken(content string) error {
	var (
		jsonDataStr string
		err         error
		mch         *entity.FcMch
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_COLLECT_TOKEN.ToString(), "", -1)
	params := new(transfer.EthCollectToken)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	// todo 应急代码，暂时硬代码写死
	for _, v := range params.From {
		if strings.ToLower(v) == "0x0000bf7f4e4b7fb2315fc6d5d0f8854c91dff1d8" {
			return fmt.Errorf("禁止操作手续费地址归集")
		}
		//if strings.ToLower(v) == "0x0093e5f2a850268c0ca3093c7ea53731296487eb" {
		//	return fmt.Errorf("禁止操作[出账地址]归集")
		//}
	}
	if params.Coin == "" || params.MchId == 0 || len(params.From) == 0 {
		return errors.New("参数错误")
	}
	mch, err = api.MchService.GetMchName(int(params.MchId))
	if err != nil || mch == nil {
		return fmt.Errorf("get mch info error,%v", err)
	}
	err = transfer2.EthCollectToken(params.Coin, mch, params.From)
	if err != nil {
		return err
	}
	// 通知成功
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s:代币%s归集成功", dingModel.DING_ETH_COLLECT_TOKEN.ToString(), params.Coin))
	return nil
}

func ethInternal(content string) error {
	var (
		jsonDataStr string
		err         error
		mch         *entity.FcMch
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_INTERNAL.ToString(), "", -1)
	params := new(transfer.EthInternal)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)

	if params.From == "" || params.MchId == 0 || params.To == "" || params.Amount == "" {
		return errors.New("参数错误")
	}
	mch, err = api.MchService.GetMchName(int(params.MchId))
	if err != nil || mch == nil {
		return fmt.Errorf("get mch info error,%v", err)
	}
	err = transfer2.EthInternal(mch, params.Amount, params.From, params.To)
	if err != nil {
		return err
	}
	// 通知成功
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s 请求成功", dingModel.DING_ETH_INTERNAL.ToString()))
	return nil
}

func ethResetNonce(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_ETH_RESET_NONCE.ToString(), "", -1)
	params := new(transfer.EthResetNonce)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.Address == "" {
		return fmt.Errorf("address is null")
	}
	// http转发
	url := conf.Cfg.Walletserver.Url + "/admin/admin/eth/resetnonce"
	resp, err := util.PostJsonByAuth(url, "rylink", "hoo123!@#", params)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("重置nonce成功，resp：%s", string(resp)))
	return nil
}

func forceRepushChainData(content, creator string) error {
	log.Infof("执行强制补数据 %s", content)
	content = strings.Replace(content, "强制", "", -1)
	jsonDataStr := strings.Replace(content, dingModel.DING_CHAIN_REPUSH.ToString(), "", -1)
	params := new(repush.DingRepush)
	err := json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	if params.Txid == "" {
		return errors.New("txid is require")
	}

	redisHelper, err := util.AllocRedisClient()
	defer redisHelper.Close()

	key := "force_repush_" + strings.ToLower(params.Txid)
	err = redisHelper.Set(key, "forceRePush")
	if err != nil {
		log.Errorf("执行强制补数据 setToCache redis Set error %v", err)
		return err
	}
	err = redisHelper.Expire(key, 300)
	if err != nil {
		log.Errorf("执行强制补数据 setToCache redis Set error %v", err)
		return err
	}

	return RepushChainData(content, creator)
}

// 补推数据
func RepushChainData(content string, creator string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_CHAIN_REPUSH.ToString(), "", -1)
	params := new(repush.DingRepush)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)

	if util.IsInArrayStr(params.Coin, []string{
		"dot", "azero", "sgb-sgb", "kar", "ksm", "bnc", "crab", "kda", "pcx", "ori", "fis", "dom", "star",
		"sol", "nyzo", "ar", "cocos", "fil", "wd", "bnb", "dhx", "flow", "nodle"}) {
		if params.Height == 0 {
			return fmt.Errorf("币种%s,需要添加高度补数据", params.Coin)
		}
	}

	if strings.ToLower(params.Txid) == "0xa09ed51f1bb73df12b908fcd1dd6f8741fa8d95b338e7e0f1e3da723900ec7aa" {
		return errors.New("该交易已经退币，不能补数据，否则会给用户多入账")
	}

	// http转发
	url := conf.Cfg.DataServer
	resp, err := util.PostJson(url, params)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("resp：%s", string(resp)))

	go func(req *repush.DingRepush) {
		existRepair, err := dao.FcGetRepairRecordByTxId(req.Txid)
		if err != nil {
			log.Infof("调用插入FcRepairRecord失败: %v", err)
			return
		}
		if existRepair == nil {
			rr := entity.FcRepairRecord{
				Chain:      req.Coin,
				TxId:       req.Txid,
				Height:     int64(req.Height),
				Creator:    creator,
				CreateTime: time.Now().Unix(),
			}
			if err = rr.Insert(); err != nil {
				log.Infof("插入FcRepairRecord失败: %v", err)
			}
			log.Infof("插入FcRepairRecord 完成")
		}

	}(params)

	return nil
}

/*
归集CKB
*/
func recycleCKB(content string) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_CKB_RECYCLE.ToString(), "", -1)
	params := new(recycle.RecycleBtmReq)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	log.Infof("jsonDataStr:%+v", params)
	if params.AppId == 0 {
		return errors.New("do not set app id")
	}
	mchinfo, err := api.MchService.GetMchName(params.AppId)
	if err != nil {
		return errors.New("缺少商户信息")
	}
	if ckbSrv == nil {
		ckbSrv = recycle2.NewCkbRecycleService()
	}
	err = ckbSrv.RecycleCoin(mchinfo)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 回收成功!!!", dingModel.DING_CKB_RECYCLE.ToString()))
	return nil
}

// 链上失败回滚
func rollBackChainData(content string, auth *dingModel.DingRoleAuth) error {
	var (
		err        error
		outOrderId string
		applyOrder *entity.FcTransfersApply
	)

	// 后缀跟的是订单号
	outOrderId = strings.Replace(content, dingModel.DING_FAIL_CHAIN.ToString(), "", -1)
	outOrderId = strings.TrimSpace(outOrderId)

	log.Infof("链上失败回滚 %v", outOrderId)
	tx, err := dao.GetOrderTxBySeqNo(outOrderId)
	if err != nil {
		return err
	}
	log.Infof("获取到的tx %v", tx)

	if tx != nil {
		// 多地址出账
		chainFailureForMultiFrom(tx)
		return nil
	}

	applyOrder, err = api.OrderService.GetApplyOrder(outOrderId)
	if err != nil {
		return fmt.Errorf("订单：%s,查询订单信息异常,error=[%s]", outOrderId, err.Error())
	}
	if !auth.HaveCoin(applyOrder.CoinName) {
		return fmt.Errorf("目前不支持该币种：%s", applyOrder.CoinName)
	}
	// if applyOrder.Status != int(entity.ApplyStatus_Ignore) {
	//	return fmt.Errorf("订单：%s,已经回滚", outOrderId)
	// }
	//
	// if applyOrder.Status != int(entity.ApplyStatus_Auditing) {
	//	return fmt.Errorf("订单：%s，apply 状态不允许回滚", outOrderId)
	// }
	err = api.OrderService.AbandonedOrder(applyOrder.OutOrderid)
	if err != nil {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，异常:%s", outOrderId, err.Error()))
		return fmt.Errorf("链上废弃订单：%s，异常:%s", outOrderId, err.Error())
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("链上废弃订单：%s，成功", outOrderId))
	return nil
}

// /*
// 支持多个币种的打手续费
// */
// func transferFee(content string)  error{
//	var (
//		jsonDataStr string
//		err         error
//		mch         *entity.FcMch
//	)
//	jsonDataStr = strings.Replace(content, dingModel.DING_COIN_FEE.ToString(), "", -1)
//	params := new(transfer.CoinTransferFee)
//	err = json.Unmarshal([]byte(jsonDataStr), params)
//	if err != nil {
//		return err
//	}
//	log.Infof("jsonDataStr:%s", jsonDataStr)
//	log.Infof("jsonDataStr:%+v", params)
//	// 判断是否支持 coin
//	if params.Coin=="" {
//		return errors.New("币种不能为空")
//	}
//	if !dingSrv.InDingSupportArray(params.Coin) {
//		return fmt.Errorf("不支持该币种 %s 的打手续费功能",params.Coin)
//	}
//	if params.MchId == 0 {
//		return errors.New("缺少MchId")
//	}
//	if params.To == "" {
//		return errors.New("缺少to地址")
//	}
//	if params.FeeFloat=="" {
//		return errors.New("缺少feeFloat地址")
//	}
//	mch, err = api.MchService.GetMchName(int(params.MchId))
//	if err != nil || mch == nil {
//		return fmt.Errorf("get mch info error,%v", err)
//	}
//	mchName := mch.Platform
//	err = dingSrv.CoinTransferFee(params.MchId,params.Coin,params.To,mchName,params.FeeFloat)
//	if err != nil {
//		return err
//	}
//	//通知成功
//	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 打手续费成功!!!", dingModel.DING_COIN_FEE.ToString()))
//	return nil
// }
//
// /*
// 支持多个币种的代币归集
// */
// func collectToken(content string) error{
//	var (
//		jsonDataStr string
//		err         error
//		mch         *entity.FcMch
//	)
//	jsonDataStr = strings.Replace(content, dingModel.DING_COIN_COLLECT_TOKEN.ToString(), "", -1)
//	params := new(transfer.CoinCollectToken)
//	err = json.Unmarshal([]byte(jsonDataStr), params)
//	if err != nil {
//		return err
//	}
//	log.Infof("jsonDataStr:%s", jsonDataStr)
//	log.Infof("jsonDataStr:%+v", params)
//	if params.Coin == "" || params.MchId == 0 || len(params.From) == 0 {
//		return errors.New("参数错误")
//	}
//	mch, err = api.MchService.GetMchName(int(params.MchId))
//	if err != nil || mch == nil {
//		return fmt.Errorf("get mch info error,%v", err)
//	}
//	err = dingSrv.CoinCollectToken(params.Coin,params.To,mch,params.From)
//	//err = transfer2.EthCollectToken(params.Coin, mch, params.From)
//	if err != nil {
//		return err
//	}
//	//通知成功
//	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s:代币%s归集成功", dingModel.DING_COIN_COLLECT_TOKEN.ToString(), params.Coin))
//	return nil
// }
//
// func findAddressFee(content string) error {
//
//	var (
//		jsonDataStr string
//		err         error
//		mch         *entity.FcMch
//	)
//	jsonDataStr = strings.Replace(content, dingModel.DING_FIND_ADDRESS_FEE.ToString(), "", -1)
//	params := new(transfer.CoinFindAddressFee)
//	err = json.Unmarshal([]byte(jsonDataStr), params)
//	if err != nil {
//		return err
//	}
//	log.Infof("jsonDataStr:%s", jsonDataStr)
//	log.Infof("jsonDataStr:%+v", params)
//	if params.Coin == "" {
//		return errors.New("coin不能为空")
//	}
//	if params.MchId == 0 {
//		return errors.New("mchId不能为空")
//	}
//	if params.Address=="" {
//		return errors.New("address不能为空")
//	}
//	var mainCoinName string
//	// 查找coin的配置
//	coins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1, "name": params.Coin})
//	if err != nil {
//		return fmt.Errorf("%s find coin set error,%v", params.Coin, err)
//	}
//	if len(coins) != 1 {
//		return fmt.Errorf("%s do not find coin set", params.Coin)
//	}
//	if coins[0].Pid==0 {
//		mainCoinName = params.Coin
//	}else{
//		//查找主链
//		mainCoin,err:=dao.FcCoinSetGetByStatus(coins[0].Pid,1)
//		if err != nil {
//			return fmt.Errorf("没有找到该币种%s的主链信息：%v",params.Coin,err)
//		}
//		mainCoinName = strings.ToLower(mainCoin.Name)
//	}
//	//判断是否支持这个币种
//	if !dingSrv.InDingSupportArray(mainCoinName){
//		return fmt.Errorf("不支持该币种%s的主链%s金额查询",params.Coin,mainCoinName)
//	}
//	mch, err = api.MchService.GetMchName(int(params.MchId))
//	if err != nil || mch == nil {
//		return fmt.Errorf("get mch info error,%v", err)
//	}
//	dbAmount,chainAmount,err:=dingSrv.CoinFindAddressFee(mainCoinName,params.Address,mch)
//	if err != nil {
//		return err
//	}
//	respStr:=fmt.Sprintf("数据库金额：%s\n链上金额：%s",dbAmount,chainAmount)
//	//通知成功
//	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("%s: 查找%s地址的手续费金额成功\n%s", dingModel.DING_FIND_ADDRESS_FEE.ToString(),params.Coin, respStr))
//	return nil
// }

// 修正余额
func fixBlance(content string, auth *dingModel.DingRoleAuth) error {
	var (
		jsonDataStr string
		err         error
	)
	jsonDataStr = strings.Replace(content, dingModel.DING_FIX_BALANCE.ToString(), "", -1)
	params := new(assist.FixBalanceParams)
	err = json.Unmarshal([]byte(jsonDataStr), params)
	if err != nil {
		return err
	}
	log.Infof("jsonDataStr:%s", jsonDataStr)
	// todo 从配置文件读取
	// url := conf.Cfg.FixServer

	coin := params.Coin
	if params.TokenName != "" {
		coin = params.TokenName
	}

	tips := ""
	coin = strings.ToLower(coin)
	addrAmount, _ := dao.FcAddressAmountFindByCoinNameAndAddress(coin, params.Address)
	if addrAmount != nil {
		if addrAmount.CoinId == 0 {
			log.Infof("纠正余额 %s %s 需要修复coinId", coin, params.Address)
			if coinSet, ok := global.CoinDecimal[coin]; ok {
				dao.FcAddressAmountUpdateCoinId(addrAmount.Id, coinSet.Id)
				tips = "提示：coinId已自动修复"
				log.Info("修复coinId完成")
			} else {
				log.Errorf("纠正余额 %s %s 需要修复coinId失败：global。CoinDecimal获取的数据为空", coin, params.Address)
			}
		}
	}

	url := ""
	if url == "" {
		url = "http://10.64.198.30:8880/repair/add_amount"
	}
	resp, err := util.PostJsonByAuth(url, "billService", "iNeed&904", params)
	if err != nil {
		return err
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("返回结果：%s\n%s", string(resp), tips))
	return nil
}

// xrpSupplemental 为XRP补充资金
func xrpSupplemental(content string) error {
	log.Infof("XRP资金补充入参 %s", content)
	jsonDataStr := strings.Replace(content, dingModel.DING_XRP_SUPPLEMENTAL.ToString(), "", -1)
	reqData := &transfer.XRPSupplemental{}
	err := json.Unmarshal([]byte(jsonDataStr), reqData)
	if err != nil {
		return err
	}

	resBytes, err := util.PostJson(conf.Cfg.Other.XRPSupplementalUrl, reqData)
	if err != nil {
		log.Errorf("XRP资金补充HTTP调用出错: %s", err.Error())
		return err
	}
	log.Infof("XRP资金补充HTTP调用返回数据 %s", string(resBytes))
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("返回结果：%s", string(resBytes)))
	return nil
}

func delRediskey(content string, auth *dingModel.DingRoleAuth) error {
	interceptKey := strings.Replace(content, dingModel.DING_DEL_KEY.ToString(), "", -1)
	interceptKey = strings.TrimSpace(interceptKey)
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return err
	}
	defer redisHelper.Close()
	// 清除rediskey
	err = redisHelper.Del(interceptKey)
	if err != nil {
		// dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("审核订单：%s，异常:%s", outOrderId, err.Error()))
		return fmt.Errorf("清除key，异常:%s", err.Error())
	}
	dingding.ReviewDingBot.NotifyStr("清除成功")
	return nil
}

// 加币刷新
func refreshkey(content string) error {
	if content == dingModel.DING_REFRESH_KEY.ToString() {
		// 刷新内存
		runtime.InitGlobalReload()
	}
	dingding.ReviewDingBot.NotifyStr("刷新成功")
	return nil
}

func tryCollect(apply *entity.FcTransfersApply) error {
	tacaList, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id, "address_flag": "to"})
	if err != nil {
		return fmt.Errorf("查询出账订单FcTransfersApplyCoinAddress错误： %v", err)
	}
	//一般出账地址只有一个
	if len(tacaList) != 1 {
		return fmt.Errorf("内部订单ID：%d,接受地址只允许一个", apply.Id)
	}

	taca := tacaList[0]

	coinName := apply.CoinName
	if apply.Eoskey != "" {
		coinName = apply.Eoskey
	}

	//币种信息
	coinSet, err := dao.FcCoinSetGetByName(coinName, 1)
	if err != nil {
		return err
	}
	amt, _ := decimal.NewFromString(taca.ToAmount)
	log.Infof("CheckIfNeedCollect %s 重推", apply.OutOrderid)
	_, err = order.CheckIfNeedCollect(apply.AppId, apply.OutOrderid, apply.CoinName, apply.Eoskey, apply.Eostoken, coinSet.CollectThreshold, amt)
	return err
}
