package main

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/ltcserver/util"
)

//客户端导入脚本，追踪地址
func main() {

	//addr1, err := util.ReadCsv("./btc_d.csv", 0)
	//if err != nil {
	//	fmt.Println(err)
	//	panic(err)
	//}
	//addr2, err := util.ReadCsv("./btc_user.csv", 0)
	//if err != nil {
	//	fmt.Println(err)
	//	panic(err)
	//}
	//addr1 = append(addr1, addr2...)
	//log.Println("csv addrs numbers:", len(addr1))
	//
	//addrs := util.StringArrayRemoveRepeatByMap(addr1)
	//addrs := []string{"3F6h8K97WjiEhYyLa7F33isuqRw5eRhFQV"}

	addrs, err := util.ReadCsv("./usdt.csv", 0)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	arr, _ := json.Marshal(addrs)
	fmt.Println(string(arr))

}
