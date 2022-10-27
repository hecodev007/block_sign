package services

import (
	"brisesign/conf"
	"brisesign/redis"
	"log"
	"math/big"
	"time"
)

const (
	feeCacheKey = "bsc_sign_fee"
)

func (s *ChainService) confirmFee() *big.Int {
	value, err := redis.Client.Get(feeCacheKey)
	if err != nil {
		log.Printf("从redis获取手续费出错: %s", err.Error())
	}

	if value != "" {
		// 缓存找到，直接返回

		feeInCache, ok := new(big.Int).SetString(value, 10)
		if !ok {
			log.Printf("从Redis获取的Fee转换成big.int出错")
			// 清除这个格式有误的缓存
			if errDel := redis.Client.Del(feeCacheKey); errDel != nil {
				log.Printf("删除redis出错：%v", errDel)
			}
			return big.NewInt(conf.Config.ChainCfg.GasPrice)
		}
		log.Printf("使用从缓存读取的手续费 %s", feeInCache.String())
		return feeInCache
	}

	log.Printf("没有从缓存获取到数据，直接从链上获取")
	fee := adjustFee(s.getFeeFromChain())
	if errSet := redis.Client.Set(feeCacheKey, fee.String(), time.Duration(conf.Config.ChainCfg.GasPriceExpirationSec)*time.Second); errSet != nil {
		log.Printf("设置redis缓存出错:%v", errSet)
	}
	return fee
}

func (s *ChainService) getFeeFromChain() *big.Int {
	log.Println("--从链上获取eth_gasPrice开始 ")
	priceOnChain, err := s.getBuildTxParams("eth_gasPrice", []interface{}{})
	// 当前从链上获取价格异常时，使用配置文件定义的值

	if err != nil {
		log.Printf("从链上获取gasPrice失败:%s", err.Error())
		return big.NewInt(conf.Config.ChainCfg.GasPrice)
	}

	if priceOnChain.Cmp(big.NewInt(0)) == -1 { // value <0
		return big.NewInt(conf.Config.ChainCfg.GasPrice)
	}
	log.Printf("从链上获取的gasPrice为 %s", priceOnChain.String())
	log.Println("--从链上获取eth_gasPrice结束 ")
	return priceOnChain
}

// gasPrice值调整
// `price`必须大于0，否则使配置文件的`gasPrice`最小值
// 如果`price`高于100GWei，在`price`基础上+30GWei
// 如果`price`等于或低于100GWei，支付1.3倍`price`价格
// 但都不可超出最低和最高伐值，一旦超出，就使用该伐值作为价格
func adjustFee(price *big.Int) *big.Int {
	maxLimit := maxGasPriceLimit()
	minLimit := minGasPriceLimit()

	if price.Cmp(big.NewInt(0)) != 1 { // price <= 0
		log.Printf("从链上获取到的gasPrice值<=0，使用配置文件的值:%s", minLimit.String())
		return minLimit
	}

	// 检查点值为 100GWei
	checkPoint := big.NewInt(100)
	checkPoint.Mul(checkPoint, GWei)

	if price.Cmp(checkPoint) == -1 { // price < 100GWei
		price.Div(price, big.NewInt(10)).Mul(price, big.NewInt(13))
		log.Printf("调整gasPrice价格 price=%s 增加百分之30", price.String())
	} else {
		thirty := big.NewInt(30)
		thirty.Mul(thirty, GWei)

		price.Add(price, thirty)
		log.Printf("调整gasPrice价格 price=%s 增加30GWei", price.String())
	}

	if price.Cmp(minLimit) == -1 { // price < minLimit
		price = minLimit
	}
	if price.Cmp(maxLimit) == 1 { // price > maxLimit
		price = maxLimit
	}
	log.Printf("调整gasPrice价格 最终计算得出的值为:%s", price.String())
	return price
}
