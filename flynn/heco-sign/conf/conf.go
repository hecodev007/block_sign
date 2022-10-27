package conf

/*
使用toml进行配置
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	//log "github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
)

var (
	Config *tomlConfig
	once   sync.Once
)

func InitConfig() {
	once.Do(func() {
		filePath, err := filepath.Abs("./conf/config.toml")
		if err != nil {
			//log.Error(err)
			panic(err)
		}
		if _, err := toml.DecodeFile(filePath, &Config); err != nil {
			//log.Errorf("Read config file error,Err=[%v]",err)
			panic(fmt.Sprintf("Read config file error,Err=[%v]", err))
		}
	})
}

type tomlConfig struct {
	Debug         bool   `toml:"debug"`
	Port          string `toml:"port"`
	CoinType      string `toml:"coinType"`
	WalletType    string `toml:"walletType"` //hot or cold wallet
	FilePath      string `toml:"filePath"`
	IsStartThread bool   `toml:"isStartThread"`
	IsStartValid  bool   `toml:"isStartValid"`
	MchId         string `toml:"mchId"`
	OrderId       string `toml:"orderId"`
	Version       string `toml:"version"`
	AuthCfg       struct {
		Enable   bool   `toml:"enable"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"auth"`

	HecoCfg struct {
		NodeUrl   string `toml:"nodeUrl"`
		User      string `toml:"user"`
		Password  string `toml:"password"`
		GasPrice  int64  `toml:"gasPrice"`
		GasLimit  int64  `toml:"gasLimit"`
		NetWorkId int64  `toml:"networkid"`
	} `toml:"heco"`
}
