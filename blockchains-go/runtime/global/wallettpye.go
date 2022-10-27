package global

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"strings"
)

//key 币种
// value cold or hot
var WalletType func(string, int) status.WalletType

var waType map[string]status.WalletType

const WalletTypeCoin = "wallettype_coin"
const WalletTypeMch = "wallettype_mch"

func InitWalletType() {
	//币种交易模型（非币种结构模型，比如BTM需要走账户模型交易流程）
	waType = make(map[string]status.WalletType, 0)
	redis, err := util.AllocRedisClient()
	if err != nil {
		log.Info("获取redis错误: ", err.Error())
	} else {
		defer redis.Close()
	}

	//初始化走冷签流程的商户, 主要为了测试
	redis.Set(fmt.Sprintf("%s_%d", WalletTypeMch, 152), status.WalletType_Cold) //goapi
	redis.Set(fmt.Sprintf("%s_%d", WalletTypeMch, 102), status.WalletType_Cold) //63

	for _, v := range conf.Cfg.WalletType.Cold {
		waType[v] = status.WalletType_Cold
		if redis != nil {
			redis.Set(fmt.Sprintf("%s_%s", WalletTypeCoin, v), status.WalletType_Cold)
		}
	}
	for _, v := range conf.Cfg.WalletType.Hot {
		waType[v] = status.WalletType_Hot
		if redis != nil {
			redis.Set(fmt.Sprintf("%s_%s", WalletTypeCoin, v), status.WalletType_Hot)
		}
	}
	WalletType = getWalletType
	log.Info("加载币种钱包类型")
}

func getWalletType(coin string, mchId int) status.WalletType {
	if strings.ToLower(coin) == "hsc" || strings.ToLower(coin) == "bsc" || strings.ToLower(coin) == "heco" {
		redis, err := util.AllocRedisClient()
		if err != nil {
			log.Info("redis链接失败, 使用配置文件. 错误: ", err.Error())
			return waType[coin]
		}
		defer redis.Close()

		mchTypes, err := redis.Get(fmt.Sprintf("%s_%d", WalletTypeMch, mchId))
		if err != nil {
			log.Info("redis获取mch类型失败, 使用配置文件. 错误: ", err.Error())
			return waType[coin]
		}
		switch mchTypes {
		case "cold":
			return status.WalletType_Cold
		case "hot":
			return status.WalletType_Hot
		}

		log.Infof("当前商户 %d 无配置冷热签类型, 走默认的币种 %s 冷热签类型 ", mchId, coin)

		coinTypes, err := redis.Get(fmt.Sprintf("%s_%s", WalletTypeCoin, coin))
		if err != nil {
			log.Info("redis获取coin类型失败, 使用配置文件. 错误: ", err.Error())
			return waType[coin]
		}

		switch coinTypes {
		case "cold":
			return status.WalletType_Cold
		case "hot":
			return status.WalletType_Hot
		default:
			return waType[coin]
		}
	}
	return waType[coin]
}
