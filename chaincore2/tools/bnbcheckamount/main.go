package main

import (
	"encoding/csv"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type BnbBalance struct {
	Symbol string `json:"symbol"`
	Amount string `json:"free"`
}

type BnbAccount struct {
	AccountNumber int64        `json:"account_number"`
	Address       string       `json:"address"`
	Balances      []BnbBalance `json:"balances"`
	Sequence      int64        `json:"sequence"`
}

func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Debug("Bnb: %v ", err)
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debug("Bnb: %v ", err)
		return nil, err
	}

	return content, nil
}

func GetAmount(fromAddr string) (error, float64) {
	//// 查找from金额
	//totalAmount := 0.0
	//url := fmt.Sprintf("%s/%s", "http://bnb.rylink.io:30080/api/v1/account", fromAddr)
	//log.Debug(url)
	//data, err := Get(url)
	//if data == nil || err != nil {
	//	return err, totalAmount
	//}
	//
	//var account bo.BnbAccount
	//if err = json.Unmarshal(data, &account); err != nil {
	//	log.Debug(err, fromAddr)
	//	return nil, totalAmount
	//}
	//
	//for i := 0; i < len(account.Balances); i++ {
	//	if account.Balances[i].Symbol == "BNB" {
	//		tmp, _ := strconv.ParseFloat(account.Balances[i].Amount, 64)
	//		totalAmount = tmp
	//		break
	//	}
	//}
	return nil, 0.0
}

func main() {
	var path string = beego.AppConfig.String("csvdir")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fs, err := os.Open(path)
	if err != nil {
		log.Debug("can not open the file, err is %+v", err)
	}
	defer fs.Close()

	list := [][]string{}
	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Debug("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}

		list = append(list, row)
	}

	clientsFile, err := os.OpenFile("./bnb_check.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Debug(err)
		return
	}
	defer clientsFile.Close()
	clients := csv.NewWriter(clientsFile)
	for i := 0; i < len(list); {
		err, amount := GetAmount(list[i][0])
		if err != nil {
			log.Debug(err, list[i][0])
			continue
		}

		list[i] = append(list[i], fmt.Sprintf("%.8f", amount))
		clients.Write(list[i])
		i++
	}
	clients.Flush()
}
