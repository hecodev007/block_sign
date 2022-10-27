package main

import (
	"algoSign/common/keystore"
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/shopspring/decimal"
	"log"
)

//go run priexport.go http://algo.rylink.io:27833 2dc5bfd8ed194b0ab271f4bcee568cacffd1adb1beadd150e594aa3e35e3729a test 123456
func init() {
	log.SetFlags(log.Llongfile)
}
func main() {
	url :="http://algo.rylink.io:27833"
	apitoken :="2dc5bfd8ed194b0ab271f4bcee568cacffd1adb1beadd150e594aa3e35e3729a"
	url = "http://algo.rylink.io:35329"
	apitoken = "9a5bbe7ecae7fa6d81495a78b113602d6629ddf3d5b4d540ffbea8cebdf2495c"

	//kmdc ,err :=kmd.MakeClient(url, apitoken)
	//if err != nil {
	//	panic(err.Error())
	//}

	algodc,err :=algod.MakeClient(url, apitoken)
	if err != nil {
		panic(err.Error())
	}
	node, err := keystore.ReadCsvFile("./algo-node.csv", false)

	if err != nil {
		panic(err.Error())
	}
	l := make([]Mount,0)
	num := uint64(0)
	for k, _ := range node {
		//k = "YDRTADVLLGLU76ARI4NU5V4XTXPHDJII7A5MY3UL6II77LGE5SPQUKJWI4"
		rsp,err :=algodc.AccountInformation(k)
		if err != nil {
			panic(err.Error())
		}
		if rsp.Amount == 0{
			continue
		}
		num += rsp.Amount
		m := Mount{Addr: k,Value: rsp.Amount}
		l = append(l,m)
		log.Println(k,rsp.Amount,len(l))
		for i := len(l)-2;i>=0;i--{
			if l[i].Value<l[i+1].Value{
				l[i],l[i+1] = l[i+1],l[i]
			} else {
				break
			}
		}

	}
	//m := Mount{Addr: "total",Value: num}
	//l = append(l,m)
	log.Println(String(l))

	for k,_ := range l{
		fmt.Println(l[k].Addr,",",decimal.NewFromInt(int64(l[k].Value)).Shift(-6).String())
	}
}

type Mount struct {
	Addr string
	Value uint64
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
