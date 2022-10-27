package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/builder"
)

type FcTransPush struct {
	Id            int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinType      string `json:"coin_type" xorm:"not null comment('币种') unique(txid_hash) VARCHAR(20)"`
	IsIn          int    `json:"is_in" xorm:"not null default 0 comment('是否转入 1 转入 2 转出') unique(txid_hash) TINYINT(1)"`
	BlockHeight   int64  `json:"block_height" xorm:"not null default 0 comment('块高度') BIGINT(20)"`
	Hash          string `json:"hash" xorm:"not null comment('块hash') VARCHAR(100)"`
	Timestamp     int    `json:"timestamp" xorm:"not null default 0 comment('块时间') INT(11)"`
	TransactionId string `json:"transaction_id" xorm:"not null comment('交易ID(唯一键)') unique(txid_hash) VARCHAR(150)"`
	MuxId         string `json:"mux_id" xorm:"not null comment('btm签名用') VARCHAR(150)"`
	TrxN          int    `json:"trx_n" xorm:"not null default 0 comment('交易序号') unique(txid_hash) INT(10)"`
	FromTrxId     string `json:"from_trx_id" xorm:"not null unique(txid_hash) VARCHAR(150)"`
	VoutId        string `json:"vout_id" xorm:"not null comment('唯一标识btm utxo') VARCHAR(150)"`
	FromAddress   string `json:"from_address" xorm:"not null comment('来源地址') VARCHAR(255)"`
	ToAddress     string `json:"to_address" xorm:"not null comment('目的地址') VARCHAR(255)"`
	Address       string `json:"address" xorm:"not null index unique(txid_hash) VARCHAR(190)"`
	Memo          string `json:"memo" xorm:"VARCHAR(60)"`
	Amount        string `json:"amount" xorm:"not null default 0.0000000000 unique(txid_hash) DECIMAL(40,10)"`
	Fee           string `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(40,10)"`
	Confirmations int    `json:"confirmations" xorm:"not null default 0 comment('确认数') SMALLINT(6)"`
	IsSpent       int    `json:"is_spent" xorm:"not null default 0 comment('是否花费 1 已花费 2 冻结') TINYINT(3)"`
	UserSubId     int    `json:"user_sub_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	AppId         int    `json:"app_id" xorm:"not null default 0 comment('商户ID') index SMALLINT(6)"`
	Type          int    `json:"type" xorm:"not null default 0 comment('地址类型 0 外部地址 1 归集地址（冷地址）  2 用户地址  3 手续费地址  4 热地址') index TINYINT(3)"`
	TradeType     int    `json:"trade_type" xorm:"not null default 0 comment('交易类型  1 出账  2 入账  3 归集  ') TINYINT(3)"`
	PushState     int    `json:"push_state" xorm:"not null default 0 comment('推送状态 0 未推送 1 已推送 2 推送失败 3 正在推送') TINYINT(3)"`
	OrderNo       string `json:"order_no" xorm:"comment('冻结与解冻utxo的订单') VARCHAR(50)"`
}

// unspents切片排序
type DBUnspentDesc []FcTransPush

//实现排序三个接口
//为集合内元素的总数
func (s DBUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s DBUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s DBUnspentDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

// unspents切片排序
type DBUnspentAsc []FcTransPush

//实现排序三个接口
//为集合内元素的总数
func (s DBUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s DBUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s DBUnspentAsc) Less(i, j int) bool {
	return s[i].Amount < s[j].Amount
}

func (f FcTransPush) FindTransPush(cond builder.Cond) ([]*FcTransPush, error) {
	res := make([]*FcTransPush, 0)
	if err := db.Conn.Where(cond).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
