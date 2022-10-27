package bsc

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type ERC20 struct {
}

func (erc ERC20) Transfer(to string, amount *big.Int) (data string, err error) {
	if !erc.isAddress(to) {
		return data, errors.New("to isn't address format")
	}
	data = fmt.Sprintf("0xa9059cbb%064s%064x", to[2:], amount)

	return data, err
}

func (erc ERC20) TransferFrom(from string, to string, amount *big.Int) (data string, err error) {
	if !erc.isAddress(from) {
		return data, errors.New("from isn't address format")
	}
	if !erc.isAddress(to) {
		return data, errors.New("to isn't address format")
	}
	data = fmt.Sprintf("0x23b872dd%064s%064s%064x", from[2:], to[2:], amount)

	return data, err
}

func (erc ERC20) Approve(to string, amount *big.Int) (data string, err error) {
	if !erc.isAddress(to) {
		return data, errors.New("to isn't address format")
	}
	data = fmt.Sprintf("0x095ea7b3%064s%064x", to[2:], amount)

	return data, err
}

func (erc ERC20) GetBalanceOf(address string) (data string, err error) {
	if !erc.isAddress(address) {
		return data, errors.New("address isn't address format")
	}
	data = fmt.Sprintf("0x70a08231%064s", address[2:])

	return data, err
}

func (erc ERC20) GetAllowance(owner string, spender string) (data string, err error) {
	if !erc.isAddress(owner) {
		return data, errors.New("address isn't address format")
	}
	if !erc.isAddress(spender) {
		return data, errors.New("spender isn't address format")
	}
	data = fmt.Sprintf("0xdd62ed3e%064s%064s", owner[2:], spender[2:])

	return data, err
}

func (ERC20) ParseTransferData(input string) (to string, amount *big.Int, err error) {
	//0xa9059cbb0000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000000000000000000000000000000000000005f5e100
	if strings.Index(input, "0xa9059cbb") != 0 {
		return to, amount, errors.New("input is not transfer data")
	}
	if len(input) < 138 {
		return to, amount, fmt.Errorf("input data isn't 138 , size %d ", 138)
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:138], 16)
	if amount.Sign() < 0 {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}

func (ERC20) ParseTransferFromData(input string) (from string, to string, amount *big.Int, err error) {
	//0x23b872dd0000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000000000000000000000000000000000000005f5e100
	if strings.Index(input, "0x23b872dd") != 0 {
		return from, to, amount, errors.New("input is not transferFrom data")
	}
	from = "0x" + input[34:74]
	to = "0x" + input[98:138]
	amount = new(big.Int)
	amount.SetString(input[138:], 16)
	if !amount.IsUint64() {
		return from, to, amount, errors.New("bad amount data")
	}
	return from, to, amount, nil
}

func (ERC20) ParseApproveData(input string) (to string, amount *big.Int, err error) {
	//0x095ea7b30000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000000000000000000000000000000000000005f5e100
	if strings.Index(input, "0x095ea7b3") != 0 {
		return to, amount, errors.New("input is not approve data")
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:], 16)
	if !amount.IsUint64() {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}

func (ERC20) isAddress(address string) bool {
	bigInt := new(big.Int)
	_, ok := bigInt.SetString(address, 0)

	if !ok || len(address) != 42 {
		return false
	} else {
		return true
	}
}
