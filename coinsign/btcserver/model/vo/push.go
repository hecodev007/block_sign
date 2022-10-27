package vo

import "github.com/group-coldwallet/btcserver/model"

type PushResult struct {
	TxID string `json:"txid"` //事务ID
	model.MchInfo
}
