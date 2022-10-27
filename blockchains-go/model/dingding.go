package model

type ActionType string

const ()

//钉钉Outgoingj机制动作设置
type DingOutgoing struct {
	Action ActionType `json:"action"`
}
