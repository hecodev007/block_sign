package entity

import (
	"time"
)

type FcHuobiCommonSymbols struct {
	Id              int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Symbol          string    `json:"symbol" xorm:"not null comment('交易对') unique VARCHAR(20)"`
	BaseCurrency    string    `json:"base_currency" xorm:"not null comment('基础币种') unique(base_quote_currency) VARCHAR(15)"`
	QuoteCurrency   string    `json:"quote_currency" xorm:"not null comment('计价币种') unique(base_quote_currency) VARCHAR(15)"`
	PricePrecision  int       `json:"price_precision" xorm:"not null comment('价格精度位数（0为个位）') INT(11)"`
	AmountPrecision int       `json:"amount_precision" xorm:"not null comment('数量精度位数（0为个位，注意不是交易限额）') INT(11)"`
	SymbolPartition string    `json:"symbol_partition" xorm:"not null comment('交易区, main主区，innovation创新区，bifurcation分叉区') VARCHAR(20)"`
	Createtime      int       `json:"createtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify      time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改时间') TIMESTAMP"`
}
