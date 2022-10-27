package entity

import (
	"time"
)

type FcCoinProcess struct {
	Id             int64     `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinApplyId    int64     `json:"coin_apply_id" xorm:"comment('上币申请单id') BIGINT(20)"`
	OuterOrderId   string    `json:"outer_order_id" xorm:"comment('自动生成外部订单号') VARCHAR(100)"`
	Platform       string    `json:"platform" xorm:"VARCHAR(100)"`
	TaskId         int64     `json:"task_id" xorm:"comment('处理任务id(自生成)') BIGINT(20)"`
	Amount         string    `json:"amount" xorm:"VARCHAR(30)"`
	FromAddress    string    `json:"from_address" xorm:"comment('测试发送地址') VARCHAR(255)"`
	ToAddress      string    `json:"to_address" xorm:"comment('测试接收地址') VARCHAR(255)"`
	ProcessStatus  int       `json:"process_status" xorm:"default 0 comment('测试状态(0:未处理, 1: 处理中,2: 处理成功, 3: 处理失败)') TINYINT(1)"`
	RechargeStatus int       `json:"recharge_status" xorm:"default 0 comment('充值测试状态(0:未处理 2:入账成功)') TINYINT(1)"`
	TradeStatus    int       `json:"trade_status" xorm:"default 0 comment('交易测试状态(0:未测试 2:交易成功,3:交易失败)') TINYINT(1)"`
	ProcessRemark  string    `json:"process_remark" xorm:"comment('测试备注(所有测试结果内容都写此)') VARCHAR(200)"`
	TxId           string    `json:"tx_id" xorm:"comment('交易id') VARCHAR(100)"`
	TxUrl          string    `json:"tx_url" xorm:"comment('交易链接') VARCHAR(200)"`
	CreateAt       time.Time `json:"create_at" xorm:"DATETIME"`
	UpdateAt       time.Time `json:"update_at" xorm:"DATETIME"`
}
