package job

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"strings"
)

//订单错误重试任务

// Job Specific Functions
type TransferApplyRetryJob struct {
}

func (e TransferApplyRetryJob) Run() {

	//查询数据,每次查询10条
	results, err := dao.FcTransfersApplyFindRetryOrder(10, global.RetryNum)
	if err != nil {
		log.Errorf("查询订单数据异常:%s", err.Error())
		return
	}

	//数据交易
	for _, v := range results {
		log.Infof("=======重推的交易=[%s]=======", v.OutOrderid)
		if v.ErrorNum >= global.RetryNum {
			log.Errorf("订单：%s,重试次数上限", v.OutOrderid)
			dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单：%s,重试次数上限", v.OutOrderid))
			continue
		}
		//查询是否存在相关交易的币种
		coinName := strings.ToLower(v.CoinName)
		_, ok := transferService[coinName]
		if !ok {
			log.Errorf("缺少相关币种服务初始化 ==> %s", coinName)
			continue
		}
		//判断币种冷热类型
		walleType := global.WalletType(v.CoinName, v.AppId)
		if walleType == status.WalletType_Cold {

			coldOrders, err := dao.FcOrderFindByOutNo(v.OutOrderid)
			if err != nil {
				log.Error(err.Error())
			}
			if len(coldOrders) != 0 {
				isOk := true
				for _, v := range coldOrders {
					if v.Status == int(status.BroadcastStatus) {
						//已经广播
						log.Errorf("订单：%s,已经广播，无需重试，请查询apply表异常", v.OuterOrderNo)
						isOk = false
						break
					} else if v.Status == int(status.BroadcastErrorStatus) || v.Status == int(status.UnknowErrorStatus) {
						log.Errorf("订单：%s,暂时禁止重试，当前状态：%d", v.OuterOrderNo, v.Status)
						isOk = false
						break
					} else if v.Status < int(status.BroadcastStatus) {
						log.Errorf("订单：%s,禁止重试,订单正在执行，当前状态：%d", v.OuterOrderNo, v.Status)
						isOk = false
						break
					}
				}
				if !isOk {
					//异常状态，不允许重推
					continue
				} else {
					log.Infof("冷钱包查询order,没有相关记录,允许重推")
				}
			}
			log.Infof("订单：%s,重试次数:%d", v.OutOrderid, v.ErrorNum)
			//修改appey状态为可执行状态
			orderService.SendApplyReviewOk(v.Id)
		} else if walleType == status.WalletType_Hot {
			hotOrders, err := dao.FcOrderHotFindByOutNo(v.OutOrderid)
			if err != nil {
				log.Error(err.Error())
			}
			if len(hotOrders) != 0 {
				//校验状态
				isOk := true
				for _, v := range hotOrders {
					if v.Status == int(status.BroadcastStatus) {
						//已经广播
						log.Errorf("订单：%s,已经广播，无需重试，请查询apply表异常", v.OuterOrderNo)
						isOk = false
						break
					} else if v.Status == int(status.BroadcastErrorStatus) || v.Status == int(status.UnknowErrorStatus) {
						log.Errorf("订单：%s,暂时禁止重试，当前状态：%d", v.OuterOrderNo, v.Status)
						isOk = false
						break
					} else if v.Status < int(status.BroadcastStatus) {
						log.Errorf("订单：%s,禁止重试,订单正在执行，当前状态：%d", v.OuterOrderNo, v.Status)
						isOk = false
						break
					}
				}
				if !isOk {
					//异常状态不允许重推
					continue
				}
			} else {
				log.Errorf("热钱包查询order,没有相关记录,允许重推")
			}
			log.Infof("订单：%s,重试次数:%d", v.OutOrderid, v.ErrorNum)
			//修改appey状态为可执行状态
			orderService.SendApplyReviewOk(v.Id)
		} else {
			//todo 钉钉
			log.Infof("订单：%s,未知类型，：%s", v.OutOrderid, string(walleType))
		}
	}
}
