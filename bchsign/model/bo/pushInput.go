package bo

import "github.com/group-coldwallet/bchsign/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
