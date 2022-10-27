package dot

type BlockTxAbnomal struct {
	Id              int64  `xorm:"pk autoincr BIGINT(20)"`
	Txid            string `xorm:"not null default '' comment('交易id') index unique(idx_txid_height) VARCHAR(100)"`
	Height          int64  `xorm:"not null default 0 comment('区块高度索引值') index unique(idx_txid_height) BIGINT(20)"`
	Hash            string `xorm:"not null default '' comment('区块hash值') index VARCHAR(100)"`
	SysFee          string `xorm:"not null default 0.000000000000000000 comment('手续费') DECIMAL(40,18)"`
	Fromaccount     string `xorm:"not null default '' comment('from') VARCHAR(100)"`
	Toaccount       string `xorm:"not null default '' comment('to') VARCHAR(100)"`
	Amount          string `xorm:"not null default 0.000000000000000000 comment('金额') DECIMAL(40,18)"`
	Memo            string `xorm:"not null default '' comment('备注') VARCHAR(255)"`
	Contractaddress string `xorm:"not null default '' comment('合约地址') VARCHAR(255)"`
	SucInfo         string `xorm:"TEXT"`
}
