package main

import (
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/shopspring/decimal"
	"os"
	"path"
	"fmt"
	"path/filepath"
	"strings"
)
import "cfxSign/utils/keystore"
func main2(){
	file_old:= "csvbak"
	dirAbsPath, err := filepath.Abs("./"+file_old)
	if err != nil {
		panic(err.Error())
	}
	csvFiles, err := keystore.ListCsvFile(dirAbsPath)
	if err != nil {
		panic(err.Error())
	}

	for _, acsv := range csvFiles {
		//if !strings.HasSuffix(acsv,"_a.csv"){
		//	continue
		//}
		//bcsv := strings.Replace(acsv,"_a.csv","_b.csv",1)
		keylist := make([]*keystore.CsvKey,0)
		akeys,err :=keystore.ReadCsvFile(acsv,false)
		if err != nil{
			panic(err.Error())
		}
		for _,v := range akeys{
			addr :=cfxaddress.MustNewFromHex(v.Address,cfxaddress.NetowrkTypeMainnetID)
			v.Address = addr.MustGetBase32Address()
			keylist = append(keylist,v)
		}
		newpath := strings.Replace(acsv,"/"+file_old+"/","/csvnew/",1)

		if err := os.MkdirAll(path.Dir(newpath), os.ModePerm); err != nil {
			panic(err.Error())
		}
		err = keystore.WriteCsvFile(keylist,newpath)
		if err != nil{
			panic(err.Error())
		}
	}

	for _, acsv := range csvFiles {
		if !strings.HasSuffix(acsv,"_a.csv"){
			continue
		}
		//bcsv := strings.Replace(acsv,"_a.csv","_b.csv",1)

		keylist := make([]*keystore.CsvKey,0)
		akeys,err :=keystore.ReadCsvFile(acsv,false)
		if err != nil{
			panic(err.Error())
		}
		for _,v := range akeys{
			//fmt.Println(v.Address)
			addr :=cfxaddress.MustNewFromHex(v.Address,cfxaddress.NetowrkTypeMainnetID)
			v.Key = addr.MustGetBase32Address()
			//v.Key = v.Address
			keylist = append(keylist,v)
		}
		newpath := strings.Replace(acsv,"/"+file_old+"/","/csvSQL/",1)

		if err := os.MkdirAll(path.Dir(newpath), os.ModePerm); err != nil {
			panic(err.Error())
		}
		err = keystore.WriteCsvFile(keylist,newpath)
		if err != nil{
			panic(err.Error())
		}
	}



	client,err :=sdk.NewClient("http://main.confluxrpc.org/v2")
	if err != nil {
		panic(err.Error())
	}
	for _, acsv := range csvFiles {
		if !strings.HasSuffix(acsv,"_a.csv"){
			continue
		}
		//bcsv := strings.Replace(acsv,"_a.csv","_b.csv",1)

		keylist := make([]*keystore.CsvKey,0)
		akeys,err :=keystore.ReadCsvFile(acsv,false)
		if err != nil{
			panic(err.Error())
		}
		for _,v := range akeys{
			fmt.Println(v.Address)
			addr :=cfxaddress.MustNewFromHex(v.Address,cfxaddress.NetowrkTypeMainnetID)
			v.Address = addr.MustGetBase32Address()
			//v.Key = v.Address
			getbalance:
			balance,err := client.GetBalance(addr)
			if err != nil {
				fmt.Println(err.Error())
				goto getbalance
			}
			v.Key = decimal.NewFromBigInt(balance.ToInt(),-18).String()
			keylist = append(keylist,v)
		}
		newpath := strings.Replace(acsv,"/"+file_old+"/","/csvbalance/",1)

		if err := os.MkdirAll(path.Dir(newpath), os.ModePerm); err != nil {
			panic(err.Error())
		}
		err = keystore.WriteCsvFile(keylist,newpath)
		if err != nil{
			panic(err.Error())
		}
	}
}