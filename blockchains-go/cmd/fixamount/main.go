package main

import (
	"flag"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/shopspring/decimal"
	"log"
)

func main() {
	var (
		//	err        error
		routineNum int
		cfgFile    string
		cfg        conf.CollectConfig
	)

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("process exit err : %v \n", err)
		}
	}()

	flag.IntVar(&routineNum, "n", 10, "each cpu's routine num")
	flag.StringVar(&cfgFile, "c", "conf/app.toml", "set the toml config file")
	flag.Parse()

	//日志
	//util.InitLog()
	if err := conf.LoadConfig3(cfgFile, &cfg); err != nil {
		panic(err)
	}
	conf.DecryptCfg(&cfg)
	//数据库加载
	db.InitOrm2(cfg.DB)

	orders := make([]*entity.FcOrderHot, 0)
	err := db.Conn.Table("fc_order_hot").
		Where("coin_name = ? and outer_order_no like ?", "cds", "COLLECT_%").Find(&orders)
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Print("len:", len(orders))
		for _, v := range orders {
			amount, _ := decimal.NewFromString(v.Quantity)
			_, err = db.Conn.Table("fc_order_hot").ID(v.Id).Update(map[string]interface{}{"quantity": amount.Shift(18).String()})
			if err != nil {
				log.Print(err.Error())
			}
		}
		log.Println("server showdown !")
	}

}
