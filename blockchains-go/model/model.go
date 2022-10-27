package model

import (
	"bytes"
	"encoding/json"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	//可以用uber的"github.com/json-iterator/go"替换
)

func DecodeSignData(ds []byte) (map[string]string, error) {
	var (
		err error
	)
	ret := map[string]string{}
	if err = json.Unmarshal(ds, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type ApplyAddrReq struct {

	//{
	//	"sign":"9c7c569508fdfa2b4acace722d7a967",
	//	"sfrom":"test",
	//	"outOrderId":"btc00000000001",
	//	"coinName":"btc",
	//	"num":1
	//}

	Sign       string `json:"sign" form:"sign"`
	Sfrom      string `json:"sfrom" form:"sfrom"`
	OutOrderId string `json:"outOrderId" form:"outOrderId"`
	CoinName   string `json:"coinName" form:"coinName"`
	Num        int64  `json:"num" form:"num"`
	util.ApiSignParams
}

func DecodeApplyAddrData(ds []byte) (ApplyAddrReq, error) {
	ret := ApplyAddrReq{}
	err := json.Unmarshal(ds, &ret)
	return ret, err
}

func DecodeSignDataInterface(ds []byte) (map[string]interface{}, error) {
	var (
		err error
	)
	ret := map[string]interface{}{}
	d := json.NewDecoder(bytes.NewReader(ds))
	d.UseNumber()
	d.Decode(&ret)
	if err = d.Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}
