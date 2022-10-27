package kardia

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	BTM_CONTRACT_ADDR  = "0xcb97e65f07da24d46bcdd078ebebd7c6e6e3d750"
	HT_CONTRACT_ADDR   = "0x6f259637dcd74c767781e37bc6133cd6a68aa161"
	OMG_CONTRACT_ADDR  = "0xd26114cd6ee289accf82350c8d8487fedb8a0c07"
	ZRX_CONTRACT_ADDR  = "0xe41d2489571d322189246dafa5ebde1f4699f498"
	CVC_CONTRACT_ADDR  = "0x41e5560054824ea6b0732e656e3ad64e20e94e45"
	RBF_CONTRACT_ADDR  = "0xd22ae98282037ba2db1cbe3b2c3ee3089b925ba4"
	KAN_CONTRACT_ADDR  = "0x1410434b0346f5be678d0fb554e5c7ab620f8f4a"
	DLB_CONTRACT_ADDR  = "0xce1d3da32e3a45d27dc841781f09e40c41cac677"
	AISA_CONTRACT_ADDR = "0xc0951f25ee235675238afef28582eec047f78e4a"
	LEEK_CONTRACT_ADDR = "0x42c41dabf7962be4f510d54aa9eb0d2240634842"
	DCC_CONTRACT_ADDR  = "0xffa93aacf49297d51e211817452839052fdfb961"
	PAY_CONTRACT_ADDR  = "0xb97048628db6b661d4c2aa833e95dbe1a905b280"
	PIN_CONTRACT_ADDR  = "0x93ed3fbe21207ec2e8f2d3c3de6e058cb73bc04d"
	PAX_CONTRACT_ADDR  = "0x005275450e77bfa6bcbd04d85175d5d0f2dfae43"
	CHAT_CONTRACT_ADDR = "0x442bc47357919446eabc18c7211e57a13d983469"
	BNB_CONTRACT_ADDR  = "0xb8c77482e45f1f44de1745f52c74426c631bdd52"
	EKT_CONTRACT_ADDR  = "0xbab165df9455aa0f2aed1f2565520b91ddadb4c8"
	AE_CONTRACT_ADDR   = "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d"
	IOST_CONTRACT_ADDR = "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab"
	BCV_CONTRACT_ADDR  = "0x1014613e2b3cbc4d575054d4982e580d9b99d7b1"
)

var ContractAddrs = map[string]string{
	"BCV": BCV_CONTRACT_ADDR,
	"HT":  HT_CONTRACT_ADDR,
}

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
