package vo

import "github.com/group-coldwallet/ucaserver/model"

type PushResult struct {
	TxID string `json:"txid"` //事务ID
	model.MchInfo
}
