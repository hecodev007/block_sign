package merge

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/merge"
	"github.com/group-coldwallet/blockchains-go/model/rediskey"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
)

var (
	fioName = "fio"
	fioFee  = decimal.NewFromFloat(2)
)

type FioMergeService struct {
	CoinName string
	limitMap sync.Map
}

func NewFioMergeService() service.MergeService {
	return &FioMergeService{CoinName: "fio", limitMap: sync.Map{}}
}

func (b *FioMergeService) MergeCoin(params merge.MergeParams) (string, error) {
	if _, ok := conf.Cfg.Merge[strings.ToLower(params.Coin)]; !ok {
		return "", fmt.Errorf("配置文件缺少币种[%s]合并配置", params.Coin)
	}

	log.Infof("合并配置：%+v", conf.Cfg.Merge[strings.ToLower(params.Coin)])

	if params.AppId == 0 || strings.ToLower(params.Coin) != b.CoinName || len(params.Froms) == 0 || params.To == "" {
		return "", errors.New("error params")
	}

	var (
		err              error
		msg              string
		addresses        = make([]*entity.FcGenerateAddressList, 0)
		isColdAddrFrom   bool //from是否是冷地址
		isColdAddrTo     bool // to是否是冷地址
		redisHelper      *util.RedisClient
		redisKeyName     string
		coinName         string
		orderReq         *transfer.FioOrderRequest
		FioEof           decimal.Decimal
		FioMergeMin      decimal.Decimal
		FioMergeTokenMin decimal.Decimal
	)

	FioEof = conf.Cfg.Merge[strings.ToLower(params.Coin)].BalanceThreshold
	if FioEof.LessThanOrEqual(fioFee) {
		FioEof = fioFee.Mul(decimal.New(2, 0))
	}
	FioMergeMin = conf.Cfg.Merge[strings.ToLower(params.Coin)].MergeThreshold
	if FioMergeMin.LessThanOrEqual(fioFee) {
		FioMergeMin = fioFee
	}
	FioMergeTokenMin = conf.Cfg.Merge[strings.ToLower(params.Coin)].MergeTokenThreshold
	if FioMergeTokenMin.LessThanOrEqual(decimal.Zero) {
		FioMergeTokenMin = decimal.New(1, -9)
	}

	coinName = params.Coin
	if params.Token != "" {
		coinName = params.Token
	}
	redisHelper, err = util.AllocRedisClient()
	if err != nil {
		return "", err
	}

	//批量找出冷地址
	addresses, err = dao.FcGenerateAddressListFindAddressesData(address.AddressTypeCold.ToInt(), int(address.AddressStatusAlloc), params.AppId, params.Coin)
	if err != nil {
		return "", err
	}

	//不允许自己归集自己
	for _, v := range params.Froms {
		if v == params.To {
			return "", errors.New("不允许from中存在to地址")
		}
	}

	//验证to地址是否是冷地址
	for _, v := range addresses {
		if v.Address == params.To {
			isColdAddrTo = true
			break
		}
	}
	if !isColdAddrTo {
		return "", fmt.Errorf("接收地址：%s,非商户冷地址", params.To)
	}

	//验证from地址是否是冷地址,用时添加进入map临时备用
	for _, from := range params.Froms {
		isColdAddrFrom = false
		errorAddr := from
		for _, v := range addresses {
			if from == v.Address {
				isColdAddrFrom = true
				break
			}
		}
		if !isColdAddrFrom {
			return "", fmt.Errorf("地址：%s,非商户冷地址", errorAddr)
		}
	}

	//验证from地址是否是冷地址
	for _, v := range params.Froms {
		redisKeyName = fmt.Sprintf(rediskey.MERGE_LOCK, v)
		if has, _ := redisHelper.Exists(redisKeyName); has {
			msg = msg + "," + fmt.Sprintf("地址[%s],合并CD尚未冷却", v)
			continue
		} else {
			//过期时间
			redisHelper.Set(redisKeyName, params)
			redisHelper.Expire(redisKeyName, rediskey.MERGE_LOCK_SECOND_TIME)
		}

		amount := decimal.Zero

		//查找这个地址的所有币种金额
		addressAmounts, err := dao.FcAddressAmountFindAddress(address.AddressTypeCold.ToInt(), params.AppId, v)
		if err != nil {
			msg = msg + "," + fmt.Sprintf("地址[%s],查询相关的余额信息异常=[%s]", v, err.Error())
			continue

		}
		if len(addressAmounts) == 0 {
			msg = msg + "," + fmt.Sprintf("地址[%s],无法查询相关的余额信息", v)
			continue
		}

		if params.Token == "" {
			//主链币合并
			for _, am := range addressAmounts {
				if strings.ToLower(am.CoinType) == fioName {
					amDecimal, _ := decimal.NewFromString(am.Amount)
					mergeDecimal := amDecimal.Sub(FioEof.Add(fioFee))
					if mergeDecimal.LessThan(FioMergeMin) {

						msg = msg + "," + fmt.Sprintf("地址[%s],扣除手续费[%s], 保留余额[%s],归集金额[%s]，小于阈值[%s],抛弃归集",
							v, fioFee.String(), FioEof.String(), mergeDecimal.String(), FioMergeMin)

						continue
					}
					amount = mergeDecimal
				}
			}

		} else {

			var isHasFio bool
			var isHasToken bool
			//代币合并,此时需要判断主链币的余额
			for _, am := range addressAmounts {
				amDecimal, _ := decimal.NewFromString(am.Amount)
				if strings.ToLower(am.CoinType) == fioName {
					if amDecimal.GreaterThanOrEqual(fioFee) {
						isHasFio = true
					}
				}

				//代币
				if strings.ToLower(am.CoinType) == coinName {
					if amDecimal.GreaterThan(FioMergeTokenMin) {
						isHasToken = true
						amount = amDecimal
					}
				}

			}

			if !isHasFio {
				msg = msg + "," + fmt.Sprintf("地址[%s],手续费不足[%s],抛弃归集",
					v, fioFee.String())
				continue
			}
			if !isHasToken {
				msg = msg + "," + fmt.Sprintf("地址[%s],没有代币[%s],抛弃归集",
					v, coinName)
				continue
			}
		}

		if amount.LessThanOrEqual(decimal.Zero) {
			log.Infof("地址:[%s]，抛弃归集", v)
			continue
		}
		//构建交易
		orderReq, err = b.createOrder(params.AppId, params.Coin, params.Token, v, params.To, amount)
		//然后直接发送给底层交易
		if err != nil {
			log.Errorf("创建交易合并订单错误:[%s]", err.Error())
			msg = msg + "," + fmt.Sprintf("创建交易合并订单错误:[%s]", err.Error())
			continue
		}
		err = b.walletServerCreateCold(orderReq)
		if err != nil {
			log.Errorf("发送交易合并订单错误:[%s]", err.Error())
			msg = msg + "," + fmt.Sprintf("发送交易合并订单错误:[%s]", err.Error())
			continue
		}
		log.Infof("合并订单[%s],已发送", orderReq.OuterOrderNo)
		if msg != "" {
			msg = msg + ","
		}
		msg = msg + fmt.Sprintf("合并订单[%s],已发送", orderReq.OuterOrderNo)
	}
	return msg, nil

}

//返回订单ID
func (b *FioMergeService) createOrder(appid int, coin, token, from, to string, amountFloat decimal.Decimal) (*transfer.FioOrderRequest, error) {
	mch, err := dao.FcMchFindById(appid)
	if err != nil {
		return nil, err
	}
	coinName := strings.ToLower(coin)
	if token != "" {
		coinName = strings.ToLower(token)
	} else {
		token = "fio"
	}

	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("合并，DB缺少币种[%s]设置", coinName)
	}

	outOrderId := fmt.Sprintf("merge-%d-%s", appid, util.GetUUID())
	orderId, _ := util.GetOrderId(outOrderId, to, amountFloat.String())
	ta := &entity.FcTransfersApply{
		Username:   "merge",
		Department: "blockchains-go",
		Purpose:    fmt.Sprintf("%s金额合并", coinName),
		OrderId:    orderId,
		Applicant:  mch.Platform,
		AppId:      mch.Id,
		CallBack:   "",
		OutOrderid: outOrderId,
		CoinName:   coin,
		Type:       string(entity.HB_ApplyType),
		Memo:       "hb-merge",
		Eoskey:     strings.ToUpper(token),
		Fee:        fioFee.String(),
		Status:     int(entity.ApplyStatus_Merge),
		Createtime: util.GetChinaTimeNow().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
	}

	tacFrom := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     from,
		AddressFlag: "from",
		ToAmount:    amountFloat.String(),
		Lastmodify:  util.GetChinaTimeNow(),
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: coinSet.Id,
		Address:     to,
		AddressFlag: "to",
		ToAmount:    amountFloat.String(),
		Lastmodify:  util.GetChinaTimeNow(),
	}

	createId, err := dao.FcTransfersApplyCreate(ta, []*entity.FcTransfersApplyCoinAddress{tacTo, tacFrom})
	if err != nil {
		return nil, err
	}
	orderReq := &transfer.FioOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      createId,
			ApplyCoinId:  int64(coinSet.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(mch.Id),
			MchName:      mch.Platform,
			CoinName:     coin,
			Worker:       "", //不填写，直接随机
		},
		FromAddress: from,
		ToAddress:   to,
		Amount:      amountFloat.Shift(9).String(),
	}

	return orderReq, nil

}

//创建交易接口参数
func (b *FioMergeService) walletServerCreateCold(orderReq *transfer.FioOrderRequest) error {
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[orderReq.CoinName].Url+"/v1/fio/transfer", conf.Cfg.HotServers[orderReq.CoinName].User, conf.Cfg.HotServers[orderReq.CoinName].Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeFioTransferResp(data)
	if result == nil {
		return fmt.Errorf("order 请求下单接口失败，send data:%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//冻结
func (c *FioMergeService) isFrzenAddress(address string) bool {
	v, ok := c.limitMap.Load(address)
	if ok {
		lastTxTime := v.(int64)
		frzenTime := addTimeSecond(lastTxTime, 3)
		if frzenTime > time.Now().Unix() {
			return false
		}
	}
	return true
}

func addTimeSecond(now, minute int64) int64 {
	return time.Unix(now, 0).Add(time.Minute * time.Duration(minute)).Unix()
}
