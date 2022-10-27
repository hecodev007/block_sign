package redis

import (
	"fmt"
)

const (
	BroadcastOuterOrderNoKey = "broadcast_order"
)

func GetBroadcastOuterOrderNoKey(outerOrderNo string) string {
	return fmt.Sprintf("%s_%s", BroadcastOuterOrderNoKey, outerOrderNo)
}
