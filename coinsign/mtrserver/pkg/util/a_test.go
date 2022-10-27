package util

import (
	"encoding/json"
	"testing"
)

func TestHttpGet(t *testing.T) {
	bb, err := HttpGet("http://mtr.rylink.io:30869/blocks/best")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(bb))

	var block map[string]interface{}
	json.Unmarshal(bb, &block)
	t.Log(block["number"])
	t.Log(block["number"] == nil)
	t.Log(block["number"].(float64) <= 0)
	if block["number"] == nil || block["number"].(float64) <= 0 {
		t.Log(111111)

	}

	bb, err = HttpGet("http://mtr.rylink.io:30869/accounts/0x79c77f43ff0b291c1ae5d5e2aa1143949e4366fb")
	balance := make(map[string]interface{}, 0)
}
