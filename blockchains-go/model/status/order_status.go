package status

//order表状态
type OrderStatus int

func (o OrderStatus) Int() int {
	return int(o)
}

const (
	All                  OrderStatus = -1
	CreateStatus         OrderStatus = 0  //构建完成
	PushStatus           OrderStatus = 1  //推入队列,签名中
	SignStatus           OrderStatus = 2  //已签名成功
	BroadcastingStatus   OrderStatus = 3  //正在广播中
	BroadcastStatus      OrderStatus = 4  //广播成功
	CreateErrorStatus    OrderStatus = 5  //构建失败
	SignErrorStatus      OrderStatus = 6  //签名失败
	BroadcastErrorStatus OrderStatus = 7  //广播失败
	PendingTimeoutStatus OrderStatus = 8  //pending超时
	UnknowErrorStatus    OrderStatus = 9  //未知异常
	AbandonedTransaction OrderStatus = 10 //丢弃的交易
	RollbackTransaction  OrderStatus = 11 //回滚的交易
	WaitResetTransaction OrderStatus = 12 //等待重置交易
	WaitSign             OrderStatus = 13 //等待签名（新签名服务）
	failureOnChain       OrderStatus = 14 //链上失败
	NotFoundOnChain      OrderStatus = 15 //链上404
)

var StatusDesc = map[OrderStatus]string{
	All:                  " ",
	CreateStatus:         "构建完成",
	PushStatus:           "推入队列,签名中",
	SignStatus:           "已签名成功",
	BroadcastingStatus:   "正在广播中",
	BroadcastStatus:      "广播成功",
	CreateErrorStatus:    "构建失败",
	SignErrorStatus:      "签名失败",
	BroadcastErrorStatus: "广播失败",
	PendingTimeoutStatus: "pending超时",
	UnknowErrorStatus:    "未知异常",
	AbandonedTransaction: "丢弃的交易",
	RollbackTransaction:  "回滚的交易",
	WaitResetTransaction: "等待重置交易",
	WaitSign:             "等待签名（新签名服务）",
	failureOnChain:       "链上失败",
	NotFoundOnChain:      "链上404",
}

func GetStatusDesc(code OrderStatus) string {
	msg, ok := StatusDesc[code]
	if ok {
		return msg
	}
	return "未知状态"
}
