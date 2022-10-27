package account

import (
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/dto"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers"
)

type Personal struct {
	provider providers.ProviderInterface
}

func NewPersonal(provider providers.ProviderInterface) *Personal {
	personal := new(Personal)
	personal.provider = provider
	return personal
}

func (personal *Personal) ListAccounts() ([]string, error) {

	pointer := &dto.RequestResult{}

	err := personal.provider.SendRequest(pointer, "personal_listAccounts", nil)

	if err != nil {
		return nil, err
	}

	return pointer.ToStringArray()

}

func (personal *Personal) NewAccount(password string) (string, error) {

	params := make([]string, 1)
	params[0] = password

	pointer := &dto.RequestResult{}

	err := personal.provider.SendRequest(&pointer, "personal_newAccount", params)

	if err != nil {
		return "", err
	}

	response, err := pointer.ToString()

	return response, err

}

func (personal *Personal) SendTransaction(transaction *dto.TransactionParameters, password string) (string, error) {

	params := make([]interface{}, 2)

	transactionParameters := transaction.Transform()

	params[0] = transactionParameters
	params[1] = password

	pointer := &dto.RequestResult{}

	err := personal.provider.SendRequest(pointer, "personal_sendTransaction", params)

	if err != nil {
		return "", err
	}

	return pointer.ToString()

}

func (personal *Personal) UnlockAccount(address string, password string, duration uint64) (bool, error) {

	params := make([]interface{}, 3)
	params[0] = address
	params[1] = password
	params[2] = duration

	pointer := &dto.RequestResult{}

	err := personal.provider.SendRequest(pointer, "personal_unlockAccount", params)

	if err != nil {
		return false, err
	}

	return pointer.ToBoolean()

}
