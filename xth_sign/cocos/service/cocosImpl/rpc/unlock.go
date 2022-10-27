package rpc

import (
	"fmt"
)

func Unlock(password string) string {
	return fmt.Sprintf(`{"jsonrpc": "2.0", "id":"2", "method": "unlock", "params": ["%s"] }`, password)
}
