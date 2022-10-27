package bo

import "github.com/group-coldwallet/ltcserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
