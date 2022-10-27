package util

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

//生成orderID
//AES规则:address和金额进行拼接。key为外部订单ID
func GetOrderId(outOrderIdKey string, address, amount string) (string, error) {
	if len(outOrderIdKey) > 16 {
		//截取16位
		outOrderIdKey = outOrderIdKey[:aes.BlockSize]
	}
	dataStr := fmt.Sprintf("%s_%s", address, amount)
	result, err := AesCrypt([]byte(dataStr), []byte(outOrderIdKey), true)
	if err != nil {
		return "", err
	}
	orderNo := base64.StdEncoding.EncodeToString(result)
	orderNo = fmt.Sprintf("%s_%d", orderNo, time.Now().Unix())
	return orderNo, nil
}

//解密orderId
func DecodeOrderId(outOrderIdKey string, orderNo string) (address, amount string, err error) {
	arrOrderNoStr := strings.Split(orderNo, "_")
	if len(arrOrderNoStr) > 2 {
		return "", "", fmt.Errorf("orderNo:%s,非法", orderNo)
	}
	if len(outOrderIdKey) > 16 {
		//截取16位
		outOrderIdKey = outOrderIdKey[:aes.BlockSize]
	}
	decodeResult, _ := AesBase64Str(arrOrderNoStr[0], outOrderIdKey, false)
	arr := strings.Split(decodeResult, "_")
	if len(arr) != 2 {
		return "", "", fmt.Errorf("orderId:%s,decodeResult error :%s", orderNo, decodeResult)
	}
	address = arr[0]
	amount = arr[1]
	if address == "" || amount == "" {
		return "", "", fmt.Errorf("orderId:%s,empty decode", orderNo)

	}
	return address, amount, nil
}
