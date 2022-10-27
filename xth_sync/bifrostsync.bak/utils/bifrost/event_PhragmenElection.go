package bifrost

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

//PhragmenElection

type EventPhragmenElection struct {
	Phase types.Phase
	Topics []types.Hash
}
