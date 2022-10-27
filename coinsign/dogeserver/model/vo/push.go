package vo

import "github.com/group-coldwallet/dogeserver/model"

type PushResult struct {
	TxID string `json:"txid"` //事务ID
	model.MchInfo
}
