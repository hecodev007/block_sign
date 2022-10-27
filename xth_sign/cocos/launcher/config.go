package launcher

import (
	"github.com/spf13/viper"
	"github.com/astaxie/beego/logs"
)

func InitConfig() {
	configPath := `./conf/`
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		logs.Debug(err.Error())
		panic(err)
	}
}
