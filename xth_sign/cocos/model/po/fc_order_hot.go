package po

import (
	"github.com/sirupsen/logrus"
	"github.com/group-coldwallet/cocos/launcher"
	"time"
)

type HotOrder struct {
	Id           int64       `json:"id,omitempty" gorm:"column:id"`
	ApplyId      int64       `json:"apply_id" gorm:"column:apply_id"`                     //'申请id'
	ApplyCoinId  int64       `json:"apply_coin_id,omitempty" gorm:"column:apply_coin_id"` //'申请币种信息id'
	OuterOrderNo string      `json:"outer_order_no" gorm:"column:outer_order_no"`         //'外部订单号'
	OrderNo      string      `json:"order_no" gorm:"column:order_no"`                     //'交易订单号'
	MchName      string      `json:"mch_name" gorm:"column:mch_name"`                     //'商户名'
	CoinName     string      `json:"coin_name" gorm:"column:coin_name"`                   //'币种'
	Token        string      `json:"token,omitempty" gorm:"column:token"`                 //eth 就是 erc20 ,eos 就是 token
	FromAddress  string      `json:"from_address,omitempty" gorm:"column:from_address"`   //'发送者地址,对于多个in多个out情况，要看fc_order_addres
	ToAddress    string      `json:"to_address,omitempty" gorm:"column:to_address"`       // '接收者地址'
	Amount       int64       `json:"amount,omitempty" gorm:"column:amount"`               //'金额 '
	Quantity     string      `json:"quantity,omitempty" gorm:"column:quantity"`           //金额是字符串就需要填写这个属性
	Memo         string      `json:"memo,omitempty" gorm:"column:memo"`
	Fee          int64       `json:"fee" gorm:"column:fee"`
	Decimal      int         `json:"decimal,omitempty" gorm:"column:decimal"`
	CreateData   string      `json:"create_data,omitempty" gorm:"column:create_data"` //'创建构造交易内容'
	ErrorMsg     string      `json:"error_msg,omitempty" gorm:"column:error_msg"`     //'构造错误信息'
	ErrorCount   int         `json:"error_count,omitempty" gorm:"column:error_count"` //'构造失败次数'
	Status       OrderStatus `json:"status" gorm:"column:status;default:0"`           //'0:构建完成,1:推入队列,2:已拉取,3:已签名,4:已广播,5:构建失败,6:签名失败7:广播失败'
	IsRetry      bool        `json:"is_retry,omitempty" gorm:"column:is_retry"`       //'0非重试1重试'
	TxId         string      `json:"tx_id,omitempty" gorm:"column:tx_id"`
	CreateAt     int64       `json:"create_at,omitempty" gorm:"column:create_at"`
	UpdateAt     int64       `json:"update_at,omitempty" gorm:"column:update_at"`
	Worker       string      `json:"worker,omitempty" gorm:"column:worker"`
}

func (o *HotOrder) TableName() string {
	return "fc_order_hot"
}

type OrderStatus int

const (
	All                  OrderStatus = -1
	CreateStatus         OrderStatus = 0 //构建完成
	PushStatus           OrderStatus = 1 //推入队列
	PullStatus           OrderStatus = 2 //已拉取
	SignStatus           OrderStatus = 3 //已签名
	BroadcastStatus      OrderStatus = 4 //广播成功
	CreateErrorStatus    OrderStatus = 5 //构建失败
	SignErrorStatus      OrderStatus = 6 //签名失败
	BroadcastErrorStatus OrderStatus = 7 //广播失败  //热钱包 交易失败
	PendingTimeoutStatus OrderStatus = 8 //pending超时
	UnknowErrorStatus    OrderStatus = 9 //未知异常  热钱包 交易失败
)

var StatusDesc = map[OrderStatus]string{
	All:                  " ",
	CreateStatus:         "构建完成",
	PushStatus:           "推入队列",
	PullStatus:           "已拉取",
	SignStatus:           "已签名",
	BroadcastStatus:      "广播成功",
	CreateErrorStatus:    "构建失败",
	SignErrorStatus:      "签名失败",
	BroadcastErrorStatus: "广播失败",
	PendingTimeoutStatus: "pending超时",
	UnknowErrorStatus:    "未知异常",
}

func GetStatusDesc(code OrderStatus) string {
	msg, ok := StatusDesc[code]
	if ok {
		return msg
	}
	return "未知状态"
}

//插入一条数据
//返回订单ID
func (o *HotOrder) InsertOrder() (int64, error) {
	nowTime := time.Now().Unix()
	o.CreateAt = nowTime
	o.UpdateAt = nowTime
	if err := launcher.MysqlDB.Create(o).Error; err != nil {
		return 0, err
	}
	return o.Id, nil
}

//更新错误状态数据
func (o *HotOrder) UpdateOrderErrorStatus(orderNo string, status OrderStatus, errorStr string) error {
	nowTime := time.Now().Unix()
	ds := make(map[string]interface{})
	ds["status"] = status
	ds["update_at"] = nowTime
	if errorStr != "" {
		ds["error_msg"] = errorStr
	}
	err := launcher.MysqlDB.Model(o).Where("order_no = ?", orderNo).Update(ds).Error
	return err
}

//更新创建交易后数据
func (o *HotOrder) UpdateOrderCreateStatus(orderNo string, createData string) error {
	nowTime := time.Now().Unix()
	ds := make(map[string]interface{})
	ds["status"] = CreateStatus
	ds["create_data"] = createData
	ds["update_at"] = nowTime
	err := launcher.MysqlDB.Model(o).Where("order_no = ?", orderNo).Update(ds).Error
	return err
}

//更新签名后的数据
func (o *HotOrder) UpdateOrderSignStatus(orderNo string, createData string) error {
	nowTime := time.Now().Unix()
	ds := make(map[string]interface{})
	ds["status"] = SignStatus
	ds["create_data"] = createData
	ds["update_at"] = nowTime
	err := launcher.MysqlDB.Model(o).Where("order_no = ?", orderNo).Update(ds).Error
	return err
}

//更新广播后的数据
func (o *HotOrder) UpdateOrderBroadcastStatus(orderNo string, txid string) error {
	nowTime := time.Now().Unix()
	ds := make(map[string]interface{})
	ds["status"] = BroadcastStatus
	ds["tx_id"] = txid
	ds["update_at"] = nowTime
	logrus.Infof("%+v", ds)
	err := launcher.MysqlDB.Model(o).Where("order_no = ?", orderNo).Update(ds).Error
	return err
}
