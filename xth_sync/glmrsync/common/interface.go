package common

import (
	"glmrsync/models/po"
)

type ProcTask interface {
	GetIrreversible() bool
	GetHeight() int64
	GetBestHeight() int64
	GetConfirms() int64
	GetBlockHash() string
	GetTxs() []interface{}
	GetBlock() interface{}
}

//type PushTask interface {
//	GetHeight() int64
//	GetTxid() string
//	GetWatchAddrs() []string
//	GetData() []byte
//}

type BaseService interface {
	Start()
	Stop()
}

type Pusher interface {
	AddPushTask(height int64, txid string, watchlist map[string]bool, pushdata []byte)
}

type Processor interface {
	Init() error
	Clear()
	SetPusher(p Pusher)
	RemovePusher()
	RepushTx(userId int64, txid string) error
	Info() (string, int64, error)
	//暂没用到 ，查询数据是否已有这个区块
	CheckIrreverseBlock(hash string) error
	//处理不可逆区块交易(推送交易，保存到数据库)
	ProcIrreverseTask(ProcTask) error
	//推送不可逆交易确认数（） 待定
	//UpdateIrreverseConfirms(ProcTask)
	//处理不可逆区块（保存到数据库）
	//ProcIrreverseBlock(ProcTask) error
	//处理可逆区块交易(是否需要推送)；有需要推送？，error
	ProcReverseTxs(ProcTask) (bool, error)
	//推送可逆区块确认数(不需要保存数据库，直接推送数据)
	PushReverseConfirms(ProcTask)
}

type Scanner interface {
	Init() error
	Clear()
	Rollback(height int64)
	//获取当前最高高度
	GetBestBlockHeight() (int64, error)
	GetCurrentBlockHeight() (int64, error)
	ScanReverseBlock(height, bestHeight int64) (ProcTask, error)
	ScanIrreverseBlock(height, bestHeight int64) (ProcTask, error)
}

type NotifyResultDB interface {
	SelectNotifyResult(id int64) (*po.NotifyResult, error)
	InsertNotifyResult(n *po.NotifyResult) (int64, error)
	UpdateNotifyResult(n *po.NotifyResult) error
	SelectWatchHeight(height int64) (map[int64]int, error)
}
