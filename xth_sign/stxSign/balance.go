package main

import (
	"fmt"
	"stxSign/common/keystore"
	"stxSign/utils/stx"
)

func main(){
	client := stx.NewRpcClient("http://stx.rylink.io:30999","","")
	keys,err := keystore.ReadCsvFile("./hoo.balance.csv",false)
	if err != nil {
		panic(err.Error())

	}
	i :=0
	for index,v := range keys{
		i++
		//fmt.Println(index,i,len(keys))
		if v.Key== "0"{
			continue
		}
		fmt.Println(index,v.Key,i)
			value,err :=client.GetBalance(v.Address)
			if err != nil{
				panic(err.Error())
			}
			if	value.String()!=v.Key{
				panic(v.Address)
			}

	}
	//err = keystore.WriteCsvFile(keylist,"./hoo.balance.csv")
	//if err != nil {
	//	panic(err.Error())
	//}

}
