package util

type JsonParam struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}
