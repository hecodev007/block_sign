package util

import (
	"encoding/json"
	"errors"
)

type RawContent struct {
	OutOrderId       string `json:"out_order_id"`
	ToAddress        string `json:"to_address"`
	MainCoin         string `json:"main_coin"`
	Token            string `json:"token"`
	ToAmountFloatStr string `json:"to_amount_float_str"`
	CreateAt         string `json:"create_at"`
}

func CreateCheckApplyContent(raw RawContent) (string, error) {
	if raw.OutOrderId == "" || raw.ToAddress == "" || raw.MainCoin == "" || raw.ToAmountFloatStr == "" || raw.CreateAt == "" {
		return "", errors.New("参数异常异常，生成加密内容失败")
	}
	key := raw.CreateAt + "hoo@hoo" + raw.CreateAt
	key = key[:16]
	dd, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return AesBase64Str(string(dd), key, true)
}

func DecodeCheckApplyContent(applyId, createAt, context string) (*RawContent, error) {
	if applyId == "" {
		return nil, errors.New("参数异常异常，生成加密内容失败")
	}
	key := createAt + "hoo@hoo" + createAt
	key = key[:16]

	result, err := AesBase64Str(context, key, false)
	if err != nil {
		return nil, err
	}
	raw := new(RawContent)
	err = json.Unmarshal([]byte(result), raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
