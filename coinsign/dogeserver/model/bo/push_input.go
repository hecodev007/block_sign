package bo

import "github.com/group-coldwallet/dogeserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
