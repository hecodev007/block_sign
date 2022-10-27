package bo

import "github.com/group-coldwallet/bchserver/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
