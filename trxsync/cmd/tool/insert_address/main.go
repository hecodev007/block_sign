package main

import (
	"flag"
	"fmt"
	"github.com/group-coldwallet/trxsync/conf"
	"github.com/group-coldwallet/trxsync/db"
	"github.com/group-coldwallet/trxsync/services/tools/insert_address"
	"log"
)

var (
	configPath string
	cfg        conf.InsertAddressConfig
)

func init() {
	flag.StringVar(&configPath, "f", "", "配置文件路径")
}
func main() {
	flag.Parse()
	if configPath == "" {
		log.Fatal("配置文件路径为空，请使用cli参数[-f=]去配置")
		return
	}
	fmt.Println(configPath)
	//初始化配置文件
	err := conf.LoadInsertAddressConfig(configPath, &cfg)
	if err != nil {
		log.Fatal("加载配置文件失败：", err)
		return
	}
	fmt.Println("开始连接数据库")
	//初始化数据库
	err = db.InitUserDB(cfg.DataBases["user"])
	if err != nil {
		log.Fatal("初始化数据库错误：", err)
		return
	}
	log.Println("初始化数据库成功过！！！")
	//操作插入地址服务
	err = insert_address.InsertAddressToDB(&cfg)
	if err != nil {
		log.Fatal("插入地址错误：", err)
		return
	}
}
