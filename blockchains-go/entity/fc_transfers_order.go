package entity

import (
	"time"
)

type FcTransfersOrder struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	ApplyId    string `json:"apply_id" xorm:"not null comment('原始订单ID集合') VARCHAR(255)"`
	OutOrderid string `json:"out_orderid" xorm:"not null comment('原始订单外部订单集合') VARCHAR(255)"`
	Code       string `json:"code" xorm:"not null comment('合并订单号，唯一订单号') VARCHAR(100)"`
	OrderId    string `json:"order_id" xorm:"not null comment('合并订单内部订单号') VARCHAR(100)"`
	MchId      int    `json:"mch_id" xorm:"not null comment('商户ID') INT(11)"`
	Platform   string `json:"platform" xorm:"not null VARCHAR(20)"`
	Type       string `json:"type" xorm:"not null comment('订单类型') VARCHAR(15)"`
	CoinName   string `json:"coin_name" xorm:"not null VARCHAR(15)"`
	Status     int    `json:"status" xorm:"not null default 1 comment('0  取消
1  等待构建
2  构建中
3  构建成功
4  广播成功
5  构建失败
6  签名失败
7  广播失败
8  签名超时
9  异常错误') TINYINT(4)"`
	Data       string    `json:"data" xorm:"comment('订单参数') TEXT"`
	CreateData string    `json:"create_data" xorm:"comment('create参数') TEXT"`
	Eoskey     string    `json:"eoskey" xorm:"VARCHAR(255)"`
	Eostoken   string    `json:"eostoken" xorm:"VARCHAR(255)"`
	Memo       string    `json:"memo" xorm:"VARCHAR(255)"`
	Fee        string    `json:"fee" xorm:"not null default 0.00000000000000000000 DECIMAL(50,20)"`
	Num        int       `json:"num" xorm:"not null default 0 comment('重推次数') INT(11)"`
	Txid       string    `json:"txid" xorm:"VARCHAR(255)"`
	Url        string    `json:"url" xorm:"not null comment('请求create接口url') VARCHAR(255)"`
	ErrorMsg   string    `json:"error_msg" xorm:"not null TEXT"`
	IsRetry    int       `json:"is_retry" xorm:"not null default 0 comment('0非手动1手动重推') TINYINT(4)"`
	IsMerge    int       `json:"is_merge" xorm:"not null default 0 comment('0非合并订单1合并订单') TINYINT(4)"`
	CreateTime int       `json:"create_time" xorm:"not null default 0 INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
}
