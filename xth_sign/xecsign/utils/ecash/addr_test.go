package ecash

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/astaxie/beego/httplib"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchd/txscript"
)

func Test_addr(t *testing.T) {
	//laddr, caddr, pri, err := GenAccount()
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//t.Log(laddr, caddr, pri)
	caddr := "1AnJeoPNk4yzRv5GqoqusbpEtbmLzVJteN"
	addr, err := ToCashAddr(caddr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr)

	caddr, err = CashToAddr(addr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(caddr)
}
func Test_script(t *testing.T) {
	script, err := hex.DecodeString("a914260617ebf668c9102f71ce24aba97fcaaf9c666a87")
	if err != nil {
		t.Fatal(err.Error())
	}
	_, addrs, num, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Log(num, addrs[0].EncodeAddress())
}

func Test_map(t *testing.T) {
	dataBytes := []byte("{\"type\":10,\"coin\":\"eth\",\"token\":\"\",\"height\":7544345,\"hash\":\"0xe03c97e725be0dc3b863d6e03917907ed1e726f20d94007ca5dd25111c98388f\",\"confirmations\":5183570,\"time\":1554952803,\"txs\":[{\"name\":\"\",\"txid\":\"0x59f72572c6d26cb37ac086be81460526bcde183cc9438d3d6a6b94758cb9b2c8\",\"fee\":\"0.000186666666648\",\"from\":\"0xd2b7b2e073a1e36326ba5d40a9f042846ef4a4a3\",\"to\":\"0x0030ef21fecbeb365ea06b301b8f62abece9adaf\",\"amount\":\"0.1\",\"memo\":\"\",\"contract\":\"\"}]}")
	data, err := IsRisk(dataBytes, t)
	t.Log(string(data), err)
	data, err = IsRisk2(dataBytes)
	t.Log(string(data), err)

}

func IsRisk(dataBytes []byte, t *testing.T) (data []byte, err error) {
	RiskCoins := make(map[string]string)
	RiskCoins["eth"] = "ETH"
	RiskCoins["ethusdt-erc20"] = "USDT_ETH"
	riskUrl := "http://47.243.51.53:9233"
	//data = dataBytes
	dataMap := make(map[string]json.RawMessage)
	if err = json.Unmarshal(dataBytes, &dataMap); err != nil {
		return
	}

	coin := strings.Replace(string(dataMap["coin"]), "\"", "", 2)

	if string(dataMap["type"]) == "10" && RiskCoins[coin] != "" {
		var txsArray []map[string]json.RawMessage
		if err = json.Unmarshal(dataMap["txs"], &txsArray); err != nil {
			return
		}
		t.Log(coin, string(dataMap["type"]), txsArray)

		for i, tx := range txsArray {
			contract := strings.Replace(string(tx["contract"]), "\"", "", 2)
			txId := strings.Replace(string(tx["txid"]), "\"", "", 2)
			fromAddress := strings.Replace(string(tx["from"]), "\"", "", 2)
			t.Log(contract, txId, fromAddress)
			//coinName := coin + contract
			if RiskCoins[coin+contract] == "" {
				continue
			}
			_, _, _ = contract, txId, fromAddress
			t.Log(contract, txId)
			httpPost := httplib.Post(riskUrl+"/v1/kyt/transfer/received").SetTimeout(time.Second*2, time.Second*5)
			httpPost.Body([]byte(fmt.Sprintf("{\"asset\": \"%s\",\"uid\":\"2\", \"hash\": \"%s\", \"address\": \"%s\"}", RiskCoins[coin+contract], txId, fromAddress)))
			result, err := httpPost.Bytes()
			if err != nil {
				return data, err
			}
			t.Log(string(result))
			riskResponse := new(RiskResponse)
			if err = json.Unmarshal(result, riskResponse); err != nil {
				return data, err
			}
			if riskResponse.Code != 200 {
				return data, errors.New(riskResponse.Msg)
			}

			riskParams := make(map[string]interface{})
			switch riskResponse.Data.Rating {
			case "highRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"] = 2
				riskParams["riskmsg"] = riskResponse.Data.Cluster.Category
			case "lowRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"] = 1
				riskParams["riskmsg"] = riskResponse.Data.Cluster.Category
			case "unknown":
				riskParams["isrisk"] = false
				riskParams["risklevel"] = 0
				riskParams["riskmsg"] = "unknown"
			default:
				riskParams["isrisk"] = false
				riskParams["risklevel"] = 0
				riskParams["riskmsg"] = riskResponse.Data.Rating
			}

			riskBytes, _ := json.Marshal(riskParams)
			riskMap := make(map[string]json.RawMessage)
			if err = json.Unmarshal(riskBytes, &riskMap); err != nil {
				return data, err
			}

			for riskey, riskvalue := range riskMap {
				txsArray[i][riskey] = riskvalue
			}
		}
		if dataMap["txs"], err = json.Marshal(txsArray); err != nil {
			return data, err
		}

		return json.Marshal(dataMap)
	} else {
		data = nil
	}
	return data, err

}

func IsRisk2(dataBytes []byte) (data []byte, err error) {
	RiskCoins := make(map[string]string)
	RiskCoins["eth"] = "ETH"
	RiskCoins["ethusdt-erc20"] = "USDT_ETH"
	riskUrl := "http://47.243.51.53:9233"

	//riskUrl :=beego.AppConfig.DefaultString("riskhost", "")
	data = dataBytes
	dataMap := make(map[string]json.RawMessage)
	if err = json.Unmarshal(dataBytes, &dataMap); err != nil {
		return
	}
	coin := strings.Replace(string(dataMap["coin"]), "\"", "", 2)
	if string(dataMap["type"]) == "10" && RiskCoins[coin] != "" {
		var txsArray []map[string]json.RawMessage
		if err = json.Unmarshal(dataMap["txs"], &txsArray); err != nil {
			return
		}
		for i, tx := range txsArray {
			contract := strings.Replace(string(tx["name"]), "\"", "", 2)
			txId := strings.Replace(string(tx["txid"]), "\"", "", 2)
			fromAddress := strings.Replace(string(tx["from"]), "\"", "", 2)
			coinName := coin + contract

			if RiskCoins[coinName] == "" {
				continue
			}
			_, _, _ = coinName, txId, fromAddress
			httpPost := httplib.Post(riskUrl+"/v1/kyt/transfer/received").SetTimeout(time.Second*2, time.Second*5)
			httpPost.Body([]byte(fmt.Sprintf("{\"asset\": \"%s\",\"uid\":\"2\", \"hash\": \"%s\", \"address\": \"%s\"}", RiskCoins[coinName], txId, fromAddress)))
			result, err := httpPost.Bytes()
			if err != nil {
				return data, err
			}
			riskResponse := new(RiskResponse)
			if err = json.Unmarshal(result, riskResponse); err != nil {
				return data, err
			}
			if riskResponse.Code != 200 {
				return data, errors.New(riskResponse.Msg)
			}

			riskParams := make(map[string]interface{})
			switch riskResponse.Data.Rating {
			case "highRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"] = 2
				riskParams["riskmsg"] = riskResponse.Data.Cluster.Category
			case "lowRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"] = 1
				riskParams["riskmsg"] = riskResponse.Data.Cluster.Category
			case "unknown":
				riskParams["isrisk"] = false
				riskParams["risklevel"] = 0
				riskParams["riskmsg"] = "unknown"
			default:
				riskParams["isrisk"] = false
				riskParams["risklevel"] = 0
				riskParams["riskmsg"] = riskResponse.Data.Rating
			}

			riskBytes, _ := json.Marshal(riskParams)
			riskMap := make(map[string]json.RawMessage)
			if err = json.Unmarshal(riskBytes, &riskMap); err != nil {
				return data, err
			}

			for riskey, riskvalue := range riskMap {
				txsArray[i][riskey] = riskvalue
			}
		}
		if dataMap["txs"], err = json.Marshal(txsArray); err != nil {
			return data, err
		}

		return json.Marshal(dataMap)
	} else {
		data = nil
	}
	return data, err

}

type RiskResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TransferReference string `json:"transferReference"`
		Asset             string `json:"asset"`
		Rating            string `json:"rating"`

		Cluster struct {
			Name     string `json:"name"`
			Category string `json:"category"`
		}
	}
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
