package service

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/userAddr"
	"strings"
	"time"

	//"custody-merchant-admin/middleware/verify"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/module/log"
	"fmt"
)

// CreateNewChainAddress 新建业务线时 创建地址
/*
每个业务线对应一个钱包商户mchid
coinName逗号拼接的多币种字符串
*/
func CreateNewChainAddress(businessId int64, accountClientId string, coinArr []string) (address map[string]string, err error) {
	//根据业务线查询clientid
	if accountClientId == "" {
		ssInfo := serviceSecurity.NewEntity()
		ssInfo.FindItemByBusinessId(businessId)
		accountClientId = ssInfo.ClientId
	}
	if accountClientId == "" {
		err = fmt.Errorf("商户clientId为空")
		return
	}
	chainArr := make([]string, 0)
	for _, item := range coinArr {
		chainName := FindChainName(item)
		chainArr = append(chainArr, chainName)
	}
	address = make(map[string]string)
	////TODO:钱包地址不足
	//for _, coin := range coinArr {
	//	address[coin] = "钱包地址不足"
	//}
	//return

	log.Errorf("创建钱包地址req coinArr :%+v", chainArr)
	log.Errorf("创建钱包地址req accountClientId :%+v", accountClientId)
	var lists map[string]string
	lists, err = blockChainsApi.BlockChainCreateLotCoinAddress(chainArr, accountClientId, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		log.Errorf("创建钱包地址req err :%v", err)
		err = global.WarnMsgError(global.DataWarnCreateAddressErr)
		return
	}
	if len(lists) < 1 {
		log.Errorf("创建钱包地址不足 err :%v", err)
		err = global.WarnMsgError(global.DataWarnCreateAddressErr)
		return
	}
	return lists, err
}

//func BlockChainBindAddress

//CreateBatchChainAddress 批量创建地址
func CreateBatchChainAddress(businessId int64, coinName string, num int) (address []string, err error) {
	//根据业务线查询clientid
	ssInfo := serviceSecurity.NewEntity()
	ssInfo.FindItemByBusinessId(businessId)
	if ssInfo.ClientId == "" {
		log.Errorf("创建钱包地址sql err :%v", err)
		return address, global.WarnMsgError(global.DataWarnNoMchComboErr)
	}

	//判断是否是代币
	coinName = FindChainName(coinName)
	address, err = blockChainsApi.BlockChainBatchCreateAddress(num, ssInfo.ClientId, coinName, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		log.Errorf("创建钱包地址req err :%v", err)
		return address, global.WarnMsgError(global.DataWarnCreateAddressErr)
	}

	return
}

//CreateUserAddress 用户地址存入数据库
func CreateUserAddress(aInfo *domain.InsertAddrInfo) error {
	//根据业务线查询clientid
	addr := userAddr.NewEntity()
	addr.MerchantId = aInfo.MerchantId
	addr.MerchantUser = aInfo.MerchantUser
	addr.CoinId = aInfo.CoinId
	addr.Address = aInfo.Address
	addr.ClinetId = aInfo.ClientId
	addr.SecureKey = aInfo.SecureKey
	addr.ServiceId = aInfo.ServiceId
	addr.ChainId = aInfo.ChainId
	addr.State = 0
	addr.CreatedAt = time.Now().Local()
	err := addr.CreateUserAddress()
	if err != nil {
		return err
	}
	return nil
}

func FindChainName(coinName string) string {
	key := fmt.Sprintf("custody:chainname:%v", coinName)
	var v string
	cache.GetRedisClientConn().Get(key, &v)
	if v != "" {
		v = strings.ToUpper(v)
		return v
	}

	//coinName = FindChainName(coinName)
	cInfo, _ := base.FindCoinsByName(coinName)
	if cInfo.ChainId != 0 {
		chainInfo, _ := base.FindChainsById(cInfo.ChainId)
		if chainInfo.Name != "" {
			coinName = chainInfo.Name
		}
	}
	coinName = strings.ToUpper(coinName)
	cache.GetRedisClientConn().Set(key, coinName, 0*time.Second)
	return coinName
}
