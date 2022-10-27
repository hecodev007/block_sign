package v3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"io/ioutil"
	"net/http"
	"strings"
)

type RequestBody struct { //要求data
	Asset         string `json:"asset"`   //传账币种
	OutputAddress string `json:"address"` //提款目标地址
}
type BadReq struct { //返回值转json，struct
	Status  int    `json:"status"`  //回传状态
	Message string `json:"message"` //失败信息

}

type ResultData []struct { //返回值转struct
	Asset   string `json:"asset"`   //币种
	Address string `json:"address"` //提款目标地址
	Cluster struct {
		Name     string `json:"name"`
		Category string `json:"category"`
	} `json:"cluster"`                                      //地址属性
	Rating                    string `json:"rating"`        //风险等级 low/medium/high/severe/other
	CustomAddress             string `json:"customAddress"` //自定义地址
	ChainalysisIdentification struct {
		AddressName  string `json:"addressName"`
		Description  string `json:"description"`
		CategoryName string `json:"categoryName"`
	} `json:"chainalysisIdentification"` //chainalysis
}

func checkRiskFromKYT(outerOrderNo, walletId, walletAsset, walletAddress string) (string, string, error) {
	if walletId == "" || walletAsset == "" || walletAddress == "" {
		return "", "", errors.New("invalid params")
	}

	mp := make(map[string]string) //转换为KYT支持的币种名称。
	mp["matic-matic"] = "matic"   //matic-matic->matic
	mp["heco"] = "ht"             //heco->ht
	mp["hsc"] = "bnb"             //hsc->bnb
	coinname, ok := mp[strings.ToLower(walletAsset)]
	if !ok {
		coinname = walletAsset
	}
	kytv1url := "https://api.chainalysis.com/api/kyt/v1/users/" + walletId + "/withdrawaladdresses" //接口地址
	token := "8e283fdd013173ce7712d8e5ae9e9d21655274956bddbb40388a4a9229b135a5"                     //Api key时效： 05/10/2021 — 05/23/2022
	kyt := RequestBody{Asset: coinname, OutputAddress: walletAddress}                               // data:byte 内容
	requestBody, err := json.Marshal(&kyt)                                                          //转json
	list := `[` + string(requestBody) + `]`                                                         //添加[]
	var data = []byte(list)
	req, err := http.NewRequest(http.MethodPost, kytv1url, bytes.NewBuffer(data))
	req.Header.Set("Token", token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusBadRequest { //请求为400回传的结果
		var badreq BadReq
		err = json.Unmarshal(body, &badreq)
		if err != nil {
			log.Error(err)
			return "", "", errors.New(badreq.Message)
		}
	}
	var result ResultData
	if resp.StatusCode == http.StatusOK { //200
		err := json.Unmarshal(body, &result)
		if err != nil {
			return "", "", err
		}
	}
	if len(result) == 0 {
		return "", "", errors.New("response data is empty")
	}
	rating := result[0].Rating
	msg := result[0].Cluster.Name + result[0].Cluster.Category
	if rating != "" && rating != "unknown" {
		// 报警
		notifyMsg := fmt.Sprintf("订单: %s chainalysis风险等级: %s 消息: %s", outerOrderNo, rating, msg)
		dingding.ErrTransferDingBot.NotifyStr(notifyMsg)
		log.Info(notifyMsg)
	}
	return rating, msg, nil
}
