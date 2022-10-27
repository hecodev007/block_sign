package v1

import (
	"github.com/eth-sign/conf"
	"github.com/sirupsen/logrus"
	"math/big"
)

func (cs *CdsService) getFeeFromChain() *big.Int {
	priceOnChain, err := cs.getBuildTxParams("eth_gasPrice", []interface{}{})
	// 当前从链上获取价格异常时，使用配置文件定义的值

	if err != nil {
		logrus.Errorf("从链上获取gasPrice失败: %s", err.Error())
		logrus.Infof("使用配置文件的gasPrice: %d", conf.Config.EthCfg.GasPrice)
		return big.NewInt(conf.Config.EthCfg.GasPrice)
	}

	if priceOnChain == nil || priceOnChain.Cmp(big.NewInt(0)) == -1 {
		logrus.Infof("使用配置文件的gasPrice: %d", conf.Config.EthCfg.GasPrice)
		return big.NewInt(conf.Config.EthCfg.GasPrice)
	}

	logrus.Infof("从链上获取的gasPrice为 %s", priceOnChain.String())
	return priceOnChain
}

/*
添加动态获取手续费的接口
*/
func (cs *CdsService) confirmFee() *big.Int {
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
		logrus.Infof("从链上获取到的gasPrice值<=0，使用配置文件的值: %s", minLimit.String())
		return minLimit
	}

	// 检查点值为 100GWei
	checkPoint := big.NewInt(100)
	checkPoint.Mul(checkPoint, GWei)

	if price.Cmp(checkPoint) == -1 { // price < 100GWei
		price.Div(price, big.NewInt(10)).Mul(price, big.NewInt(13))
		logrus.Infof("调整gasPrice价格 price= %s 增加百分之30", price.String())
	} else {
		thirty := big.NewInt(30)
		thirty.Mul(thirty, GWei)

		price.Add(price, thirty)
		logrus.Infof("调整gasPrice价格 price= %s 增加30GWei", price.String())
	}

	if price.Cmp(minLimit) == -1 { // price < minLimit
		price = minLimit
	}
	if price.Cmp(maxLimit) == 1 { // price > maxLimit
		price = maxLimit
	}
	logrus.Infof("调整gasPrice价格 最终计算得出的值为: %s", price.String())
	return price
}
