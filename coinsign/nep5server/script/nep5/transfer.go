package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/nep5server/util/neputil"
	"github.com/shopspring/decimal"
	"log"
)

type TransferParams struct {
	From          string          `toml:"fromaddr"` //
	FromKey       string          `toml:"privekey"` //
	To            string          `toml:"toaddr"`   //
	ToAmountFloat decimal.Decimal `toml:"toamount"` //
	Token         string          `toml:"token"`
}

//初始化配置文件
func InitConf(confName string) *TransferParams {
	cfg := new(TransferParams)
	if confName == "" {
		confName = "application.toml"
	}
	configFile := fmt.Sprintf("%s", confName)
	if _, err := toml.DecodeFile(configFile, cfg); err != nil {
		panic(err)
	}
	return cfg
}

//CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build transfer.go
//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build transfer.go
func main() {
	params := InitConf("/Users/hoo/workspace/go/gopath/src/github.com/group-coldwalle/nep5server/script/nep5/application.toml")
	if params.From == "" && params.FromKey == "" && params.Token == "" &&
		params.To == "" && params.ToAmountFloat.LessThanOrEqual(decimal.Zero) {
		log.Panic("参数异常")
	}
	raw, txid, err := neputil.Nep5Transfer(
		params.From,
		params.To,
		params.FromKey,
		params.Token,
		//"3e09e602eeeb401a2fec8e8ea137d59aae54a139",
		//"ab38352559b8b203bde5fddfa0b07d8b2525e132",
		params.ToAmountFloat.Shift(8).IntPart(),
	)

	//旧币
	//ab38352559b8b203bde5fddfa0b07d8b2525e132
	if err != nil {
		log.Panic(err.Error())
	}
	log.Println("raw:=====>", raw)
	log.Println("txid:=====>", txid)

}
