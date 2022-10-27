package utils

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/complex/types"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/dto"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers"
)

type Utils struct {
	provider providers.ProviderInterface
}

func NewUtils(provider providers.ProviderInterface) *Utils {
	utils := new(Utils)
	utils.provider = provider
	return utils
}

func (utils *Utils) Sha3(data types.ComplexString) (string, error) {

	params := make([]string, 1)
	params[0] = data.ToHex()

	pointer := &dto.RequestResult{}

	err := utils.provider.SendRequest(pointer, "web3_sha3", params)

	if err != nil {
		return "", err
	}

	return pointer.ToString()

}
