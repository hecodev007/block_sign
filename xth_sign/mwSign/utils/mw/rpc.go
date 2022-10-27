package mw

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mwSign/common/log"
	"net/http"
	"net/url"
	"strings"

	"github.com/shopspring/decimal"
)

type RpcClient struct {
	url string
}

func NewRpcClient(url, username, password string) *RpcClient {
	rpc := &RpcClient{
		url: url,
	}
	return rpc
}

func (rpc *RpcClient) SendRawTransaction(rawtx string) (*BroadCastRsp, error) {
	host := fmt.Sprintf("%v/sharder?requestType=broadcastTransaction", rpc.url)

	v := url.Values{}
	v.Add("transactionBytes", rawtx)

	resp, err := http.Post(host,
		"application/x-www-form-urlencoded",
		strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Infof(string(body))
	sendmony := &BroadCastRsp{}
	err = json.Unmarshal(body, sendmony)
	if err != nil {
		return nil, err
	}
	if sendmony.ErrorDescription != "" {
		return nil, errors.New(sendmony.ErrorDescription)
	}
	return sendmony, nil
}

func (rpc *RpcClient) BuildTx(fromPubkey, reciept, amountNQT, fee string, deadline uint16) (*SendMoney, error) {
	host := fmt.Sprintf("%v/sharder?requestType=sendMoney", rpc.url)
	v := url.Values{}
	v.Add("publicKey", fromPubkey)
	v.Add("recipient", reciept)
	v.Add("amountNQT", amountNQT)
	v.Add("feeNQT", fee)
	v.Add("deadline", decimal.NewFromInt(int64(deadline)).String())

	resp, err := http.Post(host,
		"application/x-www-form-urlencoded	",
		strings.NewReader(v.Encode()))
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	sendmony := &SendMoney{}
	log.Infof(string(body))
	err = json.Unmarshal(body, sendmony)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	if sendmony.ErrorDescription != "" {
		log.Info(sendmony.ErrorDescription)
		return nil, errors.New(sendmony.ErrorDescription)
	}
	return sendmony, nil
}

type BroadCastRsp struct {
	RequestProcessingTime int    `json:"requestProcessingTime"`
	Transaction           string `json:"transaction"`
	ErrorDescription      string `json:"errorDescription"`
}
type SendMoney struct {
	TransactionJSON          TransactionJSON `json:"transactionJSON"`
	UnsignedTransactionBytes string          `json:"unsignedTransactionBytes"`
	ErrorDescription         string          `json:"errorDescription"`
}

type TransactionJSON struct {
	Attachment Attachment `json:"attachment"`
}

type Attachment struct {
	CrowdMinerRewardAmount  uint64 `json:"crowdMinerRewardAmount"`
	OrdinaryPayment         uint64 `json:"version.OrdinaryPayment"`
	PublicKeyAnnouncement   uint64 `json:"version.PublicKeyAnnouncement"`
	RecipientPublicKey      uint64 `json:"recipientPublicKey"`
	BlockMiningRewardAmount uint64 `json:"blockMiningRewardAmount"`
}
