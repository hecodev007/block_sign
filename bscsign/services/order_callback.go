package services

import (
	"context"
	"encoding/json"
	"github.com/bsc-sign/conf"
	"github.com/bsc-sign/model"
	"github.com/bsc-sign/redis"
	"github.com/bsc-sign/util"
	"github.com/bsc-sign/util/dingding"
	"github.com/bsc-sign/util/log"
	"time"
)

const (
	// 存放调用blockchains接口失败的订单数据对应的redis key
	orderCallbackWaitingReCallbackKey = "bsc_sign_callback_recall"

	// 执行循环任务的时间间隔
	processCallbackInterval = time.Millisecond * 200
)

type OrderCallback struct {
	ctx context.Context
}

func NewOrderCallback(ctx context.Context) *OrderCallback {
	orderCallback := &OrderCallback{
		ctx: ctx,
	}
	go orderCallback.StartReCallLoop()
	return orderCallback
}

func (o *OrderCallback) StartReCallLoop() {
	timer := time.NewTimer(processCallbackInterval)
loop:
	for {
		select {
		case <-timer.C:
			o.processCallback()
			timer.Reset(processCallbackInterval)
		case <-o.ctx.Done():
			log.Info("reCallLoop任务已停止")
			break loop
		}
	}
}

// 处理回调blockchain接口失败的订单，这些订单都是已经执行了签名逻辑，但是通过HTTP请求调用失败
// 订单会被存放在redis的List结构，从左边pop弹出，右边push进去，所以有序
// 弹出一笔订单，首先检查是否到达执行时间`sendTime`
// 如果到达，发送HTTP请求调用blockchain接口
// 如果回调还是失败，修改各项参数，继续push回去列表（加入到链表的末尾），等待下一次执行
func (o *OrderCallback) processCallback() {
	val, err := redis.Client.ListPop(orderCallbackWaitingReCallbackKey)
	if err != nil {
		log.Errorf("读取redis失败:%v", err)
		return
	}
	if val == nil {
		// 缓存列表没有可执行的数据
		return
	}

	reCallback := &model.ReCallback{}
	unmarshalErr := json.Unmarshal(val, reCallback)
	if unmarshalErr != nil {
		log.Errorf("processCallback json unmarshal err:%v", unmarshalErr)
		o.rePushToCache(reCallback)
		return
	}

	if reCallback.SendTime > time.Now().Unix() {
		// 未到执行时间
		// 把订单放回到链表末尾
		o.rePushToCache(reCallback)
		return
	}
	log.Infof("获取到需要执行callback回调的订单数据 %s", reCallback.Data.OuterOrderNo)
	sendErr := o.sendPostRequest(&reCallback.CallbackReqParams)
	if sendErr == nil {
		// 已成功回调
		return
	}

	reCallback.SendTime = nextSendTime() // 重新计时
	reCallback.ErrMsg = sendErr.Error()
	reCallback.RetryCount += 1 // 重试次数+1

	log.Infof("调用blockchain回调接口再次失败，已重试了 %d 次 失败信息:%v", reCallback.RetryCount, sendErr)
	o.rePushToCache(reCallback)
}

func (o *OrderCallback) rePushToCache(reCallback *model.ReCallback) {
	buf, marshalErr := json.Marshal(reCallback)
	if marshalErr != nil {
		log.Errorf("[Callback] json Marshal err: %v", marshalErr)
		return
	}
	if err := redis.Client.ListRPush(orderCallbackWaitingReCallbackKey, buf); err != nil {
		log.Errorf("Push数据到redis缓存失败:%v", err)
	}
}

func (o *OrderCallback) Send(outerOrderNo string, orderHotId int, resp []byte, err error) {
	code := model.ResponseCodeSuccess
	msg := "success"
	data := model.CallbackReqParamsData{TxId: string(resp), OuterOrderNo: outerOrderNo, OrderHotId: orderHotId, Success: true}
	if err != nil {
		code = model.ResponseCodeFail
		msg = err.Error()
		data.TxId = ""
		data.Success = false
	}

	params := &model.CallbackReqParams{
		Data:    data,
		Code:    code,
		Message: msg,
	}

	sendErr := o.sendPostRequest(params)
	if sendErr == nil {
		return
	}

	// 如果回调blockchains失败
	// 钉钉通知值班人员
	dingding.NotifyError("签名服务回调blockchain失败", outerOrderNo, sendErr.Error())

	// 把数据保存到缓存
	// 然后会定时取出继续尝试回调
	reCallbackModel := model.ReCallback{
		SendTime:          nextSendTime(),
		ErrMsg:            sendErr.Error(),
		CallbackReqParams: *params,
	}
	buf, marshalErr := json.Marshal(reCallbackModel)
	if marshalErr != nil {
		log.Errorf("[Callback] json Marshal err: %v", marshalErr)
		return
	}

	if err = redis.Client.ListRPush(orderCallbackWaitingReCallbackKey, buf); err != nil {
		log.Errorf("Push数据到redis缓存失败:%v", err)
	}
	log.Infof("Push数据到redis缓存列表完成")
}

func (o *OrderCallback) sendPostRequest(params *model.CallbackReqParams) error {
	path := "/v3/sign/callback"
	reqResult, err := util.PostJson(conf.Config.Callback.Url+path, params)
	if err != nil {
		log.Errorf("订单%s 回调blockchain出错:%v", params.Data.OuterOrderNo, err)
	} else {
		log.Infof("订单%s 回调blockchain完成:%s", params.Data.OuterOrderNo, string(reqResult))
	}
	return err
}

func nextSendTime() int64 {
	return time.Now().Add(10 * time.Second).Unix()
}
