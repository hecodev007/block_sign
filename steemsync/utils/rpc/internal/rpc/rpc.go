package rpc

import (
	// Stdlib
	"encoding/json"

	// RPC
	"steemsync/utils/rpc/interfaces"

	// Vendor
	"github.com/pkg/errors"
)

func GetNumericAPIID(caller interfaces.Caller, apiName string) (int, error) {
	params := []interface{}{apiName}

	var resp json.RawMessage
	//if err := caller.Call("call", []interface{}{1, "get_api_by_name", params}, &resp); err != nil {
	//	return 0, err
	//}
	if err := caller.Call("get_api_by_name", params, &resp); err != nil {
		return 0, err
	}
	if string(resp) == "null" {
		return 0, errors.Errorf("API not available: %v", apiName)
	}

	var id int
	if err := json.Unmarshal([]byte(resp), &id); err != nil {
		return 0, err
	}
	return id, nil
}
