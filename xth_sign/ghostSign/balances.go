package main

import (
	"encoding/csv"
	"fmt"
	"ghostSign/utils/btc"
	"github.com/shopspring/decimal"
	"io"
	"io/ioutil"
	"strings"
	"os"
)

func main() {
	//go run balances.go http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@52.69.242.206:30555 dogeBalances.csv
	//go run balances.go http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@18.183.171.119:31407 bchaBalances.csv
	//http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@18.181.213.202:30232
	url := "http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@18.182.64.34:30098"
	if len(os.Args)>=2{
		url = os.Args[1]
	}
	client := btc.NewRpcClient(url, "", "")
	var utxos []UnSpend
	err := client.CallWithAuth("listunspent", client.Credentials, &utxos)
	if err != nil {
		panic(err.Error())
	}
	listBalance := make(map[string]decimal.Decimal)
	 for _,utxo := range utxos {
		if _,ok := listBalance[utxo.Address];!ok{
			listBalance[utxo.Address] = decimal.NewFromInt(0)
		}
		 //listBalance[utxo.Address] =  listBalance[utxo.Address].Add(decimal.NewFromFloat(utxo.Amount))
		 listBalance[utxo.Address] =  listBalance[utxo.Address].Add(utxo.Amount)
	 }
	 var listCsv []*CsvKey

	for addr,amount := range listBalance{
		v := &CsvKey{
			Address: addr,
			Key: amount.String(),
		}
		listCsv = append(listCsv,v)
	}
	filename := "balancelist.csv"
	if len(os.Args)>=3{
		filename = os.Args[2]
	}
	for i:=0;i<len(listCsv)-1;i++{
		for j:=len(listCsv)-1;j>i;j--{
			jkey,_:=decimal.NewFromString( listCsv[j].Key)
			j2key,_:=decimal.NewFromString( listCsv[j-1].Key)
			if jkey.Cmp(j2key)>0{
				listCsv[j],listCsv[j-1] = listCsv[j-1],listCsv[j]
			}
		}
	}
	err = WriteCsvFile(listCsv,filename)
	if err != nil {
		panic(err.Error())
	}

}

//UTXO 数据结构
type UnSpend struct {
	Txid          string
	Vout          uint
	Address       string
	RedeemScript  string
	ScriptPubKey  string
	Amount        decimal.Decimal
	Confirmations uint64
	Spendable     bool
	Solvable      bool
	Safe          bool
}

func ReadCsvFile(fileName string, toLower bool) (ret []string, err error) {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("read csv file %v", err)
		return nil, err
	}
	ret = make([]string, 0, 102000)
	//返回的是csv.Reader
	r := csv.NewReader(strings.NewReader(string(data)))
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		//fmt.Println(len(ret), line[0])
		ret = append(ret, line[0])

	}
	return ret, nil
}

// 返回任何可能发生的错误
func WriteCsvFile(cvsKeys []*CsvKey, fileName string) error {

	csvFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {

		return err
	}
	defer csvFile.Close()

	n := csv.NewWriter(csvFile)
	for _, cvsKey := range cvsKeys {
		err := n.Write([]string{cvsKey.Address, cvsKey.Key})
		if err != nil {
			return err
		}
	}

	n.Flush()
	return n.Error()
}

type CsvKey struct {
	Address string
	Key     string // a - aesprivatekey  b - aeskey    c - privatekey
}