package agent

import (
	"encoding/json"
	"fmt"
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"strings"
	"time"
)
var RiskCoins map[string]string
func init(){
	RiskCoins = make(map[string]string)
	RiskCoins["eth"] = "ETH"
	RiskCoins["eth0xdac17f958d2ee523a2206206994597c13d831ec7"] = "USDT_ETH"
}

//风险查询
//接口说明:https://documenter.getpostman.com/view/554226/TzY3Cvrj#8ce15062-6d5b-419c-8fd4-43327e3521f5
//
func IsRisk(dataBytes []byte) (data []byte,err error){
	riskUrl :=beego.AppConfig.DefaultString("riskhost", "")
	data = dataBytes
	dataMap := make(map[string]json.RawMessage)
	if err = json.Unmarshal(dataBytes,&dataMap);err != nil {
		return
	}
	coin :=  strings.Replace(string(dataMap["coin"]),"\"","",2)
	if string(dataMap["type"]) == "10" && RiskCoins[coin] != "" {
		var txsArray  []map[string]json.RawMessage
		if err = json.Unmarshal(dataMap["txs"],&txsArray);err != nil {
			return
		}
		for i,tx := range txsArray{
			contract := strings.Replace(string(tx["name"]),"\"","",2)
			txId := strings.Replace(string(tx["txid"]),"\"","",2)
			fromAddress := strings.Replace(string(tx["from"]),"\"","",2)
			coinName := strings.ToLower(coin+contract)

			if RiskCoins[coinName] == "" {
				continue
			}
			//_,_,_ = coinName,txId,fromAddress
			httpPost := httplib.Post(riskUrl+"/v1/kyt/transfer/received").SetTimeout(time.Second*2, time.Second*5)
			httpPost.Body([]byte(fmt.Sprintf("{\"asset\": \"%s\",\"uid\":\"2\", \"hash\": \"%s\", \"address\": \"%s\"}",RiskCoins[coinName],txId,fromAddress)))
			result, err := httpPost.Bytes()
			if err != nil {
				return data,err
			}
			riskResponse := new(RiskResponse)
			if err = json.Unmarshal(result,riskResponse);err != nil {
				return data,err
			}
			if riskResponse.Code != 200 {
				return data,errors.New(riskResponse.Msg)
			}

			riskParams := make(map[string]interface{})
			switch riskResponse.Data.Rating {
			case "highRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"]=2
				riskParams["riskmsg"]=riskResponse.Data.Cluster.Category
			case "lowRisk":
				riskParams["isrisk"] = true
				riskParams["risklevel"]=1
				riskParams["riskmsg"]=riskResponse.Data.Cluster.Category
			case "unknown":
				riskParams["isrisk"] = false
				riskParams["risklevel"]=0
				riskParams["riskmsg"]="unknown"
			default:
				riskParams["isrisk"] = false
				riskParams["risklevel"]=0
				riskParams["riskmsg"] = riskResponse.Data.Rating
			}

			riskBytes,_ := json.Marshal(riskParams)
			riskMap := make(map[string]json.RawMessage)
			if err = json.Unmarshal(riskBytes,&riskMap);err != nil {
				return data,err
			}

			for riskey,riskvalue := range riskMap{
				txsArray[i][riskey] = riskvalue
			}
		}
		if dataMap["txs"],err = json.Marshal(txsArray);err != nil{
			return data,err
		}

		return json.Marshal(dataMap)
	} else {
		data = nil
	}
	return data,err

}


type RiskResponse struct{
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data struct{
		TransferReference string `json:"transferReference"`
		Asset string `json:"asset"`
		Rating string `json:"rating"`

		Cluster struct{
			Name string `json:"name"`
			Category string `json:"category"`
		}
	}
}