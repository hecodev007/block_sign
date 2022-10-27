package v1

import (
	"github.com/eth-sign/conf"
	"log"
	"math/big"
)

func (cs *FTMService) getFeeFromChain() *big.Int {
	priceOnChain, err := cs.getBuildTxParams("eth_gasPrice", []interface{}{})
	// 当前从链上获取价格异常时，使用配置文件定义的值

	if err != nil {
		log.Printf("从链上获取gasPrice失败: %s", err.Error())
		log.Printf("使用配置文件的gasPrice: %d", conf.Config.EthCfg.GasPrice)
		return big.NewInt(conf.Config.EthCfg.GasPrice)
	}

	if priceOnChain == nil || priceOnChain.Cmp(big.NewInt(0)) == -1 {
		log.Printf("使用配置文件的gasPrice: %d", conf.Config.EthCfg.GasPrice)
		return big.NewInt(conf.Config.EthCfg.GasPrice)
	}

	log.Printf("从链上获取的gasPrice为 %s", priceOnChain.String())
	return priceOnChain
}

/*
添加动态获取手续费的接口
*/
func (cs *FTMService) confirmFee() *big.Int {
	return adjustFee(cs.getFeeFromChain())
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
		log.Printf("从链上获取到的gasPrice值<=0，使用配置文件的值: %s", minLimit.String())
		return minLimit
	}

	// 检查点值为 2000GWei
	checkPoint := big.NewInt(2000)
	checkPoint.Mul(checkPoint, GWei)

	if price.Cmp(checkPoint) == -1 { // price < 2000GWei
		price.Div(price, big.NewInt(10)).Mul(price, big.NewInt(18))
		log.Printf("调整gasPrice价格 price= %s 增加百分之80", price.String())
	} else {
		thirty := big.NewInt(400)
		thirty.Mul(thirty, GWei)

		price.Add(price, thirty)
		log.Printf("调整gasPrice价格 price= %s 增加400GWei", price.String())
	}

	if price.Cmp(minLimit) == -1 { // price < minLimit
		price = minLimit
	}
	if price.Cmp(maxLimit) == 1 { // price > maxLimit
		price = maxLimit
	}
	log.Printf("调整gasPrice价格 最终计算得出的值为: %s", price.String())
	return price
}
