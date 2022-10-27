package entity

type FcTxClear struct {
	Id            int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	Coin          string `json:"coin" xorm:"not null comment('主币名称') VARCHAR(15)"`
	CoinType      string `json:"coin_type" xorm:"not null comment('币种名称') unique(transaction_id) VARCHAR(15)"`
	BlockHeight   int    `json:"block_height" xorm:"not null default 0 INT(11)"`
	Hash          string `json:"hash" xorm:"not null comment('块hash') unique(transaction_id) VARCHAR(100)"`
	Timestamp     int    `json:"timestamp" xorm:"not null default 0 comment('交易时间戳') INT(11)"`
	TxId          string `json:"tx_id" xorm:"not null comment('txid') unique(transaction_id) VARCHAR(150)"`
	TxInAmount    string `json:"tx_in_amount" xorm:"not null default 0.000000000000000000 comment('交易from总额') DECIMAL(60,24)"`
	TxInNum       int    `json:"tx_in_num" xorm:"not null default 0 comment('交易from笔数') INT(11)"`
	TxOutAmount   string `json:"tx_out_amount" xorm:"not null default 0.000000000000000000 comment('交易out总额') DECIMAL(60,24)"`
	TxOutNum      int    `json:"tx_out_num" xorm:"not null default 0 comment('交易to笔数') INT(11)"`
	TxFee         string `json:"tx_fee" xorm:"not null default 0.000000000000000000 DECIMAL(60,24)"`
	TxFeeCoin     string `json:"tx_fee_coin" xorm:"not null comment('矿工费币种') VARCHAR(15)"`
	TxN           int    `json:"tx_n" xorm:"default 0 comment('多交易下标') INT(10)"`
	Memo          string `json:"memo" xorm:"comment('链上交易备注') VARCHAR(255)"`
	Confirmations int    `json:"confirmations" xorm:"not null default 0 comment('确认数') INT(10)"`
	Status        int    `json:"status" xorm:"not null default 1 comment('交易状态 1 正常  2 作废  3 虚拟补单') TINYINT(3)"`
	TransStatus   string `json:"trans_status" xorm:"VARCHAR(50)"`
	IsDump        int    `json:"is_dump" xorm:"not null default 0 comment('是否已整理') TINYINT(3)"`
	Remark        string `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	CreateAt      int    `json:"create_at" xorm:"not null default 0 INT(11)"`
	//UpdateAt      time.Time 	`json:"update_at" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	ContrastTime int `json:"contrast_time" xorm:"not null default 0 comment('对账时间') INT(11)"`
}
