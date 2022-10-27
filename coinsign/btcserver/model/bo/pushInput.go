package bo

import "github.com/group-coldwallet/btcserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
