package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"xorm.io/builder"
)

type OrderStatus int

//const (
//	All                  OrderStatus = -1
//	CreateStatus         OrderStatus = 0 //构建完成
//	PushStatus           OrderStatus = 1 //推入队列,签名中
//	SignStatus           OrderStatus = 2 //已签名成功
//	BroadcastingStatus   OrderStatus = 3 //正在广播中
//	BroadcastStatus      OrderStatus = 4 //广播成功
//	CreateErrorStatus    OrderStatus = 5 //构建失败
//	SignErrorStatus      OrderStatus = 6 //签名失败
//	BroadcastErrorStatus OrderStatus = 7 //广播失败
//	PendingTimeoutStatus OrderStatus = 8 //pending超时
//	UnknowErrorStatus    OrderStatus = 9 //未知异常
//)
//
//var StatusDesc = map[OrderStatus]string{
//	All:                  " ",
//	CreateStatus:         "构建完成",
//	PushStatus:           "推入队列,签名中",
//	SignStatus:           "已签名成功",
//	BroadcastingStatus:   "正在广播中",
//	BroadcastStatus:      "广播成功",
//	CreateErrorStatus:    "构建失败",
//	SignErrorStatus:      "签名失败",
//	BroadcastErrorStatus: "广播失败",
//	PendingTimeoutStatus: "pending超时",
//	UnknowErrorStatus:    "未知异常",
//}

type FcOrder struct {
	Id           int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	ApplyId      int    `json:"apply_id" xorm:"default 0 comment('申请id') index(apply_id) INT(11)"`
	ApplyCoinId  int    `json:"apply_coin_id" xorm:"default 0 comment('申请币种信息id') index(apply_id) INT(11)"`
	OuterOrderNo string `json:"outer_order_no" xorm:"default '' comment('外部订单号') index VARCHAR(255)"`
	OrderNo      string `json:"order_no" xorm:"default '' comment('交易订单号') unique VARCHAR(255)"`
	MchName      string `json:"mch_name" xorm:"default '' comment('商户名') VARCHAR(30)"`
	CoinName     string `json:"coin_name" xorm:"default '' comment('币种') VARCHAR(15)"`
	FromAddress  string `json:"from_address" xorm:"default '' comment('发送者地址,对于多个in多个out情况，要看fc_order_address表') VARCHAR(255)"`
	ToAddress    string `json:"to_address" xorm:"default '' comment('接收者地址') VARCHAR(255)"`
	Token        string `json:"token" xorm:"default '' VARCHAR(40)"`
	Amount       string `json:"amount" xorm:"default 0 comment('金额 ') DECIMAL(60)"`
	Quantity     string `json:"quantity" xorm:"default '' VARCHAR(60)"`
	Memo         string `json:"memo" xorm:"default '' VARCHAR(255)"`
	Fee          string `json:"fee" xorm:"default 0 comment('utxo手续费') DECIMAL(60)"`
	CreateData   string `json:"create_data" xorm:"comment('创建构造交易内容') TEXT"`
	ErrorMsg     string `json:"error_msg" xorm:"default '' comment('构造错误信息') VARCHAR(500)"`
	ErrorCount   int    `json:"error_count" xorm:"default 0 comment('构造失败次数') INT(11)"`
	Status       int    `json:"status" xorm:"default 0 comment('0:构建完成,1:推入队列,2:已拉取,3:已签名,4:已广播,5:构建失败,6:签名失败7:广播失败 8:超时') TINYINT(1)"`
	IsRetry      int    `json:"is_retry" xorm:"default 0 comment('0非重试1重试') TINYINT(1)"`
	TxId         string `json:"tx_id" xorm:"default '' comment('区块链交易id') VARCHAR(150)"`
	CreateAt     int64  `json:"create_at" xorm:"default 0 BIGINT(11)"`
	UpdateAt     int64  `json:"update_at" xorm:"default 0 comment('最后修改时间') BIGINT(11)"`
	Decimal      int64  `json:"decimal" xorm:"default 0 comment('以太坊单位精度数') BIGINT(5)"`
	Worker       string `json:"worker" xorm:"default '' comment('工作的机器') VARCHAR(50)"`
	MemoEncrypt  string `xorm:"'memo_encrypt'"` //memo加密后信息 应对浏览器信息只显示加密后的信息的币种
	TxType       int    `json:"tx_type" xorm:"default 1 comment('出账类型；1：单地址出账；2：多地址出账') INT(11)"`
	TotalAmount  string `json:"total_amount" xorm:"comment('订单总金额') DECIMAL(60,24)"`
}

func (o *FcOrder) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
func (o *FcOrder) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcOrder) Update(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Update(o)
}
func (o FcOrder) Exist(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Exist(o)
}
func (o FcOrder) Find(cond builder.Cond) ([]*FcOrder, error) {
	res := make([]*FcOrder, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o FcOrder) FindOrders(cond builder.Cond, limit int) ([]*FcOrder, error) {
	res := make([]*FcOrder, 0)
	if err := db.Conn.Where(cond).Desc("id").Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
