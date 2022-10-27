package global

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/log"
)

var MchBaseInfo map[string]*BaseInfo

type BaseInfo struct {
	AppId   int    //appid
	MchName string //商户名
}

//初始化商户基本信息
func InitMchBaseInfo() {
	MchBaseInfo = make(map[string]*BaseInfo, 0)
	mchs, err := dao.FcMchFindsVaild()
	if err != nil {
		panic(err.Error())
	}
	if len(mchs) == 0 {
		panic("没有相关有效商户记录")
	}
	for _, v := range mchs {
		log.Infof("InitMchBaseInfo v.ApiKey=%s v.Platform=%s", v.ApiKey, v.Platform)
		MchBaseInfo[v.ApiKey] = &BaseInfo{
			AppId:   v.Id,
			MchName: v.Platform,
		}
	}
}

func ReloadMchBaseInfo() {
	MchBaseInfo = make(map[string]*BaseInfo, 0)
	mchs, err := dao.FcMchFindsVaild()
	if err != nil {
		panic(err.Error())
	}
	if len(mchs) == 0 {
		panic("没有相关有效商户记录")
	}
	for _, v := range mchs {
		MchBaseInfo[v.ApiKey] = &BaseInfo{
			AppId:   v.Id,
			MchName: v.Platform,
		}
	}
}
