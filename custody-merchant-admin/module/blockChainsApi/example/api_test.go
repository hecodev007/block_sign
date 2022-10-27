package example

import (
	"custody-merchant-admin/module/blockChainsApi"
	"testing"
)

func TestApi1(t *testing.T) {

	blockChainsApi.ValidInsideAddress("d28fa2b0-d36a-4b5f-a7ff-0612bdc620d7",
		"d28fa2b0-d36a-4b5f-a7ff-0612bdc620d7",
		"31ywhtAGwh74ThyfnGHj788aVWhbViKhpZ")
}
