package transfer

import (
	"encoding/json"
	"fmt"
)

func DecodeWalletServerRespOrder(data []byte) *WalletServerRespOrder {
	if len(data) != 0 {
		result := new(WalletServerRespOrder)
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}

type TransferHotResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Txid    string      `json:"txid"`
}

func DecodeTransferHotResp(data []byte) (*TransferHotResp, error) {
	var thr TransferHotResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return nil, fmt.Errorf("parse transfer hot resp data error,Err=%v", err)
	}
	return &thr, nil
}

type ValidAddressResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func DecodeValidAddressResp(data []byte) error {
	var thr ValidAddressResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return fmt.Errorf("parse valid address resp data error,Err=%v", err)
	}
	if thr.Code == 0 && thr.Data != nil {
		if valid, ok := thr.Data.(bool); ok {
			if valid {
				return nil
			}
			return fmt.Errorf("valid address is not true: status=%v", valid)
		}
		return fmt.Errorf("retutn valid addres result is not bool : %v", thr.Data)
	}

	return fmt.Errorf("valid address error: %s", thr.Message)
}

type GetBalanceResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func DecodeGetBalanceResp(data []byte) (*GetBalanceResp, error) {
	var thr GetBalanceResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return nil, fmt.Errorf("parse get balance resp data error,Err=%v", err)
	}
	if thr.Code == 0 && thr.Data != nil {
		return &thr, nil
	}
	return nil, fmt.Errorf("get balance error: %s", thr.Message)
}

type CreateTxResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func DecodeCreateTxResp(data []byte) (*CreateTxResp, error) {
	var thr CreateTxResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return nil, fmt.Errorf("parse transfer hot resp data error,Err=%v", err)
	}
	return &thr, nil
}

type CreateUsdtResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Hash    string      `json:"hash"`
}

func DecodeCreateUsdtResp(data []byte) (*CreateUsdtResp, error) {
	var thr CreateUsdtResp
	err := json.Unmarshal(data, &thr)
	if err != nil {
		return nil, fmt.Errorf("parse transfer hot resp data error,Err=%v", err)
	}
	return &thr, nil
}
