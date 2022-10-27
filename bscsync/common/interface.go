package common

import (
	"dataserver/models/po"
	"sync"
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

// type PushTask interface {
// 	GetHeight() int64
// 	GetTxid() string
// 	GetWatchAddrs() []string
// 	GetData() []byte
// }

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

	CheckIrreverseBlock(hash string) error
	ProcIrreverseTxs([]interface{}, int64) error
	ProcReverseTxs([]interface{}, int64) error
	ProcIrreverseBlock(interface{}) error
	UpdateIrreverseConfirms()
	UpdateReverseConfirms(interface{})
}

type Scanner interface {
	Init() error
	Clear()
	Rollback(height int64)
	GetBestBlockHeight() (int64, error)
	GetCurrentBlockHeight() (int64, error)
	BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map
	ScanReverseBlock(height, bestHeight int64) (ProcTask, error)
	ScanIrreverseBlock(height, bestHeight int64) (ProcTask, error)
}

type NotifyResultDB interface {
	SelectNotifyResult(id int64) (*po.NotifyResult, error)
	InsertNotifyResult(n *po.NotifyResult) (int64, error)
	UpdateNotifyResult(n *po.NotifyResult) error
	SelectWatchHeight(height int64) (map[int64]int, error)
}
