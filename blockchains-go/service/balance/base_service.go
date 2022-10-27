package balance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"github.com/shopspring/decimal"
)

type BalanceBaseService struct {
}

func getMainCoin(coin string) *entity.FcCoinSet {
	coinSet, ok := global.CoinDecimal[coin]
	if !ok {
		log.Error("缺少币种信息 global.CoinDecimal")
		return nil
	}
	if coinSet.Pid == 0 {
		return nil
	}

	mainCoin, ok := global.CoinIdMap[coinSet.Pid]
	if !ok {
		log.Error("缺少币种信息 global.CoinIdMap")
		return nil
	}
	return mainCoin

}

func (b *BalanceBaseService) buildCoinBalance(totalAmount *dao.FcBalanceSta, outputAmount *dao.FcBalanceSta, liquidAmount *dao.FcBalanceSta, feeAmount *dao.FcBalanceSta) *model.CoinBalance {
	var err error
	coin := strings.ToLower(totalAmount.Coin)
	token := ""
	contract := ""
	total := decimal.Zero
	output := decimal.Zero
	liquid := decimal.Zero
	fee := decimal.Zero

	mainCoin := getMainCoin(coin)
	if mainCoin != nil {
		//表示是代币
		coin = mainCoin.Name
		token = totalAmount.Coin
		tokenSet, ok := global.CoinDecimal[strings.ToLower(token)]
		if ok {
			contract = tokenSet.Token
		}
	}

	if coin == "eth" {
		log.Infof("固定ETH金额总余额01 %s ", totalAmount.Amount)
	}

	total, err = decimal.NewFromString(totalAmount.Amount)
	if err != nil {
		log.Errorf("buildCoinBalance 总余额(%s)格式转换失败 %v", totalAmount.Amount, err)
	}

	if outputAmount != nil {
		output, err = decimal.NewFromString(outputAmount.Amount)
		if err != nil {
			log.Errorf("buildCoinBalance 出账地址余额(%s)格式转换失败 %v", outputAmount.Amount, err)
		}
	} else {
		log.Infof("buildCoinBalance 币种[%s %s]出账地址余额 数据为空", coin, token)
	}

	if liquidAmount != nil {
		liquid, err = decimal.NewFromString(liquidAmount.Amount)
		if err != nil {
			log.Errorf("buildCoinBalance 币种[%s %s] 可用余额(%s)格式转换失败 %v", coin, token, liquidAmount.Amount, err)
		}
	} else {
		log.Infof(" buildCoinBalance 币种[%s %s]可用余额 数据为空", coin, token)
	}

	if feeAmount != nil {
		fee, err = decimal.NewFromString(feeAmount.Amount)
		if err != nil {
			log.Errorf("buildCoinBalance 垫资地址余额(%s)格式转换失败 %v", feeAmount.Amount, err)
		}
	} else {
		log.Infof("buildCoinBalance 币种[%s %s]垫资地址余额 数据为空", coin, token)
	}
	if coin == "eth" {
		log.Infof("固定ETH金额总余额01 %s ", total.String())
	}

	if order.NewCollectEnable() && order.IsNewCollectVersion(coin) {
		v := decimal.Zero
		if !liquid.IsZero() {
			v = liquid
		}
		if output.Cmp(v) == -1 {
			output = v
		}
	}

	if liquid.Cmp(output) == -1 {
		liquid = output
	}

	m := &model.CoinBalance{
		CoinName:          coin,
		TokenName:         token,
		Balance:           total.String(),
		ActivityBalance:   output.String(),
		LiquidBalance:     liquid.String(),
		ContractAddress:   contract,
		FeeAddress:        fee.String(),
		TopsTwentyBalance: "0",
	}
	// 非账户模型，出账地址余额和总余额一致
	// 历史遗留问题，暂时按照旧的处理方式来
	coinSet, ok := global.CoinDecimal[strings.ToLower(coin)]
	if ok {
		//log.Infof("buildCoinBalance coinSet=%+v", coinSet)
		accountModel := 1
		if accountModel != coinSet.PatternType {
			m.ActivityBalance = liquid.String()
		}
	}
	return m
}

func (b *BalanceBaseService) fetchCoinBalance(coin string, balances []*dao.FcBalanceSta) *dao.FcBalanceSta {
	for _, balance := range balances {
		if strings.ToLower(coin) == strings.ToLower(balance.Coin) {
			return balance
		}
	}
	return nil
}

func (b *BalanceBaseService) GetMchAllBalanceV2(mchId int) ([]*model.CoinBalance, error) {
	log.Infof("GetMchAllBalanceV2 开始执行")
	balanceTotals, err := dao.FcAddressAmountBalanceTotal(mchId)
	if err != nil {
		log.Errorf("从数据库获取总余额统计数据失败 %v", err)
		return nil, errors.New("failed to get total balance data")
	}
	log.Infof("从数据库获取总余额统计数据 完成")
	balanceFees, err := dao.FcAddressAmountBalanceFee(mchId)
	if err != nil {
		log.Errorf("从数据库获取垫资地址统计数据失败 %v", err)
		return nil, errors.New("failed to get fee address balance data")
	}
	balanceOuts, err := dao.FcAddressAmountBalanceOut(mchId)
	if err != nil {
		log.Errorf("从数据库获取出账地址余额统计数据失败 %v", err)
		return nil, errors.New("failed to get out balance data")
	}
	log.Infof("从数据库获取出账地址余额统计数据 完成")

	balanceLiquids, err := dao.FcAddressAmountBalanceLiquid(mchId)
	if err != nil {
		log.Errorf("从数据库获取可用余额统计数据失败 %v", err)
		return nil, errors.New("failed to get liquid balance data")
	}
	log.Infof("从数据库获取可用余额统计数据 完成")

	result := make([]*model.CoinBalance, 0, len(balanceTotals))
	for _, sta := range balanceTotals {
		if sta.Coin == "eth" {
			log.Infof("固定ETH金额 %s", sta.Amount)
		}
		balance := b.buildCoinBalance(sta, b.fetchCoinBalance(sta.Coin, balanceOuts), b.fetchCoinBalance(sta.Coin, balanceLiquids), b.fetchCoinBalance(sta.Coin, balanceFees))
		result = append(result, balance)
	}
	log.Infof("GetMchAllBalanceV2 执行完毕")
	return result, nil

}

func (b *BalanceBaseService) GetMchAllBalance(mchId int) ([]*model.CoinBalance, error) {

	results := make([]*model.CoinBalance, 0)

	//查询商户购买的币种
	coins, err := dao.FcMchServiceFindsValid(mchId)
	if err != nil {
		log.Errorf("商户：%d，没有币种购买记录,error:%s", mchId, err.Error())
	}
	if len(coins) == 0 {
		log.Errorf("商户：%d，没有币种购买记录", mchId)
		return nil, fmt.Errorf("商户：%d，没有币种购买记录", mchId)
	}
	for _, v := range coins {
		coinResult, err := dao.FcCoinSetGetCoinInfo(v.CoinId)
		if err != nil {
			//跳过这个币种
			log.Errorf("商户：%d,查询币种异常：%s", mchId, err.Error())
			continue
		}
		data := &model.CoinBalance{}
		balance := decimal.Zero
		data.CoinName = coinResult.Name

		//if coinResult.Pid == 0 {
		//	data.CoinName = coinResult.Name
		//} else {
		//	//查询父类币种
		//	coinResultFa, err := dao.FcCoinSetGetCoinInfo(coinResult.Pid)
		//	if err != nil {
		//		//跳过这个币种
		//		log.Errorf("商户：%d,查询父类币种异常：%s，查询币种：%s,父类ID：%d", mchId, err.Error(), v.CoinName, coinResult.Pid)
		//		continue
		//	}
		//	data.CoinName = coinResultFa.Name
		//	data.TokenName = coinResult.Name
		//	data.ContractAddress = coinResult.Token
		//}
		result, err := dao.FcMchAmountGetByACId(mchId, coinResult.Id)
		if err != nil {
			if err.Error() != "Not Fount!" {
				log.Errorf("商户：%d,查询币种余额异常：%s", mchId, err.Error())
				continue
			}
		}
		if result != nil {
			balance, _ = decimal.NewFromString(result.Amount)
		}
		data.Balance = balance.String()
		if coinResult.PatternType == 1 {
			activityBalance, err := dao.FcAddressAmountGetTotalAmountWithType(mchId, coinResult.Name, 1)
			if err != nil {
				if err.Error() != "Not Fount!" {
					log.Errorf("商户：%d,查询币种%s余额异常：%s", mchId, coinResult.Name, err.Error())
					continue
				}
			}
			data.ActivityBalance = activityBalance.String()
		} else {
			data.ActivityBalance = balance.String()
		}

		topsTwentyBalance, err := dao.FcAddressAmountGetTotalAmountWithLimit(mchId, coinResult.Name, 20)
		if err != nil {
			if err.Error() != "Not Fount!" {
				log.Errorf("商户：%d,查询币种%s前20地址余额异常：%s", mchId, coinResult.Name, err.Error())
				continue
			}
		}
		data.TopsTwentyBalance = topsTwentyBalance.String()
		results = append(results, data)

		//尝试查询这个币种下面的代币
		tokenCoinSets, err := dao.FcCoinSetFindByPidStatus(v.CoinId, 1)
		if err != nil {
			log.Errorf("FcCoinSetFindByPidStatus error :%s", err.Error())
		}
		for _, vt := range tokenCoinSets {
			data = &model.CoinBalance{}
			balance = decimal.Zero
			result, err := dao.FcMchAmountGetByACId(mchId, vt.Id)
			if err != nil {
				if err.Error() != "Not Fount!" {
					log.Errorf("商户：%d,查询币种余额异常：%s", mchId, err.Error())
					continue
				}
			}
			if result != nil {
				balance, _ = decimal.NewFromString(result.Amount)
			}
			data.Balance = balance.String()
			activityBalance, err := dao.FcAddressAmountGetTotalAmountWithType(mchId, vt.Name, 1)
			if err != nil {
				if err.Error() != "Not Fount!" {
					log.Errorf("商户：%d,查询币种%s余额异常：%s", mchId, vt.Name, err.Error())
					continue
				}
			}
			topsTwentyBalance, err := dao.FcAddressAmountGetTotalAmountWithLimit(mchId, vt.Name, 20)
			if err != nil {
				if err.Error() != "Not Fount!" {
					log.Errorf("商户：%d,查询币种%s前20地址余额异常：%s", mchId, vt.Name, err.Error())
					continue
				}
			}

			data.CoinName = coinResult.Name
			data.TokenName = vt.Name
			data.ContractAddress = vt.Token
			data.ActivityBalance = activityBalance.String()
			data.TopsTwentyBalance = topsTwentyBalance.String()
			results = append(results, data)
		}
	}
	return results, nil

}

func (b *BalanceBaseService) GetMchCoinMaxBalance(coinName, tokenName string, mchId int) (decimal.Decimal, string, error) {
	coinType := coinName
	if tokenName != "" {
		log.Infof("GetMchCoinMaxBalance 查询代币 %s 单地址最大余额", tokenName)
		coinType = tokenName
	} else {
		log.Infof("GetMchCoinMaxBalance 查询主链币 %s 单地址最大余额", tokenName)
	}

	coinSet, err := dao.FcCoinSetGetByName(coinType, 1)
	if err != nil {
		log.Errorf("GetMchCoinMaxBalance FcCoinSetGetByName 出错 %s", err.Error())
		return decimal.Zero, "", errors.New("FcCoinSetGetByName:failed to find coin set")
	}

	utxoMode := 2
	if utxoMode == coinSet.PatternType {
		log.Infof("GetMchCoinMaxBalance UTXO模型币种 patternType=%d", coinSet.PatternType)
		result, err := dao.FcMchAmountGetByACId(mchId, coinSet.Id)
		if err != nil {
			log.Errorf("GetMchCoinMaxBalance FcMchAmountGetByACId 出错 %s", err.Error())
			return decimal.Zero, "", errors.New("FcMchAmountGetByACId:failed to find mch amount")
		}
		balance, err := decimal.NewFromString(result.Amount)
		if err != nil {
			log.Errorf("GetMchCoinMaxBalance string to decimal(%s) 出错 %s", result.Amount, err.Error())
			return decimal.Zero, "", errors.New("invalid balance amount")
		}
		return balance, "", nil
	} else {
		log.Infof("GetMchCoinMaxBalance 非UTXO模型币种 patternType=%d", coinSet.PatternType)
		addrs, err := dao.FcGenerateAddressListFindColdAddrs(mchId, coinName)
		if err != nil {
			log.Errorf("GetMchCoinMaxBalance FcGenerateAddressListFindColdAddrs 出错 %s", err.Error())
			return decimal.Zero, "", errors.New("GetMchCoinMaxBalance:failed to find address")
		}

		addrAmount, err := dao.FcAddressAmountGetByCoinAndAddrs(coinType, addrs)
		if err != nil {
			log.Errorf("GetMchCoinMaxBalance FcAddressAmountGetByCoinAndAddrs 出错 %s", err.Error())
			return decimal.Zero, "", errors.New("GetMchCoinMaxBalance:failed to find address amount")
		}
		balance := addrAmount.Amount
		log.Infof("GetMchCoinMaxBalance chain=%s", coinName)
		if order.NewCollectEnable() && order.IsNewCollectVersion(coinName) {
			log.Infof("GetMchCoinMaxBalance 检查是否新版本归集")
			liquidByCoin, err := dao.FcAddressAmountBalanceLiquidByCoin(int64(mchId), coinType)
			log.Infof("GetMchCoinMaxBalance liquidByCoin=%v", liquidByCoin)
			if err != nil {
				log.Infof("GetMchCoinMaxBalance使用新版本出账归集 FcAddressAmountBalanceLiquidByCoin err: %v", err)
			} else {
				if liquidByCoin != nil {
					log.Infof("GetMchCoinMaxBalance使用新版本出账归集 %s %s", coinType, liquidByCoin.Amount)
					balance = liquidByCoin.Amount
				}
			}
		}

		balanceDecimal, err := decimal.NewFromString(balance)
		if err != nil {
			log.Errorf("GetMchCoinMaxBalance string to decimal 出错 %s", err.Error())
			return decimal.Zero, "", errors.New("invalid amount")
		}

		return balanceDecimal, addrAmount.Address, nil
	}

}

func (b *BalanceBaseService) GetMchBalance(coinName string, mchId int) (decimal.Decimal, error) {
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, "")
	if err != nil {
		return decimal.Zero, err
	}
	result, err := dao.FcMchAmountGetByACId(mchId, coinResult.Id)
	if err != nil {
		return decimal.Zero, err
	}
	balance, err := decimal.NewFromString(result.Amount)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (b *BalanceBaseService) GetMchTokenBalance(coinName string, tokenName string, mchId int) (decimal.Decimal, error) {
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, tokenName)
	if err != nil {
		return decimal.Zero, err
	}
	result, err := dao.FcMchAmountGetByACId(mchId, coinResult.Id)
	if err != nil {
		return decimal.Zero, err
	}
	balance, err := decimal.NewFromString(result.Amount)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (b *BalanceBaseService) GetMchActivityBalance(coinName string, contractAddress string, mchId int) (decimal.Decimal, error) {
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		return decimal.Zero, err
	}
	name := coinResult.Name
	balance, err := dao.FcAddressAmountGetTotalAmountWithType(mchId, name, 1)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (b *BalanceBaseService) GetTopsTwentyAddresses(coinName string, contractAddress string, mchId int) (decimal.Decimal, error) {
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		return decimal.Zero, err
	}
	name := coinResult.Name
	balance, err := dao.FcAddressAmountGetTotalAmountWithLimit(mchId, name, 20)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func NewBalanceBaseService() service.BalanceService {
	return &BalanceBaseService{}
}
