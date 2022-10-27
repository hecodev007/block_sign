package global

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"strings"
)

//数据库币种配置，初始化程序直接读取数据库设置（包含精度，大小阈值）
var CoinDecimal map[string]*entity.FcCoinSet
var CoinIdMap map[int]*entity.FcCoinSet

//初始化和刷新都用该方法
//todo 暂时由业务系统新增，更新，修改的时候触发远程api
func InitCoinDecimal() {
	coins := make(map[string]*entity.FcCoinSet, 0)
	CoinIdMap = map[int]*entity.FcCoinSet{}
	result, err := dao.FcCoinSetFindByStatus(1)
	if err != nil {
		panic(err.Error())
	} else {
		if len(result) == 0 {
			panic("缺少币种相关设置")
		}
	}

	for k, v := range result {
		coins[strings.ToLower(v.Name)] = v
		//db设计问题，这里不得不写硬代码，存在同名合约
		if strings.HasPrefix(v.Token, "bsc:") {
			result[k].Token = strings.ReplaceAll(v.Token, "bsc:", "")
			coins[strings.ToLower(v.Name)] = result[k]
		}
		if v.Decimal < 0 {
			panic("币种精度异常，出现负数")
		}

		CoinIdMap[v.Id] = v
	}
	CoinDecimal = coins
	log.Info("加载币种配置名单")
}
