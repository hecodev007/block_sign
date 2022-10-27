package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/proto"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func isAssignTest() bool {
	testOrder, err := redis.Client.Get("test_order")
	if err != nil {
		log.Errorf("isAssignTest error:%v", err)
		return false
	}
	return testOrder != ""
}

func GetWorker(coinName string) string {
	if coinName == "" {
		return ""
	}
	workerData, err := dao.FcWorkerFind()
	if err != nil {
		log.Errorf("查询机器异常，使用随机，err:%s", err.Error())
		return ""
	}
	if len(workerData) == 0 {
		log.Error("随机分配机器")
		return ""
	}
	workers := make([]string, 0)
	for _, v := range workerData {
		arr := strings.Split(v.CoinName, ",")
		for _, av := range arr {
			if strings.ToLower(av) == strings.ToLower(coinName) {
				//如果有限定，使用限定机器
				workers = append(workers, v.WorkerCode)
			}
		}
	}
	if len(workers) == 0 {
		//如果找不到限定机器，使用随机机器
		for _, v := range workerData {
			workers = append(workers, v.WorkerCode)
		}
	}
	return workers[rand.Intn(len(workers))]

}

type CreateResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    int64  `json:"data"`
}

func HttpCreateTx(url string, createReq *proto.OrderRequest) (int64, error) {
	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	signData, err := json.Marshal(createReq)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(signData))
	if err != nil {
		log.Errorf("Url:%s  ,Broadcast error: %v", url, err.Error())
		return -1, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Url:%s  ,Broadcast error: %v", url, err.Error())
		return -1, err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Url: %s  ,Broadcast error: %v", url, err.Error())
		return -1, err
	}

	respJson := &CreateResponse{}
	err = json.Unmarshal(respData, respJson)
	if err != nil {
		log.Errorf("DecodeHttpResopne  error : ", err)
		return -1, err
	}

	if respJson.Code != 0 {
		log.Errorf("get create response code: %d, message: %s", respJson.Code, respJson.Message)
		return -1, fmt.Errorf("get create response code: %d, message: %s", respJson.Code, respJson.Message)
	}

	return respJson.Data, nil
}
