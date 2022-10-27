package testutil

import (
	"btmSign/bytom/protocol/bc"
)

var (
	MaxHash = &bc.Hash{V0: 1<<64 - 1, V1: 1<<64 - 1, V2: 1<<64 - 1, V3: 1<<64 - 1}
	MinHash = &bc.Hash{}
)
