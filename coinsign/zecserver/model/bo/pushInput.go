package bo

import "github.com/group-coldwallet/zecserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
