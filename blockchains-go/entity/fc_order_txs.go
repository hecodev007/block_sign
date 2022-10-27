package entity

import (
	"time"
)

var (
	OtxStatusMap map[OrderTxStatus]string
)

type OrderTxStatus int

const (
	OtxInit              OrderTxStatus = -2 //新添加的订单交易数据
	OtxSignDataBuilding  OrderTxStatus = -1 //签名数据构建中
	OtxWaitingSign       OrderTxStatus = 0  //构建签名数据成功，待签名
	OtxSigning           OrderTxStatus = 1  //签名中
	OtxSignSuccess       OrderTxStatus = 2  //签名成功
	OtxBroadcasting      OrderTxStatus = 3  //正在广播（需要强制重推）
	OtxBroadcastSuccess  OrderTxStatus = 4  //广播成功 最终状态（不可重推）
	OtxSignFailure       OrderTxStatus = 6  //签名失败 最终状态
	OtxBroadcastFailure  OrderTxStatus = 7  //广播失败 最终状态（需要强制重推）
	OtxSignTimeout       OrderTxStatus = 8  //签名超时 最终状态
	OtxUnknownErr        OrderTxStatus = 9  //广播出现未知错误（需要强制重推）
	OtxBuildSignDataErr  OrderTxStatus = 10 //构建签名数据失败 最终状态
	OtxTrxFeeFailure     OrderTxStatus = 11 //打手续费失败 最终状态
	OtxTrxOnChainFailure OrderTxStatus = 12 //链上订单失败 最终状态（不可重推）
	OtxCanceled          OrderTxStatus = 13 //已取消，即交易已回滚 最终状态（不可重推）

)

func init() {
	OtxStatusMap = map[OrderTxStatus]string{}
	OtxStatusMap[OtxInit] = "初始化完成"
	OtxStatusMap[OtxSignDataBuilding] = "签名数据构建中"
	OtxStatusMap[OtxWaitingSign] = "构建签名数据成功，待签名"
	OtxStatusMap[OtxSigning] = "签名中"
	OtxStatusMap[OtxSignSuccess] = "签名成功"
	OtxStatusMap[OtxBroadcasting] = "正在广播"
	OtxStatusMap[OtxBroadcastSuccess] = "广播成功"
	OtxStatusMap[OtxSignFailure] = "签名失败，可直接重推"
	OtxStatusMap[OtxBroadcastFailure] = "广播失败"
	OtxStatusMap[OtxSignTimeout] = "签名超时，可直接重推"
	OtxStatusMap[OtxUnknownErr] = "广播出现未知错误"
	OtxStatusMap[OtxBuildSignDataErr] = "构建签名数据失败，可直接重推"
	OtxStatusMap[OtxTrxFeeFailure] = "打手续费失败，可直接重推"
	OtxStatusMap[OtxTrxOnChainFailure] = "×链上订单失败"
	OtxStatusMap[OtxCanceled] = "×已取消"
}

type FcOrderTxs struct {
	Id           int64         `json:"id" xorm:"not null pk autoincr BIGINT(11)"`
	SeqNo        string        `json:"seq_no"  xorm:" NOT NULL comment('交易流水号号')  VARCHAR(64)"`
	ParentSeqNo  string        `json:"parent_seq_no" xorm:" default '' comment('父流水号')  VARCHAR(64)"`
	TxId         string        `json:"tx_id" xorm:" comment('交易哈希')  VARCHAR(256)"`
	OuterOrderNo string        `json:"outer_order_no" xorm:" comment('外部订单号')  VARCHAR(128)"`
	InnerOrderNo string        `json:"inner_order_no" xorm:" comment('内部订单号')  VARCHAR(128)"`
	Mch          string        `json:"mch" xorm:" comment('商户名')  VARCHAR(36)"`
	Chain        string        `json:"chain" xorm:" comment('链')  VARCHAR(24)"`
	CoinCode     string        `json:"coin_code" xorm:" comment('代币编码')  VARCHAR(24)"`
	Contract     string        `json:"contract" xorm:" comment('合约地址')  VARCHAR(256)"`
	FromAddress  string        `json:"from_address" xorm:" comment('from地址')  VARCHAR(256)"`
	ToAddress    string        `json:"to_address" xorm:" comment('to地址')  VARCHAR(256)"`
	Amount       string        `json:"amount" xorm:" comment('交易金额')  VARCHAR(64)"`
	Status       OrderTxStatus `json:"status" xorm:"default -1 comment('状态')  INT"`
	Sort         int           `json:"sort" xorm:" comment('排序')  INT"`
	Nonce        int64         `json:"nonce" xorm:" default -1 comment('随机数')  BIGINT"`
	SignReqData  string        `json:"sign_req_data" xorm:" comment('待签名数据')  VARCHAR(256)"`
	SignerNo     string        `json:"signer_no" xorm:" comment('签名机编号')  VARCHAR(24)"`
	ErrMsg       string        `json:"err_msg" xorm:" comment('错误信息')  VARCHAR(2048)"`
	FreezeUnlock int           `json:"freeze_unlock" xorm:" comment('冻结的金额是否已解锁，0：未解锁；1：已解锁')  INT"`
	CreateTime   time.Time     `json:"create_time" xorm:"not null comment('创建时间') DATETIME"`
	UpdateTime   time.Time     `json:"update_time" xorm:"not null comment('修改时间') DATETIME"`
}

func (o *FcOrderTxs) TableName() string {
	return "fc_order_txs"
}

func (o *FcOrderTxs) IsCompleted() bool {
	return o.Status == OtxBroadcastSuccess
}

func (o *FcOrderTxs) IsChainFailure() bool {
	return o.Status == OtxTrxOnChainFailure
}

func (o *FcOrderTxs) IsCanceled() bool {
	return o.Status == OtxCanceled
}

func (o *FcOrderTxs) CanForceCancel() bool {
	if !o.CanRePush() {
		return false
	}
	return o.Status == OtxBroadcasting || o.Status == OtxBroadcastFailure || o.Status == OtxUnknownErr
}

func (o *FcOrderTxs) CanCancel() bool {
	return o.CanRePush()
}

func (o *FcOrderTxs) NeedPush() bool {
	return o.Status == OtxTrxOnChainFailure || o.Status == OtxBroadcastSuccess
}

func (o *FcOrderTxs) IsProcessing() bool {
	return o.Status == OtxInit ||
		o.Status == OtxSignDataBuilding ||
		o.Status == OtxWaitingSign ||
		o.Status == OtxSigning ||
		o.Status == OtxSignSuccess ||
		o.Status == OtxBroadcasting
}

// CanRePush 是否忽略统计到订单金额
// 链上失败的交易和已废弃的交易不需要统计
func (o *FcOrderTxs) CanRePush() bool {
	return o.Status == OtxSignFailure ||
		o.Status == OtxSignTimeout ||
		o.Status == OtxBuildSignDataErr ||
		o.Status == OtxTrxFeeFailure
}
