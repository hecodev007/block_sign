package security

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/transfer"
	"github.com/shopspring/decimal"
)

type SecurityService struct {
	transferService map[string]service.TransferService
}

func NewSecurityService() service.TransferSecurityService {
	coinSrv := make(map[string]service.TransferService, 0)
	coinSrv["btc"] = transfer.NewBtcTransferService()
	coinSrv["cocos"] = transfer.NewCocosTransferService()
	coinSrv["mdu"] = transfer.NewMduTransferService()
	coinSrv["zvc"] = transfer.NewZvcTransferService()
	coinSrv["fo"] = transfer.NewFoTransferService()
	coinSrv["klay"] = transfer.NewKlayTransferService()
	coinSrv["waxp"] = transfer.NewWaxpTransferService()
	coinSrv["eth"] = transfer.NewEthTransferService()
	coinSrv["gxc"] = transfer.NewGxcTransferService()
	coinSrv["eos"] = transfer.NewEosTransferService()
	coinSrv["etc"] = transfer.NewEtcTransferService()
	coinSrv["seek"] = transfer.NewSeekTransferService()
	coinSrv["usdt"] = transfer.NewUsdtTransferService()
	coinSrv["kava"] = transfer.NewKavaTransferService()
	coinSrv["luna"] = transfer.NewLunaTransferService()
	coinSrv["lunc"] = transfer.NewLuncTransferService()
	// write by jun 2020/4/29
	coinSrv["bnb"] = transfer.NewBnbTransferService()
	coinSrv["xlm"] = transfer.NewXlmTransferService()
	coinSrv["rub"] = transfer.NewRubTransferService()
	coinSrv["hx"] = transfer.NewHxTransferService()
	coinSrv["cds"] = transfer.NewCdsTransferService()
	coinSrv["ont"] = transfer.NewOntTransferService()
	coinSrv["ar"] = transfer.NewARTransferService()
	coinSrv["ksm"] = transfer.NewKsmTransferService()
	coinSrv["bnc"] = transfer.NewBncTransferService()
	coinSrv["crust"] = transfer.NewCRustTransferService()
	coinSrv["crab"] = transfer.NewCringTransferService()
	coinSrv["hnt"] = transfer.NewHntTransferService()
	coinSrv["vet"] = transfer.NewVetTransferService()
	coinSrv["bsv"] = transfer.NewBsvTransferService()
	coinSrv["ltc"] = transfer.NewLtcTransferService()
	coinSrv["uca"] = transfer.NewUcaTransferService()
	coinSrv["celo"] = transfer.NewCeloTransferService()
	coinSrv["mtr"] = transfer.NewMtrTransferService()
	coinSrv["fio"] = transfer.NewFioTransferService()
	coinSrv["qtum"] = transfer.NewQtumTransferService()
	coinSrv["sol"] = transfer.NewSolTransferService()
	coinSrv["tlos"] = transfer.NewTlosTransferService()
	// coinSrv["pcx"] = transfer.NewPcxTransferService()
	coinSrv["ghost"] = transfer.NewGhostTransferService()
	coinSrv["dot"] = transfer.NewDotTransferService()
	coinSrv["azero"] = transfer.NewAzeroTransferService()
	coinSrv["sgb-sgb"] = transfer.NewSgbTransferService()
	coinSrv["kar"] = transfer.NewKarTransferService()
	coinSrv["nodle"] = transfer.NewNodleTransferService()
	coinSrv["bch"] = transfer.NewBchTransferService()
	coinSrv["zec"] = transfer.NewZecTransferService()
	coinSrv["dcr"] = transfer.NewDcrTransferService()
	coinSrv["btm"] = transfer.NewBtmTransferService()
	coinSrv["ckb"] = transfer.NewCkbTransferService()
	coinSrv["hc"] = transfer.NewHcTransferService()
	// coinSrv["stx"] = transfer.NewStxTransferService()
	coinSrv["stx"] = transfer.NewStxNewTransferService()
	coinSrv["nas"] = transfer.NewNasTransferService()
	coinSrv["doge"] = transfer.NewDogeTransferService()
	coinSrv["avax"] = transfer.NewAvaxTransferService()
	coinSrv["bsc"] = transfer.NewBscTransferService()
	coinSrv["fil"] = transfer.NewFilTransferService()
	coinSrv["wd"] = transfer.NewWd_wdTransferService()
	coinSrv["dash"] = transfer.NewDashTransferService()
	coinSrv["biw"] = transfer.NewBiwTransferService()
	coinSrv["atom"] = transfer.NewAtomTransferService()
	coinSrv["near"] = transfer.NewNearTransferService()
	coinSrv["yta"] = transfer.NewYtaTransferService()
	coinSrv["cfx"] = transfer.NewCfxTransferService()
	coinSrv["star"] = transfer.NewStarTransferService()
	coinSrv["fis"] = transfer.NewFisTransferService()
	coinSrv["oneo"] = transfer.NewNeoTransferService()
	coinSrv["atp"] = transfer.NewAtpTransferService()
	coinSrv["cph-cph"] = transfer.NewCphTransferService()
	coinSrv["pcx"] = transfer.NewChainXTransferService() // chainX2.0
	// coinSrv["pcx"] = transfer.NewPcxTransferService()
	coinSrv["bcha"] = transfer.NewBchaTransferService()
	coinSrv["xec"] = transfer.NewXecTransferService()
	coinSrv["ada"] = transfer.NewAdaTransferService()
	coinSrv["trx"] = transfer.NewTrxTransferService()
	coinSrv["zen"] = transfer.NewZenTransferService()
	coinSrv["mw"] = transfer.NewMwTransferService()
	coinSrv["dip"] = transfer.NewDipTransferService()
	coinSrv["algo"] = transfer.NewAlgoTransferService()
	coinSrv["ori"] = transfer.NewOriTransferService()
	coinSrv["bos"] = transfer.NewBosTransferService()
	coinSrv["okt"] = transfer.NewOktTransferService()
	coinSrv["waves"] = transfer.NewWavesTransferService()
	coinSrv["glmr"] = transfer.NewGlmrTransferService()
	coinSrv["avaxcchain"] = transfer.NewAvaxcchainTransferService()
	coinSrv["heco"] = transfer.NewHecoTransferService()
	coinSrv["nyzo"] = transfer.NewNyzoTransferService()
	coinSrv["xdag"] = transfer.NewXdagTransferService()
	coinSrv["iost"] = transfer.NewIostTransferService()
	coinSrv["hsc"] = transfer.NewHscTransferService()
	coinSrv["dhx"] = transfer.NewDhxTransferService()
	coinSrv["dom"] = transfer.NewDomTransferService()
	coinSrv["wtc"] = transfer.NewWtcTransferService()
	coinSrv["moac"] = transfer.NewMoacTransferService()
	coinSrv["satcoin"] = transfer.NewSatcoinTransferService()
	coinSrv["eac"] = transfer.NewEacTransferService()
	coinSrv["iota"] = transfer.NewIotaTransferService()
	coinSrv["kai"] = transfer.NewKaiTransferService()
	coinSrv["rbtc"] = transfer.NewRbtcTransferService()
	coinSrv["movr"] = transfer.NewMovrTransferService()
	coinSrv["sep20"] = transfer.NewSep20TransferService()
	coinSrv["rei"] = transfer.NewReiTransferService()
	coinSrv["dscc"] = transfer.NewDsccTransferService()
	coinSrv["dscc1"] = transfer.NewDscc1TransferService()
	coinSrv["brise-brise"] = transfer.NewBriseTransferService()
	coinSrv["ccn"] = transfer.NewCcnTransferService()
	coinSrv["optim"] = transfer.NewOptimTransferService()
	coinSrv["ftm"] = transfer.NewFtmTransferService()
	coinSrv["welups"] = transfer.NewWelTransferService()
	coinSrv["rose"] = transfer.NewRoseTransferService()
	coinSrv["one"] = transfer.NewOneTransferService()
	coinSrv["rev"] = transfer.NewRevTransferService()
	coinSrv["ron"] = transfer.NewRonTransferService()
	coinSrv["tkm"] = transfer.NewTkmTransferService()
	coinSrv["neo"] = transfer.NewN3neoTransferService()
	coinSrv["flow"] = transfer.NewFlowTransferService()
	coinSrv["icp"] = transfer.NewIcpTransferService()
	coinSrv["uenc"] = transfer.NewUencTransferService()
	coinSrv["cspr"] = transfer.NewCsprTransferService()
	coinSrv["matic-matic"] = transfer.NewMaticTransferService()
	coinSrv["iotx"] = transfer.NewIotexTransferService()
	coinSrv["evmos"] = transfer.NewEvmosTransferService()
	coinSrv["aur"] = transfer.NewAurTransferService()
	coinSrv["mob"] = transfer.NewMobTransferService()
	coinSrv["deso"] = transfer.NewDesoTransferService()
	coinSrv["lat"] = transfer.NewLatTransferService()
	coinSrv["hbar"] = transfer.NewHbarTransferService()
	coinSrv["steem"] = transfer.NewSteemTransferService()

	return &SecurityService{
		transferService: coinSrv,
	}
}

func (s *SecurityService) IsRunningOrder(outOrderNo, coinName string, mchId int) (bool, error) {
	//if _, ok := global.WalletType(coinName); !ok {
	//	return true, fmt.Errorf("缺少该币种的钱包类型设置，币种：%s", coinName)
	//}
	waType := global.WalletType(coinName, mchId)
	switch waType {
	case status.WalletType_Cold:
		return s.checkColdOrders(outOrderNo)
	case status.WalletType_Hot:
		return s.checkHotOrders(outOrderNo)
	default:
		return true, errors.New("未知的钱包配置类型")
	}
}

func (s *SecurityService) checkHotOrders(outOrderNo string) (bool, error) {
	list, err := dao.FcOrderHotFindListByOutNo(outOrderNo)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		return false, nil
	}
	for _, orderHot := range list {
		if orderHot.Status <= 4 {
			return true, fmt.Errorf("订单：%s 已经广播 或 正在等待处理 ", outOrderNo)
		}
	}
	return false, nil
}

func (s *SecurityService) checkColdOrders(outOrderNo string) (bool, error) {
	list, err := dao.FcOrderFindListByOutNo(outOrderNo)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		return false, nil
	}
	for _, orderHot := range list {
		if orderHot.Status <= 4 {
			return true, fmt.Errorf("订单：%s 已经广播 或 正在等待处理 ", outOrderNo)
		}
	}
	return false, nil
}

func (s *SecurityService) IsAssignAddress(coinName string, mchId int) (bool, error) {
	has, err := dao.FcGenerateAddressListIsAssign(coinName, mchId)
	return has, err
}

func (s *SecurityService) IsDuplicateApplyOrder(outOrderNo string, mchName string) (bool, error) {
	ta, err := dao.FcTransfersApplyByOutOrderNo(outOrderNo)
	if err != nil {
		if err.Error() == "Not Fount!" {
			// 允许交易
			return false, nil
		} else {
			// 不允许交易
			return true, fmt.Errorf("异常订单号：%s", outOrderNo)
		}
	}
	if ta == nil {
		// 允许交易
		return false, nil
	} else {
		// 不允许交易
		return true, fmt.Errorf("已存在相同订单：%s,db id:%d", outOrderNo, ta.Id)
	}

	// if err != nil {
	//	if err.Error() == "Not Fount!" {
	//		//没有相关记录，可以交易
	//		return false, nil
	//	}
	//	log.Errorf("IsDuplicateOrder 查询异常：%s", err.Error())
	//	return true, err
	// }
	// if ta == nil {
	//	//没有相关记录，可以交易
	//	return false, nil
	// }

	// if _, ok := global.WalletType[ta.CoinName]; !ok {
	//	return true, fmt.Errorf("缺少该币种的钱包类型设置，币种：%s", ta.CoinName)
	// }
	// waType := global.WalletType[ta.CoinName]
	// //0:构建完成,1:推入队列,2:已拉取,3:已签名,4:已广播,5:构建失败,6:签名失败7:广播失败 8:超时
	// switch waType {
	// case status.WalletType_Cold:
	//	//查询是否已经广播
	//	if ok := dao.FcOrderHaveByOutOrderNo(outOrderNo, 4); ok {
	//		//已经广播
	//		return true, fmt.Errorf("商户:%s，订单：%s 已经广播 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHaveByOutOrderNo(outOrderNo, 1); ok {
	//		//正在等待交易
	//		return true, fmt.Errorf("商户:%s，订单：%s 已经推入队列 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHaveByOutOrderNo(outOrderNo, 2); ok {
	//		//交易已经拉取
	//		return true, fmt.Errorf("商户:%s，订单：%s 正在被拉取 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHaveByOutOrderNo(outOrderNo, 3); ok {
	//		return true, fmt.Errorf("商户:%s，订单：%s 已签名，等待广播中..", mchName, outOrderNo)
	//	}
	//	return false, nil
	// case status.WalletType_Hot:
	//	//查询是否已经广播
	//	if ok := dao.FcOrderHotHaveByOutOrderNo(outOrderNo, 4); ok {
	//		//已经广播
	//		return true, fmt.Errorf("商户:%s，订单：%s 已经广播 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHotHaveByOutOrderNo(outOrderNo, 1); ok {
	//		//正在等待交易
	//		return true, fmt.Errorf("商户:%s，订单：%s 已经推入队列 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHotHaveByOutOrderNo(outOrderNo, 2); ok {
	//		//交易已经拉取
	//		return true, fmt.Errorf("商户:%s，订单：%s 正在被拉取 ", mchName, outOrderNo)
	//	}
	//	if ok := dao.FcOrderHotHaveByOutOrderNo(outOrderNo, 3); ok {
	//		return true, fmt.Errorf("商户:%s，订单：%s 已签名，等待广播中..", mchName, outOrderNo)
	//	}
	//	return false, nil
	// default:
	//	return true, errors.New("未知的钱包配置类型")
	// }

}

// 风险验证
// 单笔出账限额，每小时出账限额，每日出账限额
func (s *SecurityService) VerifyRisk(coinName string, amount decimal.Decimal, mchId int) (bool, error) {
	_, ok := global.CoinDecimal[coinName]
	if !ok {
		return false, fmt.Errorf("Miss Coin:%s,", coinName)
	}
	v := global.CoinDecimal[coinName]

	// 最大金额
	dbHugeNum, err := decimal.NewFromString(v.HugeNum)
	if err != nil {
		return false, fmt.Errorf("read error,coinName:%s", coinName)
	}
	// 最小金额
	dbMinNum, err := decimal.NewFromString(v.Num)
	if err != nil {
		return false, fmt.Errorf("read error,coinName:%s", coinName)
	}
	if dbHugeNum.Equals(decimal.Zero) && dbMinNum.Equals(decimal.Zero) {
		// 不做限制
		return true, nil
	}
	if amount.GreaterThanOrEqual(dbHugeNum) || amount.LessThan(dbMinNum) {
		return false, fmt.Errorf(" amount error max:%s,min:%s", dbHugeNum.String(), dbMinNum.String())
	}
	return true, nil
}

// 验证币种是否支持
func (s *SecurityService) VerifyCoin(coinName string, mchId int) (bool, error) {
	if _, ok := global.CoinDecimal[coinName]; ok {
		return true, nil
	}
	return false, fmt.Errorf("Miss Coin:%s,", coinName)
}

// 验证币种关闭还是开放
func (s *SecurityService) VerifyCoinPermission(coinName string, mchId int) (bool, error) {
	_, ok := global.CoinDecimal[coinName]
	if !ok {
		return false, fmt.Errorf("Miss Coin:%s,", coinName)
	}
	v := global.CoinDecimal[coinName]
	if v.WStatus == 0 {
		// 提现关闭
		return false, fmt.Errorf("coinName:%s,closed", coinName)
	}
	return true, nil

}

// 验证币种精度
func (s *SecurityService) VerifyCoinDecimal(coinName string, amount decimal.Decimal) (bool, error) {
	bit := amount.Exponent() * -1
	_, ok := global.CoinDecimal[coinName]
	if !ok {
		return false, fmt.Errorf("Miss Coin:%s,", coinName)
	}
	v := global.CoinDecimal[coinName]
	log.Infof("全局存储精度：%d,传入精度:%d", v.Decimal, bit)
	if v.Decimal < int(bit) {
		// 超过精度
		return false, fmt.Errorf("coinName:%s,amount:%s,decimal error,system:%d", coinName, amount, v.Decimal)
	}
	return true, nil
}

// 验证api是否允许访问
func (s *SecurityService) VerifyApiPermission(path, coinName, ip string, mchId int) (bool, error) {
	log.Infof("传入参数 path：%s  coinName：%s  ip：%s  mchId：%d", path, coinName, ip, mchId)
	// for k, _ := range global.MchAuth {
	//	log.Infof("存在的id：%d", k)
	// }

	// 先查询商户是否存在
	_, ok := global.MchAuth[mchId]
	if !ok {
		return false, fmt.Errorf("ID：%d,不存在", mchId)
	}

	// 打印开始
	// vMap := global.MchAuth[mchId]
	// log.Infof("商户：%d 允许访问的路径", mchId)
	// for k, v := range vMap.Api {
	//	log.Infof("%s------------%+v", k, v)
	// }
	// 打印结束

	nowTime := time.Now().Unix()
	mchAuth := global.MchAuth[mchId]
	if mchAuth.StartTime > nowTime || mchAuth.EndTime < nowTime {
		return false, fmt.Errorf(
			"ID：%d,币种使用不在合法期限内，start:%d,end:%d,now:%d",
			mchId,
			mchAuth.StartTime,
			mchAuth.EndTime,
			nowTime,
		)
	}

	// 查询商户是否购买币种
	_, ok = mchAuth.Api[coinName]
	if !ok {
		return false, fmt.Errorf("ID：%d,缺少购买币种 %s", mchId, coinName)
	}

	coins := mchAuth.Api[coinName]
	_, ok = coins.Auth[path]
	if !ok {
		return false, fmt.Errorf("ID：%d,缺少购买币种 %s,路径：%s 的访问设置", mchId, coinName, path)
	}
	pathAuths := coins.Auth[path]
	if len(pathAuths.IP) == 0 {
		return false, fmt.Errorf("ID：%d,缺少购买币种 %s,路径：%s 的访问的IP设置", mchId, coinName, path)
	}

	// 判断IP
	for _, v := range pathAuths.IP {
		if v == ip {
			return true, nil
		}
	}
	return false, fmt.Errorf("ID：%d，path：%s,ip:%s,验证错误,", mchId, path, ip)

}

// 验证商户余额
// param ok：true 满足出账  false 不满足出账
// param mchBalance： 商户数据库余额
// param err：错误描述
func (s *SecurityService) VerifyMchBalance(coinName, contractAddress string, transferAmount decimal.Decimal, mchId int) (ok bool, mchBalance decimal.Decimal, err error) {
	coinResult, err := dao.FcCoinSetGetCoinId(coinName, contractAddress)
	if err != nil {
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,查询币种异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	result, err := dao.FcMchAmountGetByACId(mchId, coinResult.Id)
	if err != nil {
		if err.Error() == "Not Fount!" {
			return false, decimal.Zero, fmt.Errorf("商户余额不足，余额0")
		}
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,查询余额异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	balance, err := decimal.NewFromString(result.Amount)
	if err != nil {
		log.Errorf("mchId:%d,coin:%s,contractAddress:%s,转换余额异常：%s", mchId, coinName, contractAddress, err.Error())
		return false, decimal.Zero, err
	}
	// todo 计算预扣手续费
	if strings.ToLower(coinName) == "usdt" {
		if transferAmount.Equals(decimal.Zero) || transferAmount.GreaterThan(balance) {
			return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s (尚未扣除手续费)", balance.String(), transferAmount.String())
		}
	} else if strings.ToLower(coinName) == "ar" {
		// 总余额需要扣除0.26 因为第一笔手续费需要大于0.25
		if transferAmount.Equals(decimal.Zero) || transferAmount.GreaterThan(balance.Sub(decimal.NewFromFloat(0.26))) {
			return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s ，商户余额需要预扣0.26手续费", balance.String(), transferAmount.String())
		}

	} else {
		if contractAddress == "" {
			if transferAmount.GreaterThanOrEqual(balance) {
				// 主链币
				return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s (尚未扣除手续费)", balance.String(), transferAmount.String())
			}
		} else {
			if transferAmount.GreaterThan(balance) {
				return false, balance, fmt.Errorf("商户余额：%s，需要出账余额：%s", balance.String(), transferAmount.String())
			}
		}
	}

	return true, balance, nil

}

// //验证合法地址
// func (s *SecurityService) VerifyAddress(address, coinName string) (bool, error) {
//	//各个币种的验证没有统一只能先分开写
//	var (
//		data          []byte
//		url           string
//		addressResult bool
//		err           error
//	)
//	addressResult = false
//	if _, ok := global.CheckAddressServer[coinName]; !ok {
//		return false, fmt.Errorf("不存在该币种的验证服务api，币种：%s", coinName)
//	}
//	url = global.CheckAddressServer[coinName]
//	switch coinName {
//	case "btc", "usdt", "zec", "ltc":
//		url = fmt.Sprintf(url, address)
//		data, err = util.Get(url)
//		if err != nil {
//			addressResult = false
//			err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", coinName, address, err.Error())
//			break
//		}
//		log.Infof("验证地址返回结果：%s", string(data))
//		btcResp := decodeBtcAddressResult(data)
//		if btcResp != nil && btcResp.Data != nil {
//			if btcResp.Data.Isvalid {
//				addressResult = true
//				err = nil
//				break
//			}
//		}
//		addressResult = false
//		err = fmt.Errorf("验证地址错误，%s,address:%s", coinName, address)
//		break
//	case "cocos":
//		addressResult = true
//		err = nil
//		url = fmt.Sprintf(url, address)
//		data, err = util.GetByAuth(url, "cocos", "cocospwd123.")
//		if err != nil {
//			addressResult = false
//			err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", coinName, address, err.Error())
//			break
//		}
//		log.Infof("验证地址返回结果：%s", string(data))
//		btcResp := decodeCocosAccountResult(data)
//		if btcResp != nil && btcResp.Message != "" {
//			if btcResp.Code == 0 {
//				addressResult = true
//				err = nil
//				break
//			}
//		}
//		addressResult = false
//		err = fmt.Errorf("验证地址错误，%s,address:%s", coinName, address)
//		break
//	case "zvc":
//		mapData := make(map[string]string, 0)
//		mapData["coinname"] = "zvc"
//		mapData["address"] = address
//		zvcdata, _ := json.Marshal(mapData)
//		data, err = util.PostJsonData(url, zvcdata)
//		if err != nil {
//			addressResult = false
//			err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", coinName, address, err.Error())
//			break
//		}
//		log.Infof("验证地址返回结果：%s", string(data))
//		zvcResp := decodeZvcAddrResult(data)
//		if zvcResp != nil && zvcResp.Code == 0 {
//			addressResult = true
//			err = nil
//			break
//		}
//		addressResult = false
//		err = fmt.Errorf("验证地址错误，%s,address:%s", coinName, address)
//		break
//	default:
//		addressResult = false
//		err = fmt.Errorf("不支持币种%s", coinName)
//		break
//
//	}
//	return addressResult, err
// }

// 验证合法地址
func (s *SecurityService) VerifyAddress(address, coinName string) (bool, error) {
	// if _, ok := global.CheckAddressServer[coinName]; !ok {
	//	return false, fmt.Errorf("不支持币种，不存在该币种的验证服务api，币种：%s", coinName)
	// }

	var err error

	if s.needCheck(coinName) {
		coinSet, err := dao.FcCoinSetGetByName(coinName, 1)
		if err != nil {
			log.Errorf("coi %s not found in database", coinName)
			return false, err
		}

		// 目前只做账户模型的校验
		accountMode := 1
		if accountMode == coinSet.PatternType {
			if err := s.checkColdAndContractAddress(address, coinName, coinSet); err != nil {
				return false, err
			}
		}
	}

	log.Infof("地址 %s 币种 %s checkColdAndContractAddress 通过", address, coinName)

	if _, ok := s.transferService[coinName]; !ok {
		return false, fmt.Errorf("不支持币种，没有初始化该币种的验证服务api，币种：%s", coinName)
	}
	err = s.transferService[coinName].VaildAddr(address)
	if err != nil {
		log.Errorf("%s", err.Error())
		return false, err
	}
	return true, nil
}

func (s *SecurityService) checkColdAndContractAddress(address string, coinName string, coinSet *entity.FcCoinSet) error {
	// 1、限制每条链：不能提现到本链出账地址里。
	// 比如用户提ETH链，那么他提任何ETH链的币，地址都不能够输入我们ETH链的出账地址。
	// 一旦输入，就在前端显示地址错误
	count, err := dao.FcGenerateAddressColdCount(coinName, address)
	if err != nil {
		log.Errorf("VerifyAddress FcGenerateAddressColdCount err %s", err.Error())
	}

	if count > 0 {
		return fmt.Errorf("此地址为出账地址，暂不支持 %s %s", coinName, address)
	}

	// 2、限制每条链：不能提现到本链的合约地址里
	// 比如用户提BSC链，那么他提任何BSC链的币，地址都不能够输入我们BSC链的合约地址。
	// 一旦输入，就在前端显示地址错误
	contractSet, err := dao.FcCoinSetGetCoinByContractAndPid(address, coinSet.Id)
	if err != nil {
		log.Errorf("VerifyAddress FcCoinSetGetCoinByContract err %s", err.Error())
	}

	if contractSet != nil {
		if err != nil {
			log.Errorf("VerifyAddress FcCoinSetGetByName err %s", err.Error())
		}
		if contractSet.Pid == coinSet.Id {
			log.Errorf("此地址为合约地址，暂不支持 %s %s", coinName, address)
			return errors.New("contractAddrNotAllow")
		}
	}
	return nil
}

func (s *SecurityService) needCheck(coin string) bool {
	coin = strings.ToLower(coin)
	needCheckArr := []string{"eth", "bsc", "hsc", "heco"}
	for _, item := range needCheckArr {
		if coin == item {
			return true
		}
	}
	return false
}

func (s *SecurityService) IsInsideAddress(addr string) (bool, error) {
	return dao.IsInsideAddress(addr)
}
