package model

import "errors"

type MchSearchBalance struct {
	Sign            string `form:"sign"`
	Sfrom           string `form:"sfrom"`
	CoinName        string `form:"coinName"`
	TokenName       string `form:"tokenName"`
	ContractAddress string `form:"contractAddress"`
}

func (m *MchSearchBalance) IsContract() bool {
	if m.TokenName != "" || m.ContractAddress != "" {
		return true
	}
	return false
}

func (m *MchSearchBalance) CheckContract() error {
	if m.TokenName == "" {
		return errors.New("miss tokenName")
	}
	if m.ContractAddress == "" {
		return errors.New("miss contractAddress")
	}
	return nil
}
