package v3

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

const (
	mchAllBalanceCacheKey  = "bchsc_mchAllBalance_"
	mchAllBalanceExpireSec = 300

	mchAllBalanceCacheKeyV2  = "bchsc_mchAllBalancev2_"
	MchAllBalanceExpireSecV2 = 20
)

func GetMchBalance(c *gin.Context) {
	var balance decimal.Decimal
	sfrom := c.PostForm("sfrom")
	coinName := strings.ToLower(c.PostForm("coinName"))
	tokenName := strings.ToLower(c.PostForm("tokenName"))
	contractAddress := strings.ToLower(c.PostForm("contractAddress"))

	mchInfo, err := api.MchService.GetAppId(sfrom)
	if err != nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}
	if tokenName != "" || contractAddress != "" {
		//合约token查询
		if tokenName == "" {
			httpresp.HttpRespCodeError(c, httpresp.PARAM_ERROR, "选择合约查询，缺失tokenName", nil)
			return
		}
		if contractAddress == "" {
			httpresp.HttpRespCodeError(c, httpresp.PARAM_ERROR, "选择合约查询，缺失contractAddress", nil)
			return
		}
		balance, err = api.BalanceService.GetMchTokenBalance(tokenName, contractAddress, mchInfo.Id)

	} else {
		balance, err = api.BalanceService.GetMchBalance(coinName, mchInfo.Id)
	}

	if err != nil {
		if err.Error() != "Not Fount!" {
			log.Errorf("BalanceService error:%s", err.Error())
			httpresp.HttpRespErrorOnly(c)
			return
		}
	}
	var coinType string
	if tokenName != "" {
		coinType = tokenName
	} else {
		coinType = coinName
	}
	activityBalance, err := api.BalanceService.GetMchActivityBalance(coinType, contractAddress, mchInfo.Id)
	if err != nil {
		if err.Error() != "Not Fount!" {
			log.Errorf("BalanceService error:get activity error: %s", err.Error())
			httpresp.HttpRespErrorOnly(c)
			return
		}
	}

	// 2021-03-15 添加获取前20个地址的总余额
	topsTwentyBalance, err := api.BalanceService.GetTopsTwentyAddresses(coinType, contractAddress, mchInfo.Id)
	if err != nil {
		if err.Error() != "Not Fount!" {
			log.Errorf("BalanceService error:get tops twenty address amount error: %s", err.Error())
			httpresp.HttpRespErrorOnly(c)
			return
		}
	}

	result := &model.CoinBalance{
		CoinName:          coinName,
		Balance:           balance.String(),
		TokenName:         tokenName,
		ContractAddress:   contractAddress,
		ActivityBalance:   activityBalance.String(),
		TopsTwentyBalance: topsTwentyBalance.String(),
	}
	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
}

//
//func GetFromCache(chain, coin string, mchId int) (string, error) {
//	chain = strings.ToLower(chain)
//	coin = strings.ToLower(coin)
//	redisHelper, err := util.AllocRedisClient()
//	if err != nil {
//		return "", err
//	}
//	defer redisHelper.Close()
//
//	key := getCacheKeyV2(strconv.Itoa(mchId))
//	cacheData := getFromCache(redisHelper, key)
//	if cacheData == nil {
//		return "", errors.New("no cache data")
//	}
//	for _, c := range *cacheData {
//		if chain == c.CoinName && coin == c.TokenName {
//			return c.LiquidBalance, err
//		}
//	}
//	return "", nil
//}

func GetMchAllBalanceV2(c *gin.Context) {
	clientId := c.PostForm("client_id")
	mchVal := global.MchBaseInfo[clientId]
	if mchVal == nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}

	mchId := mchVal.AppId
	redisHelper, err := util.AllocRedisClient()
	defer redisHelper.Close()

	key := GetCacheKeyV2(strconv.Itoa(mchId))

	cacheData := GetFromCache(redisHelper, key)
	if cacheData != nil {
		log.Infof("GetMchAllBalanceV2 AppId=%d 从缓存获取数据成功", mchId)
		httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), *cacheData)
		return
	}

	log.Infof("GetMchAllBalanceV2 AppId=%d 从缓存获取数据失败，继续从DB获取", mchId)
	result, err := api.BalanceService.GetMchAllBalanceV2(mchId)
	if err != nil {
		log.Errorf("GetMchAllBalanceV2 error:%s", err.Error())
		httpresp.HttpRespErrorOnly(c)
		return
	}
	log.Infof("GetMchAllBalanceV2 AppId=%d 设置数据到缓存", mchId)
	SetToCache(redisHelper, key, MchAllBalanceExpireSecV2, &result)

	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
}

func GetMchAllBalance(c *gin.Context) {
	clientId := c.PostForm("client_id")
	mchVal := global.MchBaseInfo[clientId]
	if mchVal == nil {
		httpresp.HttpRespErrorOnly(c)
		return
	}

	mchId := mchVal.AppId

	redisHelper, _ := util.AllocRedisClient()
	defer redisHelper.Close()

	v2, _ := redisHelper.Get("balanceupdatev2")
	if v2 != "yes" {
		key := getCacheKey(strconv.Itoa(mchId))
		cacheData := GetFromCache(redisHelper, key)
		if cacheData != nil {
			log.Infof("GetMchAllBalance AppId=%d 从缓存获取数据成功", mchId)
			httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), *cacheData)
			return
		}
		result, err := api.BalanceService.GetMchAllBalance(mchId)
		if err != nil {
			log.Errorf("GetMchAllBalance error:%s", err.Error())
			httpresp.HttpRespErrorOnly(c)
			return
		}
		log.Infof("GetMchAllBalance AppId=%d 设置数据到缓存", mchId)
		SetToCache(redisHelper, key, mchAllBalanceExpireSec, &result)

		httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), result)
	} else {
		GetMchAllBalanceV2(c)
	}
}

func GetMchCoinMaxBalance(ctx *gin.Context) {
	clientId := ctx.PostForm("client_id")
	coinName := ctx.PostForm("coinName")
	tokenName := ctx.PostForm("tokenName")
	log.Infof("GetMchCoinMaxBalance 请求参数 clientId=%s coinName=%s tokenName=%s", clientId, coinName, tokenName)

	if clientId == "" {
		httpresp.HttpRespErrWithMsg(ctx, "clientId require")
		return
	}
	if coinName == "" {
		httpresp.HttpRespErrWithMsg(ctx, "coinName require")
		return
	}

	mchVal := global.MchBaseInfo[clientId]
	if mchVal == nil {
		httpresp.HttpRespErrWithMsg(ctx, "invalid clientId")
		return
	}

	coinName = strings.ToLower(coinName)
	tokenName = strings.ToLower(tokenName)

	balanceDecimal, addr, err := api.BalanceService.GetMchCoinMaxBalance(coinName, tokenName, mchVal.AppId)
	if err != nil {
		httpresp.HttpRespErrWithMsg(ctx, err.Error())
		return
	}
	result := model.CoinAddrMaxBalance{
		CoinName:  coinName,
		TokenName: tokenName,
		Balance:   balanceDecimal.String(),
		Address:   addr,
	}
	log.Infof("GetMchCoinMaxBalance 返回结果 %v", result)
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), result)
}

func GetFromCache(redisHelper *util.RedisClient, key string) *[]*model.CoinBalance {
	cacheStr, err := redisHelper.Get(key)
	if err != nil {
		log.Errorf("GetMchAllBalance redis.Get error:%s", err.Error())
		return nil
	}
	cacheByte := []byte(cacheStr)
	if cacheByte == nil {
		return nil
	}

	m := &[]*model.CoinBalance{}
	if err := json.Unmarshal(cacheByte, m); err != nil {
		log.Errorf("GetMchAllBalance redis Unmarshal error:%s", err.Error())
		return nil
	}
	return m
}

func SetToCache(redisHelper *util.RedisClient, key string, expire int64, data *[]*model.CoinBalance) {
	ms, err := json.Marshal(data)
	if err != nil {
		log.Errorf("setToCache json.Marshal error:%v", err)
		return
	}

	err = redisHelper.Set(key, string(ms))
	if err != nil {
		log.Errorf("setToCache redis Set error %v", err)
		return
	}
	err = redisHelper.Expire(key, expire)
	if err != nil {
		log.Errorf("setToCache redis Set error %v", err)
		return
	}
}

func getCacheKey(appID string) string {
	return mchAllBalanceCacheKey + appID
}

func GetCacheKeyV2(appID string) string {
	return mchAllBalanceCacheKeyV2 + appID
}
