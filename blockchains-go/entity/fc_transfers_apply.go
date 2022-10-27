package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
	"xorm.io/builder"
)

type TxType int

const (
	SingleAddrTx TxType = 1
	MultiAddrTx  TxType = 2
)

type ApplyType string

const (
	CZ_ApplyType   ApplyType = "cz"
	DS_ApplyType   ApplyType = "ds"
	FEE_ApplyType  ApplyType = "fee"
	GJ_ApplyType   ApplyType = "gj"
	BUY_ApplyType  ApplyType = "buy"
	SELL_ApplyType ApplyType = "sell"
	HB_ApplyType   ApplyType = "hb"
	BD_ApplyType   ApplyType = "bd"
)

var ApplyTypeDesc = map[ApplyType]string{
	CZ_ApplyType:   "出账",
	DS_ApplyType:   "打散UTXO",
	FEE_ApplyType:  "打手续费",
	GJ_ApplyType:   "归集",
	BUY_ApplyType:  "买内存",
	SELL_ApplyType: "卖内存",
	HB_ApplyType:   "合并",
	BD_ApplyType:   "补单",
}

//fc_transfer_apply status状态
type ApplyStatus int

const (
	ApplyStatus_Ignore    ApplyStatus = -41 //预防与php冲突
	ApplyStatus_Auditing  ApplyStatus = 40  //预防与php冲突
	ApplyStatus_AuditOk   ApplyStatus = 41  //预防与php冲突
	ApplyStatus_AuditFail ApplyStatus = 42  //预防与php冲突
	ApplyStatus_Creating  ApplyStatus = 43  //预防与php冲突
	ApplyStatus_CreateOk  ApplyStatus = 47  //预防与php冲突
	//ApplyStatus_CreateRetry ApplyStatus = 48  //预防与php冲突
	ApplyStatus_CreateFail ApplyStatus = 49 //预防与php冲突
	ApplyStatus_Merge      ApplyStatus = 50 //预防与php冲突
	ApplyStatus_Fee        ApplyStatus = 51 //预防与php冲突
	ApplyStatus_Rollback   ApplyStatus = 52 //标记为回滚
	ApplyStatus_TransferOk ApplyStatus = 30 //预防与php冲突 先从30开始

)

var ApplyStatusDesc = map[ApplyStatus]string{
	ApplyStatus_Ignore:    "忽略",
	ApplyStatus_Auditing:  "未审核",
	ApplyStatus_AuditOk:   "审核通过",
	ApplyStatus_AuditFail: "审核驳回",
	ApplyStatus_Creating:  "构建中",
	ApplyStatus_CreateOk:  "构建成功",
	//ApplyStatus_CreateRetry: "构建失败，等待重试",
	ApplyStatus_CreateFail: "构建失败",
	ApplyStatus_Merge:      "归集构建",
	ApplyStatus_Fee:        "打手续费构建",
	ApplyStatus_TransferOk: "交易成功",
}

type FcTransfersApply struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Username   string    `json:"username" xorm:"not null default '' comment('登录账号') VARCHAR(32)"`
	Department string    `json:"department" xorm:"not null default '' comment('申请部门') VARCHAR(255)"`
	Applicant  string    `json:"applicant" xorm:"not null default '' comment('申请人') VARCHAR(255)"`
	OutOrderid string    `json:"out_orderid" xorm:"not null default '' comment('合作方订单ID') unique VARCHAR(64)"`
	OrderId    string    `json:"order_id" xorm:"comment('内部订单id') VARCHAR(255)"`
	Operator   string    `json:"operator" xorm:"not null default '' comment('操作人') VARCHAR(255)"`
	CoinName   string    `json:"coin_name" xorm:"VARCHAR(15)"`
	Type       string    `json:"type" xorm:"not null comment('转手续费,出账,归集,买内存,卖内存,打散utxo,补单,公链云出账,合并utxo') ENUM('bd','buy','cz','ds','fee','gj','gly','hb','sell')"`
	Purpose    string    `json:"purpose" xorm:"not null default '' comment('出库用途') VARCHAR(255)"`
	Status     int       `json:"status" xorm:"not null default 0 comment('0未审核1审核通过2审核驳回3构造中4构造成功5构造失败6构造失败待重试10 未审核11 等待创建订单12 创建订单完成13 等待构建14 构建成功15 广播成功16 构建失败17 签名失败18 广播失败19 签名超时20 异常错误') TINYINT(3)"`
	CallBack   string    `json:"call_back" xorm:"not null default '' comment('回调url') VARCHAR(255)"`
	ErrorNum   int       `json:"error_num" xorm:"default 0 comment('错误次数') INT(11)"`
	Createtime int64     `json:"createtime" xorm:"not null default 0 comment('申请时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	Memo       string    `json:"memo" xorm:"not null default '' comment('eos memo') VARCHAR(100)"`
	Fee        string    `json:"fee" xorm:"default 0.00000000000000000000 comment('手续费') DECIMAL(50,20)"`
	Eostoken   string    `json:"eostoken" xorm:"not null default '' comment('eostoken') VARCHAR(50)"`
	Eoskey     string    `json:"eoskey" xorm:"not null default _'' comment('eoskey') VARCHAR(50)"`
	AppId      int       `json:"app_id" xorm:"not null default 0 comment('商户id') INT(11)"`
	IsDing     int       `json:"is_ding" xorm:"not null default 0 TINYINT(1)"`
	Remark     string    `json:"remark" xorm:"comment('忽略订单原因') TEXT"`
	Source     int       `json:"source" xorm:"default 1 comment('1.API订单2管理员订单3手动订单') TINYINT(4)"`
	Code       string    `json:"code" xorm:"comment('订单表，订单code') VARCHAR(100)"`
	IsExamine  int       `json:"is_examine" xorm:"not null default 0 comment('0未人工审核1已人工审核') TINYINT(4)"`
	Isforce    int       `json:"isForce" xorm:"not null default 0 comment('0不强制修正手续费1强制修正手续费') TINYINT(4)"`
	Sort       int       `xorm:"'sort'"`
	TxType     TxType    `json:"tx_type" xorm:"default 1 comment('出账类型；1：单地址出账；2：多地址出账') INT(11)"`
}

func (o *FcTransfersApply) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
func (o *FcTransfersApply) Get(cond builder.Cond) (bool, error) {
	return db.Conn.Where(cond).Desc("id").Get(o)
}
func (o FcTransfersApply) Update(cond builder.Cond) (int64, error) {
	return db.Conn.Where(cond).Update(o)
}

func (o FcTransfersApply) Find(cond builder.Cond, limit int) ([]*FcTransfersApply, error) {
	res := make([]*FcTransfersApply, 0)
	if err := db.Conn.Where(cond).Limit(limit).Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
func (o *FcTransfersApply) TransactionAdd(tacs []*FcTransfersApplyCoinAddress) (int64, error) {
	session := db.Conn.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return -1, err
	}
	_, err = session.InsertOne(o)
	if err != nil {
		session.Rollback()
		return -1, err
	}
	appId := int64(o.Id)
	if len(tacs) > 0 {
		for _, tac := range tacs {
			tac.ApplyId = appId
		}
		_, err = session.Insert(tacs)
		if err != nil {
			session.Rollback()
			return -1, err
		}
	}
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return -1, err
	}
	return appId, nil
}
