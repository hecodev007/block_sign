package service

import (
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model"
)

//订单经过一系列认证 保存在表，然后有个服务专门从db拉订单，异常订单都是靠这个服务拉单重推
type OrderService interface {
	//商户提交订单保存(utxo交易模型)
	SaveTransferByUtxo(params model.TransferParams, callBackUrl string) (orderId int64, err error)

	//商户提交订单保存(账户交易模型)
	//注意某些特殊utxo币种，在交易过程走账户模型模式（比如BTM，有21utxo限制）
	SaveTransferByAccount(params model.TransferParams, callBackUrl string) (orderId int64, err error)

	// 更改需要改变orderId，原先php业务设计表问题。
	//订单发送给walletserver成功之后的事件触发，更改apply表状态，状态3，处于构造中状态
	//主要是为了避免 发送之后发生异常没有更改状态会造成重复出账
	SendApplyWait(applyId int) error

	//审核驳回
	SendAuditFail(applyId int) error

	// 更改需要改变orderId，原先php业务设计表问题。
	//订单发送给walletserver成功之后的事件触发，更改apply表状态，状态7
	SendApplyCreateSuccess(applyId int) error

	// 更改需要改变orderId，原先php业务设计表问题。
	//订单发送给walletserver成功之后的事件触发，更改apply表状态，状态30
	SendApplyTransferSuccess(applyId int) error

	// 更改需要改变orderId，原先php业务设计表问题。
	//订单发送给walletserver失败之后的事件触发，更改apply表状态,等待重试状态 8
	SendApplyRetry(applyId int) error

	//修改为重试一次
	SendApplyRetryOnce(applyId int, errNum int64) error

	// 更改需要改变orderId，原先php业务设计表问题。
	//订单发送给walletserver失败之后的事件触发，更改apply表状态，重试状态一定次数之后定义为9 最终失败
	SendApplyFail(applyId int) error

	// 更改需要改变orderId，原先php业务设计表问题。
	//设置审核完成状态，主要用于重置一些特殊失败状态，5，6，8等失败状态，7，8状态需要人工校验
	SendApplyReviewOk(applyId int) error

	//设置回滚状态
	//一般是出账失败订单 需要回滚的时候设置回滚（比如eos账号不存在，eth合约地址转账等等）
	SendApplyRollback(applyId int) error

	//异步回调给商户
	NotifyToMch(ta *entity.FcTransfersApply) error

	//同步回调给商户
	NotifyToMchByOutOrderId(outOrderId string) error

	//多地址出账版本 同步回调给交易所
	NotifyToMchV2(ta *entity.FcTransfersApply, order *entity.FcOrder, orderTx *entity.FcOrderTxs) error

	//订单是否允许重推
	IsAllowRepush(outOrderId string) error

	//获取订单信息
	GetApplyOrder(outOrderId string) (*entity.FcTransfersApply, error)

	//获取订单信息
	GetApplyOrderByOrderNo(orderNo string) (*entity.FcTransfersApply, error)

	//废弃订单，只要状态是30 并且钱包状态为4
	AbandonedOrder(outOrderId string) error
}
