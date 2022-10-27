package bo

import "github.com/group-coldwallet/btcsign/model"

type PushInput struct {
	Hex string `json:"hex"`
	model.MchInfo
}
