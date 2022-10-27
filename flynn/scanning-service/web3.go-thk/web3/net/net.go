package net

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/dto"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers"
	"math/big"
)

type Net struct {
	provider providers.ProviderInterface
}

func NewNet(provider providers.ProviderInterface) *Net {
	net := new(Net)
	net.provider = provider
	return net
}

func (net *Net) IsListening() (bool, error) {

	pointer := &dto.RequestResult{}

	err := net.provider.SendRequest(pointer, "net_listening", nil)

	if err != nil {
		return false, err
	}

	return pointer.ToBoolean()

}

func (net *Net) GetPeerCount() (*big.Int, error) {

	pointer := &dto.RequestResult{}

	err := net.provider.SendRequest(pointer, "net_peerCount", nil)

	if err != nil {
		return nil, err
	}

	return pointer.ToBigInt()

}

func (net *Net) GetVersion() (string, error) {

	pointer := &dto.RequestResult{}

	err := net.provider.SendRequest(pointer, "net_version", nil)

	if err != nil {
		return "", err
	}

	return pointer.ToString()

}
