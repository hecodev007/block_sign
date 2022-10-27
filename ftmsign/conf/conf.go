package conf

/*
使用toml进行配置
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	// log "log"
	"path/filepath"
	"sync"
)

var (
	Config *tomlConfig
	once   sync.Once
)

func InitConfig() {
	once.Do(func() {
		filePath, err := filepath.Abs("./conf/app.toml")
		if err != nil {
			// log.Println(err)
			panic(err)
		}
		if _, err := toml.DecodeFile(filePath, &Config); err != nil {
			// log.Printf("Read config file error,Err=[%v]",err)
			panic(fmt.Sprintf("Read config file error,Err=[%v]", err))
		}
	})
}

type tomlConfig struct {
	Debug         bool   `toml:"debug"`
	Port          string `toml:"port"`
	CoinType      string `toml:"coinType"`
	FilePath      string `toml:"filePath"`
	IsStartThread bool   `toml:"isStartThread"`
	MchId         string `toml:"mchId"`
	OrderId       string `toml:"orderId"`
	Version       string `toml:"version"`
	AuthCfg       struct {
		Encrypt  bool   `toml:"encrypt"`
		Enable   bool   `toml:"enable"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"auth"`

	EthCfg struct {
		NodeUrl         string `toml:"nodeUrl"`
		User            string `toml:"user"`
		Password        string `toml:"password"`
		GasPrice        int64  `toml:"gasPrice"`
		GasLimit        int64  `toml:"gasLimit"`
		MaxGasPriceGwei int64  `toml:"maxGasPriceGwei"`
		MinGasPriceGwei int64  `toml:"minGasPriceGwei"`
		NetWorkId       int64  `toml:"networkid"`
	} `toml:"ftm"`
}
