package entity

type FcMchMoneyLog struct {
	Id            int    `json:"id" xorm:"not null pk autoincr index(enter_id) INT(11)"`
	AppId         int    `json:"app_id" xorm:"not null default 0 comment('商户id') index(enter_id) INT(10)"`
	CoinId        int    `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	CoinName      string `json:"coin_name" xorm:"comment('币种名称') VARCHAR(15)"`
	TradeNo       string `json:"trade_no" xorm:"comment('交易流水号') VARCHAR(100)"`
	Type          int    `json:"type" xorm:"not null default 0 comment('订单类型 1 购买服务  2 加购地址 3 提现  4充值   5转账服务费  6后台充值 7后台减少') index(enter_id) TINYINT(3)"`
	Dir           int    `json:"dir" xorm:"not null default 0 comment('资金流向  1 可用转冻结   2 冻结转可用  3 冻结减少 4 可用减少 5 可用增加') TINYINT(3)"`
	Amount        string `json:"amount" xorm:"not null default 0.000000000000000000 comment('影响金额') DECIMAL(40,18)"`
	AccountAmount string `json:"account_amount" xorm:"not null default 0.000000000000000000 comment('可用金额') DECIMAL(40,18)"`
	FreezeAmount  string `json:"freeze_amount" xorm:"not null default 0.000000000000000000 comment('冻结金额') DECIMAL(40,18)"`
	Info          string `json:"info" xorm:"not null comment('备注') VARCHAR(255)"`
	AddTime       int    `json:"add_time" xorm:"not null default 0 comment('添加时间') INT(10)"`
}
