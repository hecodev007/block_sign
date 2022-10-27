package job

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

//轮询数据库数据 进行交易
// Job Specific Functions
type TransferApplyBaseJob struct {
	CoinName    string
	LimitNum    int
	SleepSecond int64
}

//查找fc_transfers_apply表status字段=1的数据，构建数据结构发送到walletserver里面去(查找数据后，status字段变为7)

func forceColdSign(apply *entity.FcTransfersApply) (bool, error) {
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id, "address_flag": "to"})
	if err != nil {
		return false, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return false, errors.New("address more than 1")
	}
	return "0x13a72f3363776b64a7b03257dd55d41fe95ed962" == strings.ToLower(toAddrs[0].Address), nil
}

// 轮询数据库数据 进行交易
//redis 用于短暂延迟某个币种的出账，一个币种的并发容易造成大量的余额不足错误
func (e TransferApplyBaseJob) Run() {
	cacheKey := fmt.Sprintf("blockchain_busy_%s", e.CoinName)
	//TODO:本地运行错误，只是注释，未做其他处理
	//if busyCache != "" {
	//	log.Infof("%s 上一次的任务尚未完成，本次不继续执行", e.CoinName)
	//	return
	//}

	redis.Client.Set(cacheKey, e.CoinName, 100*time.Second)

	log.Infof("检索%s链订单", e.CoinName)
	var (
		err     error
		results []*entity.FcTransfersApply
	)

	nowTime := util.GetChinaTimeNowFormat()
	log.Infof("%s 准备从数据库查询待执行订单", e.CoinName)
	results, err = dao.FcTransfersApplyGroupValidOrderByName(e.CoinName, e.LimitNum)

	if err != nil {
		log.Errorf("查询订单数据异常:%s", err.Error())
		return
	}
	if len(results) > 0 {
		log.Infof("=======%s处理出账订单=======", nowTime)
		log.Infof("=======%s订单数量:%d=======", e.CoinName, len(results))
	}
	//数据交易
	for i, v := range results {
		// bsc和hsc有自己的blockchain，如果走热签
		// 他们在这里不需要处理
		if "bsc" == e.CoinName || "hsc" == e.CoinName {
			walletType := global.WalletType(e.CoinName, v.AppId)
			if walletType == status.WalletType_Hot {
				log.Infof("%s 商户:%d 订单 %s 为热签订单不处理.", e.CoinName, v.AppId, v.OutOrderid)
				continue
			}
		}

		log.Infof("当前%d,订单ID：%s", i, v.OutOrderid)

		//必须每次去池里面拿一遍，不然会阻塞
		redisHelper, err := util.AllocRedisClient()
		if err != nil {
			log.Error(err)
			return
		}
		err = tmpMethodAll(v, redisHelper, e.SleepSecond)
		if err != nil {
			log.Errorf("%s TransferApplyBaseJob error,[%s]", v.CoinName, v.OutOrderid)
		}
	}
	redis.Client.Del(cacheKey)
}

func tmpMethodAll(v *entity.FcTransfersApply, redisHelper *util.RedisClient, sleepSecond int64) error {
	defer redisHelper.Close()

	log.Infof("%s问题排查1，订单：%s,时间：%s", v.CoinName, v.OutOrderid, util.GetChinaTimeNowFormat())
	//进行短暂拦截延迟，需要在error和成功后进行删除
	interceptKey := fmt.Sprintf("%d_%s", v.AppId, v.OutOrderid)
	reply, err := redisHelper.SetNx(interceptKey, v.OutOrderid)
	if err == nil {
		if reply.(int64) == 1 {
			log.Infof("%s问题排查2，订单：%s,时间：%s", v.CoinName, v.OutOrderid, util.GetChinaTimeNowFormat())
			redisHelper.Expire(interceptKey, 86400) //一天过期
		} else {
			redisOutorderNo, _ := redisHelper.Get(interceptKey)
			log.Infof("存在正在交易中的：%s,订单号：%s，暂时跳过此订单：%s,等待下一次执行", v.CoinName, redisOutorderNo, v.OutOrderid)

			log.Infof("%s问题排查3，订单：%s,时间：%s", v.CoinName, v.OutOrderid, util.GetChinaTimeNowFormat())
			walleType := global.WalletType(v.CoinName, v.AppId)
			if walleType == status.WalletType_Cold {
				_, err := dao.FcOrderFindSuccessOrder(v.OutOrderid)
				if err == nil {
					_ = dao.FcTransfersApplyUpdateByOutNOAddErr(v.OutOrderid, int(entity.ApplyStatus_TransferOk))
				}
			} else {
				_, err := dao.FcOrderHotGetByOutOrderNo(v.OutOrderid, int(status.BroadcastStatus))
				if err == nil {
					_ = dao.FcTransfersApplyUpdateByOutNOAddErr(v.OutOrderid, int(entity.ApplyStatus_TransferOk))
				}
			}
			log.Infof("%s问题排查4，订单：%s,时间：%s", v.CoinName, v.OutOrderid, util.GetChinaTimeNowFormat())

			return err
		}
	} else {
		log.Info("redis 异常", err.Error())
		return err
	}
	log.Infof("执行订单号：%s，进行短暂拦截延迟，id：%d", v.OutOrderid, v.Id)
	//wg.Add(1)
	//go transferApplyBase(v, nil)
	go func(fc *entity.FcTransfersApply) {
		transferApplyBase(fc, nil)
	}(v)
	time.Sleep(time.Duration(sleepSecond) * time.Second)
	return nil
}

//func transferApply(v *entity.FcTransfersApply, wg *sync.WaitGroup) {
//	defer wg.Done()
//	return
//}

func transferApplyBase(v *entity.FcTransfersApply, wg *sync.WaitGroup) {
	//defer wg.Done()
	var (
		coinName string //币种名
		has      bool   //是否存在
		err      error
		//cache    string
	)

	log.Infof("%s执行订单号：%s", v.CoinName, v.OutOrderid)

	log.Infof("%s问题排查5，订单：%s,时间：%s", v.CoinName, v.OutOrderid, util.GetChinaTimeNowFormat())
	//查询是否存在相关交易的币种
	coinName = strings.ToLower(v.CoinName)
	_, has = transferService[coinName]
	if !has {
		log.Errorf("缺少相关币种服务初始化 ==> %s", coinName)
		return
	}
	//if coinName == "heco" || coinName == "tpt-heco" {
	//	log.Infof("减缓发送时间")
	//	time.Sleep(10 * time.Second)
	//}

	//验证是否已经广播或者已经存在正在执行的订单，order表 status = 4 为已经广播，status < 4为正在执行

	has, err = transferSecurityService.IsRunningOrder(v.OutOrderid, v.CoinName, v.AppId)
	if err != nil {
		log.Errorf("查询商户ID：%d,查询重复订单异常:%s，订单号：%s,币种：%s", v.AppId, err.Error(), v.OutOrderid, v.CoinName)
	}
	if has {
		log.Errorf("商户ID：%d,重复订单，订单号：%s,币种：%s", v.AppId, v.OutOrderid, v.CoinName)
		return
	}

	//判断提取数据的状态
	if v.Status != int(entity.ApplyStatus_AuditOk) {
		log.Errorf("商户：%d,订单异常状态：%d，订单号：%s,币种：%s", v.AppId, v.Status, v.OutOrderid, v.CoinName)
		return
	}
	//提前修改状态，为了避免后续修改不成功，造成重复出账
	err = orderService.SendApplyWait(v.Id)
	if err != nil {
		log.Errorf("更改订单正在构造中状态异常，outOrderId:%s,orderId:%s, error:%s", v.OutOrderid, v.OrderId, err.Error())
		return
	}
	// 验证订单
	err = validTransferApplyBase(v.Id)
	if err != nil {
		dingding.WarnDingBot.NotifyStr(fmt.Sprintf("入侵订单，outOrderId:%s,applyId:%d, error:%s", v.OutOrderid, v.Id, err.Error()))
		log.Errorf("入侵订单，outOrderId:%s,applyId:%d, error:%s", v.OutOrderid, v.Id, err.Error())
		return
	}
	log.Infof("检查订单成功： %d", v.Id)
	//热钱包需要即时回调给商户，冷钱包等待异步回调给接口通知商户
	var txid string
	//判断是热钱包币种还是冷钱包币种
	walletType := global.WalletType(coinName, v.AppId)
	if status.WalletType_Cold == walletType {
		//冷钱包
		err = transferService[coinName].TransferCold(v)
		////解冻
		//if global.TransferModel[v.CoinName] == model.TransferModelAccount {
		//	//解除拦截
		//	redisHelper.Del(interceptKey)
		//}
		if err != nil {
			log.Errorf("执行订单异常，outOrderId:%s,orderId:%s,error:%s", v.OutOrderid, v.OrderId, err.Error())
			//状态9
			err2 := orderService.SendApplyFail(v.Id)
			if err2 != nil {
				log.Errorf("更改订单失败状态异常，outOrderId:%s,orderId:%s，error：%s", v.OutOrderid, v.OrderId, err2.Error())
			}
			//不回调直接 报警IM工具，人工处理
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("执行订单异常 \n币种：%s \noutOrderId:%s\n\n orderId:%s\n\n error:%s", v.CoinName, v.OutOrderid, v.OrderId, err.Error()))
			return

		}
		if v.Eoskey == "" {
			log.Infof("冷钱包发送订单交易给server端成功，币种：%s，outOrderId:%s,orderId:%s", v.CoinName, v.OutOrderid, v.OrderId)
		} else {
			log.Infof("冷钱包代币发送订单交易给server端成功，币种：%s，outOrderId:%s,orderId:%s", v.Eoskey, v.OutOrderid, v.OrderId)
		}
		//发送成功改变 apply表的状态
		//状态7
		err = orderService.SendApplyCreateSuccess(v.Id)
		if err != nil {
			log.Errorf("更改订单成功状态异常，outOrderId:%s,orderId:%s", v.OutOrderid, v.OrderId)
		}
		return
	} else if status.WalletType_Hot == walletType {
		log.Infof("热钱包执行币种：%s", coinName)
		txid, err = transferService[coinName].TransferHot(v)
		//if global.TransferModel[v.CoinName] == model.TransferModelAccount {
		//	redisHelper.Del(interceptKey)
		//}
		if txid != "" {
			log.Info("交易成功，回调给商户")
			dao.UpdatePriorityCompletedIfExist(v.Id)
			err = orderService.NotifyToMch(v)
			log.Infof("回调状态：%t", err == nil)
			if err != nil {
				log.Errorf("交易失败，回调给商户异常：%s，outOrderId:%s,orderId:%s", err.Error(), v.OutOrderid, v.OrderId)
			}

			err = orderService.SendApplyTransferSuccess(v.Id)
			if err != nil {
				log.Errorf("更改订单交易成功状态异常，outOrderId:%s,orderId:%s", v.OutOrderid, v.OrderId)
			}

			//  完成订单后立马删除验证订单
			err = dao.CheckApplyDeleteByApplyId(int64(v.Id))
			if err != nil {
				dingding.WarnDingBot.NotifyStr(fmt.Sprintf("删除验证订单错误，outOrderId:%s,applyId：%d, error:%s", v.OutOrderid, v.Id, err.Error()))
			}
			return
		} else {
			errstr := ""
			if err != nil {
				log.Errorf("热钱包交易失败,商户：%d,币种：%s，outOrderId:%s,err:%s", v.AppId, v.CoinName, v.OutOrderid, err.Error())
				errstr = err.Error()
			}
			//txid为空，但是err没有异常
			err2 := orderService.SendApplyFail(v.Id)
			if err2 != nil {
				log.Errorf("更改订单失败状态异常，outOrderId:%s,orderId:%s", v.OutOrderid, v.OrderId)
			}
			//不回调直接 报警IM工具，人工处理
			//orderService.NotifyToMch(v, "", err.Error())
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("需要人工确认：热钱包交易失败,商户：%d,币种：%s，outOrderId:%s,error:%s", v.AppId, v.CoinName, v.OutOrderid, errstr))
			return
		}
	} else {
		log.Errorf("交易异常，币种类型配置文件缺少相关配置，币种：%s", coinName)
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("交易异常，币种类型配置文件缺少相关配置，币种：%s", coinName))
		return
	}
}

/*
验证apply表的数据
*/
func validTransferApplyBase(applyId int) error {
	log.Infof("开始检查订单： %d", applyId)
	defer log.Infof("结束检查订单： %d", applyId)
	//1. 查询db2数据库
	ca, err := dao.CheckApplyFindByApplyId(int64(applyId))
	if err != nil {
		return fmt.Errorf("查询验证订单错误： %v", err)
	}
	//2. 解密数据
	rc, err := util.DecodeCheckApplyContent(fmt.Sprintf("%d", ca.ApplyId), fmt.Sprintf("%d", ca.CreateAt), ca.Content)
	if err != nil {
		return fmt.Errorf("解密验证订单错误： %v", err)
	}
	//3. 查询出账数据
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": applyId, "address_flag": "to"})
	if err != nil {
		return fmt.Errorf("查询出账订单错误： %v", err)
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return fmt.Errorf("内部订单ID：%d,接受地址只允许一个", applyId)
	}
	toAddr := toAddrs[0].Address

	toAmount, err := decimal.NewFromString(toAddrs[0].ToAmount)

	if err != nil {
		return err
	}
	//4. 判断to地址和金额是否相同
	if strings.ToLower(rc.ToAddress) != strings.ToLower(toAddr) {
		return fmt.Errorf("to address is not equal cryptAddress=[%s],toAddress=[%s]", rc.ToAddress, toAddr)
	}
	rcStr, err := decimal.NewFromString(rc.ToAmountFloatStr)
	if err != nil || rcStr.Equal(decimal.Zero) {
		return fmt.Errorf("cryptoAmount is zero,err=%v", err)
	}
	if !rcStr.Equal(toAmount) {
		return fmt.Errorf("to amount is not equal cryptAmount=[%s],toAmounts=[%s]", rc.ToAmountFloatStr, toAmount.String())
	}
	return nil
}
