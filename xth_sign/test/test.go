package main

import (
	"encoding/csv"
	"fmt"
	"github.com/xutonghuary/signServer/ghostSign/utils/btc"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

func main() {
	keys, err := ReadCsvFile("./202010141603_d.csv", false)
	if err != nil {
		panic(err.Error())
	}
	client := btc.NewRpcClient("http://rylink:4CpmLnbOiaTbD20gPdsRYY6WiMFDyF8N8QzGGYrAfIw=@dash.rylink.io:30098", "", "")
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	for k := 0; k < len(keys); k++ {
		wg.Add(1)

		v := keys[k]
		fmt.Println(v)
		go func(v string) {
			defer wg.Done()
		st:
			if _, err := client.Importaddrs(v); err != nil {
				fmt.Println(err.Error())
				goto st
			}

		}(v)
		//st:
		fmt.Println(v, k, "/", len(keys))
		if k%10 == 0 {
			wg.Wait()
		}
		//	ret, err := client.Importaddrs(v)
		//	if err != nil {
		//		fmt.Println(err.Error(), k, "/", len(keys))
		//		//goto st
		//	}
		//	retjson, _ := json.Marshal(ret)
		//	fmt.Println(v, string(retjson), k, "/", len(keys))

	}
	wg.Wait()
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
