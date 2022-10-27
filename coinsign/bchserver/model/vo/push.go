package vo

import "github.com/group-coldwallet/bchserver/model"

type PushResult struct {
	TxID string `json:"txid"` //事务ID
	model.MchInfo
}
