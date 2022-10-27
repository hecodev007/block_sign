package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/group-coldwallet/scanning-service/utils"
)

type InsertAddressConfig struct {
	Mode      string `toml:"mode"`
	CoinName  string `toml:"coin_name"`
	UserId    int64  `toml:"user_id"`
	CsvPath   string `toml:"csv_path"`
	DataBases map[string]DatabaseConfig
}

func LoadInsertAddressConfig(cfgPath string, cfg *InsertAddressConfig) error {

	if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
		fmt.Println(err)
		return err
	}
	if cfg.Mode == "prod" {
		for k, v := range cfg.DataBases {
			v.Name, _ = utils.AesBase64Str(v.Name, DefaulAesKey, false)
			v.Url, _ = utils.AesBase64Str(v.Url, DefaulAesKey, false)
			v.User, _ = utils.AesBase64Str(v.User, DefaulAesKey, false)
			v.PassWord, _ = utils.AesBase64Str(v.PassWord, DefaulAesKey, false)
			cfg.DataBases[k] = v
		}
		//cfg.Push.MqUrl, _ = utils.AesBase64Str(cfg.Push.MqUrl, DefaulAesKey, false)
	}
	return nil
}
