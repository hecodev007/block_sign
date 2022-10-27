package entity

import (
	"time"
)

type FcHooTradeListOrig struct {
	Id           int    `json:"id" xorm:"not null pk autoincr comment('è‡ªå¢žID') INT(11)"`
	TradeId      string `json:"trade_id" xorm:"not null comment('äº¤æ˜“IDï¼Œå”¯â¼€äº¤æ˜“ID') unique(tradeid_wallettype_status) VARCHAR(64)"`
	WalletType   string `json:"wallet_type" xorm:"not null default '' comment('é’±åŒ…ç±»åž‹ï¼Œä¸ªäººé’±åŒ…(wallet)ï¼Œ å…±ç®¡(multiwallet)') unique(tradeid_wallettype_status) VARCHAR(18)"`
	Status       int    `json:"status" xorm:"not null comment('äº¤æ˜“çŠ¶æ€ï¼Œ1ï¼ˆæˆåŠŸï¼‰ 2ï¼ˆå†»ç»“ï¼‰') unique(tradeid_wallettype_status) TINYINT(2)"`
	Ticket       string `json:"ticket" xorm:"not null default '' comment('è¯·æ±‚å‡­è¯') VARCHAR(50)"`
	CoinName     string `json:"coin_name" xorm:"not null comment('å¸åç§°ï¼Œå¸çš„è‹±æ–‡å¤§å†™å­—æ¯ç®€ç§°ï¼Œå¦‚:BTC') VARCHAR(15)"`
	Platform     string `json:"platform" xorm:"not null default '' comment('æŽ¥å…¥å¹³å°') VARCHAR(15)"`
	Address      string `json:"address" xorm:"not null default '' comment('交易地址，每笔交易对应的地址') VARCHAR(255)"`
	FromAddress  string `json:"from_address" xorm:"not null default '' comment('å‘é€åœ°å€') VARCHAR(255)"`
	ToAddress    string `json:"to_address" xorm:"not null default '' comment('æŽ¥æ”¶åœ°å€') VARCHAR(255)"`
	TradeType    string `json:"trade_type" xorm:"not null default '' comment('äº¤æ˜“ç±»åž‹') VARCHAR(40)"`
	TradeAmount  string `json:"trade_amount" xorm:"not null default 0.00000000000000000000 comment('交易金额') DECIMAL(65,20)"`
	TradeFee     string `json:"trade_fee" xorm:"not null default 0.00000000000000000000 comment('交易费用') DECIMAL(65,20)"`
	FeeUnit      string `json:"fee_unit" xorm:"not null default '' comment('费用的单位') VARCHAR(15)"`
	RamAmount    string `json:"ram_amount" xorm:"not null default 0.00000000000000000000 comment('ram数量') DECIMAL(65,20)"`
	RamCoinPrice string `json:"ram_coin_price" xorm:"not null default 0.00000000000000000000 comment('ram兑币的价格') DECIMAL(65,20)"`
	RamPrice     string `json:"ram_price" xorm:"not null default 0.00000000000000000000 comment('ram价格（eos/kb)') DECIMAL(65,20)"`
	TradeTime    int    `json:"trade_time" xorm:"not null default 0 comment('äº¤æ˜“æ—¶é—´ï¼ˆæ—¶é—´æˆ³ï¼‰ï¼Œå•ä½ç§’') INT(11)"`
	Memo         string `json:"memo" xorm:"not null default '' comment('EOSçš„ Memo') VARCHAR(100)"`
	Txid         string `json:"txid" xorm:"not null default '' comment('é“¾ä¸Šäº¤æ˜“ID') VARCHAR(80)"`
	Remark       string `json:"remark" xorm:"not null default '' comment('交易备注') VARCHAR(255)"`
	Phone        string `json:"phone" xorm:"not null default '' comment('手机号, 用于eos小程序收入') VARCHAR(20)"`
	//Status       int       `json:"status_" xorm:"not null default 0 comment('è¯¥æ¡è®°å½•çš„çŠ¶æ€ï¼Œ0-æœªå¤„ç†çš„åŽŸå§‹è®°å½•ï¼Œ1-å¯¹åº”ä¸åŒçš„é”™è¯¯ç±»åž‹') TINYINT(2)"`
	Createtime time.Time `json:"createtime" xorm:"not null default '0000-00-00 00:00:00' comment('åˆ›å»ºæ—¶é—´') DATETIME"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('æœ€è¿‘ä¿®æ”¹æ—¶é—´') TIMESTAMP"`
}
