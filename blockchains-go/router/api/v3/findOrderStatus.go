package v3

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
)

const (
	UnRollback = iota //不可回滚		4,7,9
	Ready             //准备就绪		0
	OnGoing           //执行中			1,2,3
	Rollback          //可回滚	 		5,6,8,10,11
)

func FindOrderStatus(c *gin.Context) {
	//1. 获取外部订单参数
	clientId := c.PostForm("client_id")
	outOrderId := c.PostForm("out_order_id")
	log.Infof("client_id=%s out_order_id=%s", clientId, outOrderId)
	//2。 检查外部订单参数是否为空
	if outOrderId == "" || clientId == "" {
		log.Errorf("请求参数错误 error: %s", outOrderId)
		httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", nil)
		return
	}

	mch, err := dao.FcMchFindByApikey(clientId)
	if err != nil {
		log.Errorf("FcMchFindByApikey error: %s", err)
		httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", nil)
		return
	}

	appId := mch.Id
	log.Infof("FindOrderStatus outOrderId=%s appId=%d", outOrderId, appId)
	var result = make(map[string]int)
	result["queueState"] = UnRollback
	//3。 根据外部订单查询apply表为49的数据
	apply, err := dao.FcTransfersApplyByOutOrderNoAndApplyId(outOrderId, appId)
	log.Infof("FindOrderStatus %v err=%v", apply, err)
	if err != nil {
		if err.Error() == "Not Fount!" {
			result["queueState"] = Rollback
			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
			return
		}
		log.Errorf("查询apply表失败：%v", err)
		httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
		return
	}

	//if apply.Status == 49 || apply.Status == -41 {
	//	//4. 根据外部订单查询order或者order_hot表
	//	order, err := dao.FcOrderFindByOutNoAndMchName(outOrderId, mchName)
	//	if err != nil || len(order) == 0 {
	//		//处理order_hot表
	//		orderHot, err := dao.FcOrderHotFindByOutNoAndMchName(outOrderId, mchName)
	//		//if err != nil {
	//		//	log.Errorf("order或者order_hot表没有查询到该外部订单：%v", err)
	//		//	httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//		//	return
	//		//}
	//		//if len(orderHot) == 0 {
	//		//	log.Errorf("order_hot表查询到该外部订单数据为空：%s", outOrderId)
	//		//	httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//		//	return
	//		//}
	//
	//		if err != nil || len(orderHot) == 0 {
	//			//判断apply表时间是否超过15分钟，如果超过的话，就重滚
	//			ct := apply.Createtime
	//			if ct < 1000000000 {
	//				log.Errorf("apply createtime is less 1000000000：ct=%d", ct)
	//				httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//				return
	//			}
	//			timestamp := formatTime(ct)
	//			if time.Now().Unix() >= addTimeSecond(timestamp, 15) {
	//				//如果 大于15分钟
	//				//可以重推
	//				result["queueState"] = Rollback
	//				httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//				return
	//			} else {
	//				log.Errorf("apply表时间没有超过15分钟，不可以重推：%v", err)
	//				httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//				return
	//			}
	//		}
	//
	//		unRollback := false
	//		for _, oh := range orderHot {
	//			if oh.Status == 4 || oh.Status == 7 || oh.Status == 9 {
	//				unRollback = true
	//				break
	//			}
	//		}
	//		if unRollback {
	//			// 表示不可以回滚
	//			result["queueState"] = UnRollback
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		ongoing := false
	//		for _, oh := range orderHot {
	//			if oh.Status == 1 || oh.Status == 2 || oh.Status == 3 {
	//				ongoing = true
	//				break
	//			}
	//		}
	//		if ongoing {
	//			// 表示正在执行
	//			result["queueState"] = OnGoing
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		prepare := true
	//		for _, oh := range orderHot {
	//			if oh.Status != 0 {
	//				prepare = false
	//				break
	//			}
	//		}
	//		if prepare {
	//			// 表示准备就绪
	//			result["queueState"] = Ready
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		//最后表示可以回滚
	//		rollback := true
	//		unKnownStatus := -1
	//		for _, oh := range order {
	//			if oh.Status != 5 {
	//				if oh.Status != 6 {
	//					if oh.Status != 8 {
	//						if oh.Status != 11 {
	//							rollback = false //避免后续添加新的状态
	//							unKnownStatus = oh.Status
	//						}
	//					}
	//				}
	//			}
	//		}
	//		if rollback {
	//			result["queueState"] = Rollback
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		log.Errorf("未知order_hot状态： %d", unKnownStatus)
	//		httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//		return
	//	} else {
	//		// 1。 判断4，7，9状态
	//		unRollback := false
	//		for _, o := range order {
	//			if o.Status == 4 || o.Status == 7 || o.Status == 9 {
	//				unRollback = true
	//				break
	//			}
	//		}
	//		if unRollback {
	//			// 表示不可以回滚
	//			result["queueState"] = UnRollback
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		//2. 判断是否正在执行
	//		ongoing := false
	//		for _, o := range order {
	//			if o.Status == 1 || o.Status == 2 || o.Status == 3 {
	//				ongoing = true
	//				break
	//			}
	//		}
	//		if ongoing {
	//			// 表示不可以回滚
	//			result["queueState"] = OnGoing
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		//3。 判断是否准备就绪
	//		ready := true
	//		for _, oh := range order {
	//			if oh.Status != 0 {
	//				ready = false
	//				break
	//			}
	//		}
	//		if ready {
	//			// 表示准备就绪
	//			result["queueState"] = Ready
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//
	//		//最后 表示可以回滚
	//		rollback := true
	//		unKnownStatus := -1
	//		for _, oh := range order {
	//			if oh.Status != 5 {
	//				if oh.Status != 6 {
	//					if oh.Status != 8 {
	//						if oh.Status != 11 {
	//							rollback = false //避免后续添加新的状态
	//							unKnownStatus = oh.Status
	//						}
	//					}
	//				}
	//			}
	//		}
	//		if rollback {
	//			result["queueState"] = Rollback
	//			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	//			return
	//		}
	//		log.Errorf("未知order状态： %d", unKnownStatus)
	//		httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	//		return
	//	}
	//} else
	if apply.Status == 52 {
		// 钱包标记为可回滚
		result["queueState"] = Rollback
		httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
		return
	} else if apply.Status == 40 {
		log.Error("apply状态为审核状态可以回滚")
		result["queueState"] = Rollback
		httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
		return
	}
	log.Errorf("apply状态不合符回滚: %d", apply.Status)
	httpresp.HttpRespError(c, httpresp.FAIL, "不可回滚", result)
	return

}
