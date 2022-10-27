package job

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
)

// 回调任务
// 目前在php这边
// walletserver回调结果到/v2/receive_info接口
// 判断回调状态，如果成功，回调txid给商户.(广播失败全部不重推，，签名失败重推5次)
// 分析，推送结果会写表
type TransferApplyCallBackJob struct {
}

func (e TransferApplyCallBackJob) Run() {

	//查询10条，重试小于6，等待推送状态的数据
	results, err := dao.FcTransactionsStateFindByStatus(entity.FcTransactionsStatesWait, global.RetryCallback, 10)
	if err != nil {
		log.Errorf("查询需要回调的信息异常：%s", err.Error())
		return
	}
	if len(results) > 0 {
		log.Info("=======执行回调的信息=======")
	}

	for _, v := range results {

		mch, err := dao.FcMchFindById(v.AppId)

		//respData, err := util.PostJsonData(v.CallBack, []byte(v.Data))
		respData, err := util.PostByteForCallBack(v.CallBack, []byte(v.Data), mch.ApiKey, mch.ApiSecret)
		if err != nil {
			log.Errorf("发送给商户回调异常：URL：%s，outOrderNo：%s, error:%s", v.CallBack, v.OutOrderid, err.Error())
			err = dao.FcTransactionsStateUpdateAddErr(v.Id, entity.FcTransactionsStatesWait, "", err.Error())
			if err != nil {
				log.Errorf("更新回调异常记录异常：%s", err.Error())
			}
			continue
		}
		log.Infof("outOrderNo:%s,回调返回内容：%s", v.OutOrderid, string(respData))
		err = dao.FcTransactionsStateUpdateState(v.Id, entity.FcTransactionsStatesSuccess, string(respData))
		if err != nil {
			log.Errorf("更新回调异常记录异常：%s,请求返回内容:%s", err.Error(), string(respData))
		} else {
			log.Infof("订单：%s,回调商户成功", v.OutOrderid)
		}

		//resp := model.DecodeBCallbackResp(respData)
		//if resp.Code != 0 {
		//	//回调失败
		//	log.Errorf("回调异常,请求返回内容:%s", string(respData))
		//	err = dao.FcTransactionsStateUpdateAddErr(v.Id, entity.FcTransactionsStatesWait, "", string(respData))
		//	if err != nil {
		//		log.Errorf("更新回调异常记录异常：%s,请求返回内容:%s", err.Error(), string(respData))
		//	}
		//	continue
		//} else {
		//	//回调成功
		//	log.Infof("回调成功,请求返回内容:%s", string(respData))
		//	err = dao.FcTransactionsStateUpdateState(v.Id, entity.FcTransactionsStatesSuccess, string(respData))
		//	if err != nil {
		//		log.Errorf("更新回调异常记录异常：%s,请求返回内容:%s", err.Error(), string(respData))
		//	} else {
		//		log.Infof("订单：%s,回调商户成功", v.OutOrderid)
		//	}
		//	continue
		//}
	}

}
