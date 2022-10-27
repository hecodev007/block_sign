package global

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
	"strings"
)

//商户访问权限，ip白名单设置，api状态是否可用
//原数据库设计问题，只能把这些功能集中起来
//todo 暂时由业务系统新增，更新，修改的时候触发远程api

//key为path
type ApiAuth struct {
	CallBackUrl string //币种回调路径
	IP          []string
}

//api路径的设置,key为api路径名，value为相关设置
type MchApi struct {
	StartTime int64               //有效开始时间 unix time
	EndTime   int64               //失效时间 unix time
	Auth      map[string]*ApiAuth //key为path路径
}

type MchCoin struct {
	StartTime int64              //有效开始时间 unix time
	EndTime   int64              //失效时间 unix time
	Api       map[string]*MchApi //币种对应的设置 key为币种名
}

//全局变量，给外部调用,key 为商户ID
var MchAuth map[int]*MchCoin

//初始化商户购买路径的权限
func InitMchApiAuth() {
	mchauth := make(map[int]*MchCoin, 0)
	mchs, err := dao.FcMchFindsVaild()
	if err != nil {
		panic(err.Error())
	}
	if len(mchs) == 0 {
		panic("没有相关有效商户记录")
	}

	for _, v := range mchs {
		//查询购买的币种
		coins, err := dao.FcMchServiceFindsValid(v.Id)
		if err != nil {
			//log.Errorf("商户：%s,查询购买的币种异常：%s", v.Platform, err.Error())
		}
		if len(coins) == 0 {
			//log.Infof("商户：%s,没有购买相关币种", v.Platform)
			continue
		}
		coinMap := make(map[string]*MchApi, 0)
		for _, coin := range coins {
			//查询相关路径的校验设置：目前为ip
			apiAuths, err := dao.FcApiPowerFindsValidCoin(coin.CoinId, v.Id)
			if err != nil {
				log.Errorf("商户：%s,查询api power数据异常：%s", v.Platform, err.Error())
			}
			if len(apiAuths) == 0 {
				log.Infof("商户：%s,coin：%s 缺少api power的相关配置", v.Platform, coin.CoinName)
				continue
			}
			apiMap := make(map[string]*ApiAuth, 0)
			for _, apiVal := range apiAuths {
				//查询apiid对应的路径名
				urlData, err := dao.FcApiListGetByID(apiVal.ApiId)
				if err != nil {
					if err.Error() != "Not Fount!" {
						log.Errorf("商户：%s,查询api list数据异常,api Id :%d,err:%s", v.Platform, apiVal.ApiId, err.Error())
					}
					continue
				}
				ips := strings.Split(apiVal.Ip, ",")
				if apiVal.Ip == "" {
					//如果没有设置IP 不允许访问，忽略掉这个path
					continue
				}
				apiMap[urlData.ApiSuffix] = &ApiAuth{
					CallBackUrl: apiVal.Url,
					IP:          ips,
				}
			}
			//币种对应的url的ip名单设置
			mchApi := &MchApi{
				StartTime: coin.StartTime,
				EndTime:   coin.EndTime,
				Auth:      apiMap,
			}
			//币种对应的设置
			coinMap[coin.CoinName] = mchApi
			if coin.CoinName == "uca" {
				log.Info(v.Platform + "添加uca")
			}
			//商户币种信息
			mchauth[v.Id] = &MchCoin{
				StartTime: coin.StartTime,
				EndTime:   coin.EndTime,
				Api:       coinMap,
			}
		}

	}
	log.Infof("加载权限配置的商户名单数量：%d", len(mchauth))
	MchAuth = mchauth
	//for k, v := range MchAuth {
	//	for k1, v1 := range v.Api {
	//		for k2, v2 := range v1.Auth {
	//			data, _ := json.Marshal(v2.IP)
	//			log.Infof("加载权限配置商户id:%d,币种：%s,路径：%s,允许的ip：%s", k, k1, k2, string(data))
	//		}
	//	}
	//}

}
