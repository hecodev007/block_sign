package thk

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/common/hexutil"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/dto"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/thk/abi"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/thk/util"
	"math/big"
	"strings"
)

type Contract struct {
	super     *Thk
	abi       abi.ABI
	functions map[string][]string
}

//新合约
func (thk *Thk) NewContract(abistr string) (*Contract, error) {

	contract := new(Contract)
	var mockInterface interface{}

	err := json.Unmarshal([]byte(abistr), &mockInterface)

	if err != nil {
		return nil, err
	}

	jsonInterface := mockInterface.([]interface{})
	contract.functions = make(map[string][]string)
	for index := 0; index < len(jsonInterface); index++ {
		function := jsonInterface[index].(map[string]interface{})

		if function["type"] == "constructor" || function["type"] == "fallback" {
			function["name"] = function["type"]
		}

		functionName := function["name"].(string)
		contract.functions[functionName] = make([]string, 0)

		if function["inputs"] == nil {
			continue
		}

		inputs := function["inputs"].([]interface{})
		for paramIndex := 0; paramIndex < len(inputs); paramIndex++ {
			params := inputs[paramIndex].(map[string]interface{})
			contract.functions[functionName] = append(contract.functions[functionName], params["type"].(string))
		}

	}
	readerstr := strings.NewReader(abistr)
	Abi, err := abi.JSON(readerstr)
	if err != nil {
		return nil, err
	}
	contract.abi = Abi
	contract.super = thk

	return contract, nil
}

func (contract *Contract) getHexValue(inputType string, value interface{}) (string, error) {

	var data string

	if strings.HasPrefix(inputType, "int") ||
		strings.HasPrefix(inputType, "uint") ||
		strings.HasPrefix(inputType, "fixed") ||
		strings.HasPrefix(inputType, "ufixed") {

		bigVal := value.(*big.Int)

		// Checking that the string actually is the correct inputType
		if strings.Contains(inputType, "128") {
			// 128 bit
			if bigVal.BitLen() > 128 {
				return "", errors.New(fmt.Sprintf("Input type %s not met", inputType))
			}
		} else if strings.Contains(inputType, "256") {
			// 256 bit
			if bigVal.BitLen() > 256 {
				return "", errors.New(fmt.Sprintf("Input type %s not met", inputType))
			}
		}

		data += fmt.Sprintf("%064s", fmt.Sprintf("%x", bigVal.String()))
	}

	if strings.Compare("address", inputType) == 0 {
		data += fmt.Sprintf("%064s", value.(string)[2:])
	}

	if strings.Compare("string", inputType) == 0 {
		data += fmt.Sprintf("%064s", fmt.Sprintf("%x", value.(string)))
	}

	return data, nil

}

//
func (contract *Contract) Send(transaction util.Transaction, functionName string, privatekey *ecdsa.PrivateKey, args ...interface{}) (string, error) {

	// transaction, err := contract.prepareTransaction(transaction, functionName, args)
	fixedArrStrPack, err := contract.abi.Pack(functionName, args...)
	if err != nil {
		return "", err
	}
	transaction.Input = hexutil.Encode(fixedArrStrPack)
	if err = contract.super.SignTransaction(&transaction, privatekey); err != nil {
		return "", err
	}
	return contract.super.SendTx(&transaction)

}

func (contract *Contract) Deploy(transaction util.Transaction, bytecode string, privatekey *ecdsa.PrivateKey, args ...interface{}) (string, error) {

	fixedArrStrPack, err := contract.abi.Pack("", args...)
	if err != nil {
		return "", err
	}
	transaction.Input = bytecode + hexutil.Encode(fixedArrStrPack)[2:]
	err = contract.super.SignTransaction(&transaction, privatekey)
	if err != nil {
		return "", err
	}
	return contract.super.SendTx(&transaction)
}

func (contract *Contract) Call(transaction util.Transaction, functionName string, args ...interface{}) (*dto.TxResult, error) {

	// transaction, err := contract.prepareTransaction(transaction, functionName, args)
	fixedArrStrPack, err := contract.abi.Pack(functionName, args...)
	if err != nil {
		return nil, err
	}
	transaction.Input = hexutil.Encode(fixedArrStrPack)
	return contract.super.CallTransaction(&transaction)

}

//解析
func (contract *Contract) Parse(callRes *dto.TxResult, name string, args interface{}) error {
	res, err := hexutil.Decode(callRes.Out)
	if err = contract.abi.Unpack(args, name, res); err != nil {
		return err
	} else {
		return nil
	}
}
func (contract *Contract) GetInput(functionName string, args ...interface{}) (string, error) {
	fixedArrStrPack, err := contract.abi.Pack(functionName, args...)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(fixedArrStrPack), err

}
func (contract *Contract) SendTransaction(transaction util.Transaction) (string, error) {
	return contract.super.SendTx(&transaction)
}
