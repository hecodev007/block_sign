package vo

import "github.com/group-coldwallet/btcsign/model"

type PushResult struct {
	TxID string `json:"txid"` //事务ID
	model.MchInfo
}
