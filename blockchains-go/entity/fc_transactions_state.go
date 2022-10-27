package entity

import (
	"time"
)

type FcTransactionsState struct {
	Id          int                     `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId       int                     `json:"app_id" xorm:"comment('商户ID') INT(11)"`
	Sfrom       string                  `json:"sfrom" xorm:"not null comment('商户名') VARCHAR(50)"`
	CoinName    string                  `json:"coin_name" xorm:"not null comment('币种名称') VARCHAR(15)"`
	OrderId     string                  `json:"order_id" xorm:"not null comment('订单ID') VARCHAR(100)"`
	OutOrderid  string                  `json:"out_orderid" xorm:"not null default '' comment('商户推送交易ID') VARCHAR(100)"`
	Txid        string                  `json:"txid" xorm:"not null default '' comment('txid') VARCHAR(150)"`
	CallBack    string                  `json:"call_back" xorm:"not null default '' comment('回调url') VARCHAR(255)"`
	Msg         string                  `json:"msg" xorm:"not null default '' comment('失败信息') VARCHAR(200)"`
	Eoskey      string                  `json:"eoskey" xorm:"VARCHAR(15)"`
	Eostoken    string                  `json:"eostoken" xorm:"VARCHAR(100)"`
	Memo        string                  `json:"memo" xorm:"not null default '' comment('eos memo') VARCHAR(100)"`
	Data        string                  `json:"data" xorm:"LONGTEXT"`
	RetryNum    int                     `json:"retry_num" xorm:"not null default 0 comment('重试次数') TINYINT(4)"`
	CallbackMsg string                  `json:"callback_msg" xorm:"TEXT"`
	CreateTime  int64                   `json:"create_time" xorm:"INT(11)"`
	Lastmodify  time.Time               `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	PushStatus  int                     `json:"push_status" xorm:"not null default 1 comment('推送状态 0已推送 1未推送 2推送失败') index TINYINT(4)"`
	Status      FcTransactionsStateCode `json:"status" xorm:"not null default 0 comment('成功状态 0成功 1失败') TINYINT(4)"`
	MemoEncrypt string                  `json:"status" xorm:"'memo_encrypt'"` //memo加密后信息 应对浏览器信息只显示加密后的信息的币种
	//State       int       `json:"state" xorm:"not null default 0 comment('1未推送2推送成功') TINYINT(4)"`
}

type FcTransactionsStateCode int

const (
	FcTransactionsStatesWait    FcTransactionsStateCode = -1
	FcTransactionsStatesSuccess FcTransactionsStateCode = 0
	FcTransactionsStatesFail    FcTransactionsStateCode = 1
)
