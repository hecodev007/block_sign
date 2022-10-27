package conf

/*
使用toml进行配置
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	// "github.com/bsc-sign/util/log"
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
			// log.Error(err)
			panic(err)
		}
		if _, err := toml.DecodeFile(filePath, &Config); err != nil {
			// log.Errorf("Read config file error,Err=[%v]",err)
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

	Log struct {
		Level     string `toml:"level"       json:"level"`
		Formatter string `toml:"formatter"   json:"formatter"`
		OutFile   string `toml:"outfile"    json:"outfile"`
		ErrFile   string `toml:"errfile"    json:"errfile"`
	} `toml:"log"`

	OrderKeeper struct {
		CacheExpirationSec int64 `toml:"cacheExpirationSec"`
		KeeperSize         int64 `toml:"keeperSize"`
	} `toml:"orderkeeper"`

	ChainCfg struct {
		NodeUrl               string `toml:"nodeUrl"`
		User                  string `toml:"user"`
		Password              string `toml:"password"`
		GasPrice              int64  `toml:"gasPrice"`
		GasLimit              int64  `toml:"gasLimit"`
		MaxGasPriceGwei       int64  `toml:"maxGasPriceGwei"`
		MinGasPriceGwei       int64  `toml:"minGasPriceGwei"`
		GasPriceExpirationSec int64  `toml:"gasPriceExpirationSec"`
		NetWorkId             int64  `toml:"networkid"`
	} `toml:"chain"`

	Callback struct {
		Url      string `toml:"url"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"callback"`

	Redis struct {
		Addr string `toml:"addr"`
		Pwd  string `toml:"pwd"`
		DB   int    `toml:"db"`
	} `toml:"redis"`

	Secret struct {
		Salt string `toml:"salt"`
	} `toml:"secret"`

	IMBot struct {
		DingErrorToken string `toml:"dingErrorToken"` // 钉钉工具token 异常使用
		DingWarnToken  string `toml:"dingWarnToken"`  // 钉钉工具token 警告使用
		DingInfoToken  string `toml:"dingInfoToken"`  // 钉钉工具token 信息使用
	} `toml:"im"`

	Gas struct {
		Special string `toml:"special"`
	} `toml:"gas"`
}
