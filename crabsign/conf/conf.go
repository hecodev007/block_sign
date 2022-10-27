package conf

/*
使用toml进行配置
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"wallet-sign/util"

	//log "github.com/sirupsen/logrus"
	"path/filepath"
	"sync"
)

const (
	aesKey = "viXLnRAAiiUf2YMo27my/CmWYr807SFd"
)

var (
	Config *tomlConfig
	once   sync.Once
)

func InitConfig() {
	once.Do(func() {

		filePath, err := filepath.Abs("conf/app.toml")
		if err != nil {
			panic("do not find any config file")
		}
		if _, err := toml.DecodeFile(filePath, &Config); err != nil {
			//log.Errorf("Read config file error,Err=[%v]",err)
			panic(fmt.Sprintf("Read config file error,Err=[%v]", err))
		}
		//解密 auth
		if Config.AuthCfg.Encrypt {
			err = decryptAuthConfig(Config)
			if err != nil {
				panic(fmt.Sprintf("decode auth info error,Err=%v", err))
			}
		}
	})
}

func decryptAuthConfig(cfg *tomlConfig) error {
	user, err := util.AesBase64Crypt([]byte(cfg.AuthCfg.User), []byte(aesKey), false)
	if err != nil {
		return err
	}
	cfg.AuthCfg.User = string(user)
	password, err := util.AesBase64Crypt([]byte(cfg.AuthCfg.Password), []byte(aesKey), false)
	if err != nil {
		return err
	}
	cfg.AuthCfg.Password = string(password)
	return nil
}

type tomlConfig struct {
	Debug         bool   `toml:"debug"`
	Port          string `toml:"port"`
	CoinType      string `toml:"coinType"`
	WalletType    string `toml:"walletType"` //hot or cold wallet
	FilePath      string `toml:"filePath"`
	IsStartThread bool   `toml:"isStartThread"`
	MchId         string `toml:"mchId"`
	OrderId       string `toml:"orderId"`
	Version       string `toml:"version"`
	NodeUrl       string `toml:"nodeUrl"`
	AuthCfg       struct {
		Encrypt  bool   `toml:"encrypt"`
		Enable   bool   `toml:"enable"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"auth"`

	RedisConfig struct {
		Cluster bool   `toml:"cluster"`
		Addr    string `toml:"addr"`
		Pwd     string `toml:"pwd"`
	} `toml:"redis"`

	GxcCfg struct {
		ChainId string `toml:"chainId"`
		NodeUrl string `toml:"nodeUrl"`
		MemoKey string `toml:"memoKey"`
	} `toml:"gxc"`
	ARCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"ar"`
	HntCfg struct {
		NodeUrl  string `toml:"nodeUrl"`
		LockTime int64  `toml:"lockTime"`
	} `toml:"hnt"`
	CdsCfg struct {
		NodeUrl   string `toml:"nodeUrl"`
		User      string `toml:"user"`
		Password  string `toml:"password"`
		GasPrice  int64  `toml:"gasPrice"`
		NetWorkId int    `toml:"networkid"`
	} `toml:"cds"`
	KsmCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"ksm"`
	CrabCfg struct {
		NodeUrl string `toml:"nodeUrl"`
		WsUrl   string `toml:"wsUrl"`
	} `toml:"crab"`
	NearCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"near"`
	CocosCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"cocos"`
	FioCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"fio"`
	DotCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"dot"`
	AzeroCfg struct {
		RpcUrl  string `toml:"rpcUrl"`
		WsUrl   string `toml:"wsUrl"`
		ScanUrl string `toml:"scanUrl"`
		ScanKey string `toml:"scanKey"`
	} `toml:"azero"`
	FisCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"fis"`
	OriCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"ori"`
	TkmCfg struct {
		NodeUrl string `toml:"nodeUrl"`
	} `toml:"tkm"`
	SolCfg struct {
		NodeUrl  string `toml:"nodeUrl"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"sol"`
	PcxCfg struct {
		NodeUrl string `toml:"nodeUrl"`
		WsUrl   string `toml:"wsUrl"`
	} `toml:"pcx"`
	BscCfg struct {
		NodeUrl   string `toml:"nodeUrl"`
		User      string `toml:"user"`
		Password  string `toml:"password"`
		GasPrice  int64  `toml:"gasPrice"`
		GasLimit  int64  `toml:"gasLimit"`
		NetWorkId int64  `toml:"networkid"`
	} `toml:"bsc"`
	HecoCfg struct {
		NodeUrl   string `toml:"nodeUrl"`
		User      string `toml:"user"`
		Password  string `toml:"password"`
		GasPrice  int64  `toml:"gasPrice"`
		GasLimit  int64  `toml:"gasLimit"`
		NetWorkId int64  `toml:"networkid"`
	} `toml:"heco"`
	CphCfg struct {
		NodeUrl   string `toml:"nodeUrl"`
		User      string `toml:"user"`
		Password  string `toml:"password"`
		GasPrice  int64  `toml:"gasPrice"`
		NetWorkId int    `toml:"networkid"`
	} `toml:"cph"`
	TrxCfg struct {
		NodeUrl  string   `toml:"nodeUrl"`
		BackUrls []string `toml:"backUrls"`
		User     string   `toml:"user"`
		Password string   `toml:"password"`
	} `toml:"trx"`
	DipCfg struct {
		NodeUrl  string `toml:"nodeUrl"`
		ApiUrl   string `toml:"apiUrl"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"dip"`
	XtzCfg struct {
		NodeUrl  string `toml:"nodeUrl"`
		User     string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"xtz"`
}
