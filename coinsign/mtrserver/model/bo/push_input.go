package bo

import "github.com/group-coldwallet/mtrserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
