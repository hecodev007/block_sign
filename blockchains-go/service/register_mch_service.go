package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"strings"
)

var (
	TESTAPIID = []int{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18}
	LINEAPIID = []int{1,2,3,8,9,10,11,17,18}
)

//RegisterFcMchService 托管后台注册商户
func RegisterFcMchService(req model.RegisterMchRequest)(mch interface{} ,err error) {
	//插入
	apiKey := uuid.New().String()+uuid.New().String()
	apiKey = strings.Replace(apiKey,"-","",-1)
	apiSecret := uuid.New().String()
	mchItem := entity.FcMch{
		Platform :req.Name,
		LoginName: req.Phone,
		ApiKey   :apiKey,
		ApiSecret:apiSecret,
		Status:2,
	}
	tx := db.Conn.NewSession()
	tx.Begin()
	err = dao.InsertNewMchItem(tx,mchItem)
	if err != nil {
		tx.Rollback()
		return nil,err
	}
	newMch,err := dao.CustodyFcMchFindByApikey(tx,apiKey)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	mchProfileItem := entity.FcMchProfile{
		MchId: newMch.Id,
		FirstName:req.Name,
		LastName:req.Name,
		Email:req.Email,
		Mobile:req.Phone,
		CompanyImg:req.CompanyImg,
	}
	err = dao.InsertNewMchProfileItem(tx,mchProfileItem)
	if err != nil {
		tx.Rollback()
		return nil,err
	}
	mchMoneyItem := entity.FcMchMoney{
		AppId:newMch.Id,
		Address : uuid.New().String(),
		Amount:"0",
		AmountFreeze:"0",
		Status: 1,
	}
	err = dao.InsertNewMchMoneyItem(tx,mchMoneyItem)
	if err != nil {
		tx.Rollback()
		return nil,err
	}
	err = tx.Commit()
	l := len(apiSecret)
	apiSecretBack := apiSecret[0:4] + " **** **** " + apiSecret[l-4:]
	backData := map[string]string{
		"client_id":apiKey,
		"secret":apiSecretBack,
	}
	return backData,err
}


func ReSetFcMchSecretService(req model.RegisterMchRequest)(mch interface{} ,err error) {
	//插入
	apiSecret := uuid.New().String()
	mchItem := entity.FcMch{
		ApiKey   :req.ApiKey,
		ApiSecret: apiSecret,
	}
	err = dao.CustodyFcMchUpdateByApikey(mchItem)
	if err != nil {
		return nil,err
	}
	l := len(apiSecret)
	apiSecretBack := apiSecret[0:4] + " **** **** " + apiSecret[l-4:]
	backData := map[string]string{
		"client_id":req.ApiKey,
		"secret":apiSecretBack,
	}
	return backData,err
}

func SearchFcMchSecretService(req model.RegisterMchRequest)(mch interface{} ,err error) {

	var mchItem *entity.FcMch
	tx := db.Conn.NewSession()
	mchItem,err = dao.CustodyFcMchFindByApikey(tx, req.ApiKey)
	if err != nil {
		return nil, err
	}
	l := len(mchItem.ApiSecret)
	apiSecretBack := mchItem.ApiSecret[0:4] + " **** **** " + mchItem.ApiSecret[l-4:]
	backData := map[string]string{
		"client_id":req.ApiKey,
		"secret":apiSecretBack,
	}
	return backData, err
}

//绑定地址api_power
func BindAddressService(req model.BindMchRequest)(err error) {
	//查询商户
	var mInfo *entity.FcMch
	mInfo,err = dao.FcMchFindByApikey(req.ApiKey)
	if err != nil {
		log.Error(err)
		return
	}
	if mInfo.Id == 0 {
		err = fmt.Errorf("商户不存在")
		log.Error(err)
		return
	}
	//查询币种
	coinArr := strings.Split(req.CoinName,",")
	for _,item := range coinArr {
		cInfo,err1 := dao.FcCoinSetGetByName(item,1)
		if err1 != nil || cInfo.Id == 0 {
			err = fmt.Errorf("%v币不支持",req.CoinName)
			return err
		}
		err = dao.InsertCustobyPowerItem(TESTAPIID,cInfo.Id,mInfo.Id,item,req.Address,req.WhiteIp)
	}
	return err
}
