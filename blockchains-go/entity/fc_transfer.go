package entity

import (
	"time"
)

type FcTransfer struct {
	Id           int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId        int       `json:"app_id" xorm:"not null index INT(11)"`
	CoinId       string    `json:"coin_id" xorm:"not null VARCHAR(20)"`
	Amount       string    `json:"amount" xorm:"not null default 0.0000000000 comment('申请转账金额') DECIMAL(23,10)"`
	Fee          string    `json:"fee" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(23,10)"`
	ActualAmount string    `json:"actual_amount" xorm:"not null default 0.0000000000 comment('实际到账金额') DECIMAL(23,10)"`
	Balance      string    `json:"balance" xorm:"not null default 0.0000000000 comment('平台余额') DECIMAL(23,10)"`
	Updatetime   time.Time `json:"updatetime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	Mation       string    `json:"mation" xorm:"not null default '0' comment('构造交易返回的交易数据') VARCHAR(200)"`
	Txid         string    `json:"txid" xorm:"not null default '0' index VARCHAR(80)"`
	InAddress    string    `json:"in_address" xorm:"not null index VARCHAR(200)"`
	Status       int       `json:"status" xorm:"not null default 0 comment('0,待处理,1.成功，2正在审核，3,正在处理,4驳回，5，异常') TINYINT(4)"`
	AppAddressId int       `json:"app_address_id" xorm:"not null comment('客户端地址ID') INT(11)"`
	OutAddress   string    `json:"out_address" xorm:"not null comment('转出地址') VARCHAR(200)"`
	MuAddressId  int       `json:"mu_address_id" xorm:"not null comment('后端转出地址ID') INT(11)"`
	AuditorKey   int       `json:"auditor_key" xorm:"not null default 0 comment('审核等级') TINYINT(4)"`
	Type         int       `json:"type" xorm:"comment('2巨额用户提现，1普通') TINYINT(4)"`
	Addtime      int64     `json:"addtime" xorm:"BIGINT(20)"`
	Remark       string    `json:"remark" xorm:"comment('记录') VARCHAR(50)"`
	Edition      int       `json:"edition" xorm:"default 0 comment('区块上执行次数') INT(11)"`
	Implement    int       `json:"implement" xorm:"default 0 comment('0未执行1已执行2执行完成') TINYINT(1)"`
	TradeTime    int64     `json:"trade_time" xorm:"not null comment('交易时间戳') BIGINT(20)"`
	Block        int       `json:"block" xorm:"comment('区块高度') INT(11)"`
	TradeSn      string    `json:"trade_sn" xorm:"unique VARCHAR(20)"`
}
