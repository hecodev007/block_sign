package recycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type HcRecycleService struct {
	CoinName string
}

func NewHcRecycleService() service.RecycleService {
	return &HcRecycleService{CoinName: "hc"}
}

//params model : 0小额合并 1大额合并
func (b *HcRecycleService) RecycleCoin(reqHead *transfer.OrderRequestHead, toAddr string, feeFloat decimal.Decimal, model int) (msg string, err error) {
	var (
		//transPush []*entity.FcTransPush
		//scanNum   int
		tpl *transfer.HcRecycleOrderRequest //模板
	)
	////获取币种的配置
	//hcCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": reqHead.CoinName}))
	//if err != nil {
	//	return "", err
	//}
	//if len(hcCoins) == 0 {
	//	return "", errors.New("do not find dot coin")
	//}
	////只有一个coin
	//coin := hcCoins[0]
	//if conf.Cfg.UtxoScan.Num <= 0 {
	//	scanNum = 10
	//} else {
	//	scanNum = conf.Cfg.UtxoScan.Num
	//}
	//scanNum = 10
	//step1：to地址
	if toAddr == "" {
		return "", errors.New("缺少to地址")
	}
	//redisHelper, err := util.AllocRedisClient()
	//if err != nil {
	//	return "", err
	//}
	////step2：判断模式，小的合并还是大的合并，查询相关地址
	//if model == 0 {
	//	//小金额回收
	//	transPush, err = dao.FcTransPushAddressValidUTXO(reqHead.MchId, reqHead.CoinName, scanNum, "asc")
	//} else {
	//	//大金额回收
	//	transPush, err = dao.FcTransPushAddressValidUTXO(reqHead.MchId, reqHead.CoinName, scanNum, "desc")
	//}
	////transPush, err = dao.FcTransPushAddressValidUTXO(reqHead.MchId, reqHead.CoinName, scanNum, "")
	//
	////计算总的amount
	//var totalAmount = decimal.Zero
	//totalNum := 0
	//for _, utxo := range transPush {
	//	rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.Hc_UTXO_LOCK, utxo.TransactionId, utxo.TrxN)
	//	if has, _ := redisHelper.Exists(rediskeyName); has {
	//		//已经占用utxo，跳过
	//		continue
	//	}
	//	amount, _ := decimal.NewFromString(utxo.Amount)
	//	totalAmount = totalAmount.Add(amount)
	//	//临时存储进入redis 锁定3分钟
	//	redisHelper.Set(rediskeyName, reqHead.OuterOrderNo)
	//	redisHelper.Expire(rediskeyName, rediskey.Hc_UTXO_LOCK_SECOND_TIME)
	//	totalNum++
	//}
	//if totalNum < 2 {
	//	return "", errors.New("UTXO数量过少")
	//}
	//log.Infof("hc 零散归集总查找金额为： %s", totalAmount.String())
	////totalAmount = totalAmount.Sub(decimal.NewFromFloat(0.1))
	//tmFloat64, _ := totalAmount.Float64()
	//
	////floorAmount := math.Floor(tmFloat64)
	//log.Infof("hc 零散归集向下取整金额为： %v", tmFloat64)
	//if tmFloat64 <= 0 {
	//	return "", fmt.Errorf("零散回收金额为：%f ", tmFloat64)
	//}
	//toAmount := decimal.NewFromFloat(tmFloat64).Shift(int32(coin.Decimal)).String()
	//var toList []*transfer.HcOrderToAddressList
	//tl := &transfer.HcOrderToAddressList{
	//	Address:  toAddr,
	//	Quantity: toAmount,
	//}
	//toList = append(toList, tl)
	////查询找零地址
	//changes, err := dao.FcGenerateAddressListFindChangeAddr(int(reqHead.MchId), reqHead.CoinName)
	//if err != nil {
	//	return "", err
	//}
	//if len(changes) == 0 {
	//	return "", fmt.Errorf("商户=[%d],查询%s找零地址失败", reqHead.MchId, reqHead.CoinName)
	//}
	////随机选择
	//randIndex := util.RandInt64(0, int64(len(changes)))
	//changeAddress := changes[randIndex]
	//构建订单
	tpl = &transfer.HcRecycleOrderRequest{
		MchId:      reqHead.MchName,
		OutOrderId: reqHead.OuterOrderNo,
		ApplyId:    reqHead.ApplyId,
		ToAddress:  toAddr,
		CoinName:   reqHead.CoinName,
		Model:      model,
		FeeFloat:   feeFloat,
	}
	//tpl = &transfer.HcOrderRequest{
	//	OrderRequestHead: transfer.OrderRequestHead{
	//		ApplyId:      reqHead.ApplyId,
	//		ApplyCoinId:  reqHead.ApplyCoinId,
	//		OuterOrderNo: reqHead.OuterOrderNo,
	//		OrderNo:      reqHead.OrderNo,
	//		MchId:        reqHead.MchId,
	//		MchName:      reqHead.MchName,
	//		CoinName:     reqHead.CoinName,
	//		Worker:       reqHead.Worker,
	//	},
	//	ChangeAddress: changeAddress,
	//	ToList:        toList,
	//}

	err = b.walletServerCreate(tpl)
	if err != nil {
		return "", fmt.Errorf("hc 零散回收失败，模式：%d，err:%s", model, err.Error())
	}
	return fmt.Sprintf("hc 零散合并成功，模式%d，outOrderId:%s", model, reqHead.OuterOrderNo), nil
}

func (srv *HcRecycleService) walletServerCreate(orderReq *transfer.HcRecycleOrderRequest) error {
	params, _ := json.Marshal(orderReq)
	log.Infof("发送内容：%s", string(params))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/hc/collect", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,data:%s", orderReq.OutOrderId, string(data))
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OutOrderId)
	}
	return nil
}
