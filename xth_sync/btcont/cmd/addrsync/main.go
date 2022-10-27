package main

import (
	"btcont/common/db"
	"btcont/common/log"
	"btcont/common/model"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			coin_name := os.Args[i]
			addrlist := new(model.FcGenerateBeforeAddressList).List(coin_name)
			log.Info(coin_name, len(addrlist))
			if len(addrlist) == 0 {
				continue
			}
			affected := 0
			addres := new(model.Addresses).Sets(addrlist)
			for k, v := range addres {
				insertId, err := db.AddrmanageConn.InsertOne(v)
				if err != nil {
					//log.Info(err.Error())
					continue
				}
				log.Info(insertId, v.Id, k, len(addrlist))
				affected++
			}

			log.Info(coin_name, "成功", affected)
		}
	}
}
