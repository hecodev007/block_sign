package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleEstimateNetworkHashesPerSecond handles the respectively named RPC command
func HandleEstimateNetworkHashesPerSecond(
	context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	estimateNetworkHashesPerSecondRequest := request.(*appmessage.EstimateNetworkHashesPerSecondRequestMessage)

	windowSize := int(estimateNetworkHashesPerSecondRequest.WindowSize)
	startHash := model.VirtualBlockHash
	if estimateNetworkHashesPerSecondRequest.StartHash != "" {
		var err error
		startHash, err = externalapi.NewDomainHashFromString(estimateNetworkHashesPerSecondRequest.StartHash)
		if err != nil {
			response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
			response.Error = appmessage.RPCErrorf("StartHash '%s' is not a valid block hash",
				estimateNetworkHashesPerSecondRequest.StartHash)
			return response, nil
		}
	}

	networkHashesPerSecond, err := context.Domain.Consensus().EstimateNetworkHashesPerSecond(startHash, windowSize)
	if err != nil {
		response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
		response.Error = appmessage.RPCErrorf("could not resolve network hashes per "+
			"second for startHash %s and window size %d: %s", startHash, windowSize, err)
		return response, nil
	}

	return appmessage.NewEstimateNetworkHashesPerSecondResponseMessage(networkHashesPerSecond), nil
}
