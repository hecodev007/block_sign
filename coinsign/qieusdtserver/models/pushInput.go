package models

type PushInput struct {
	Complete bool        `json:"complete,omitempty"`
	Hex      string      `json:"hex,omitempty"`
	Error    interface{} `json:"error,omitempty"`
	MchInfo
}
