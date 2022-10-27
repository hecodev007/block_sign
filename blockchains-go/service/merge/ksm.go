package merge

import (
	"encoding/json"
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
)

var (
	ksmName    = "ksm"
	ksmDecimal = int32(12)
	ksmMinFee  = decimal.NewFromFloat(0.001)
)

type KsmMergeService struct {
	CoinName string
}

func NewKsmMergeService() service.MergeService {
	return &KsmMergeService{CoinName: "ksm"}
}

func (b *KsmMergeService) MergeCoin(pksmams merge.MergeParams) (string, error) {
	if _, ok := conf.Cfg.Merge[strings.ToLower(pksmams.Coin)]; !ok {
		return "", fmt.Errorf("配置文件缺少币种[%s]合并配置", pksmams.Coin)
	}

	log.Infof("合并配置：%+v", conf.Cfg.Merge[strings.ToLower(pksmams.Coin)])

	if pksmams.AppId == 0 || strings.ToLower(pksmams.Coin) != b.CoinName || len(pksmams.Froms) == 0 || pksmams.To == "" {
		return "", errors.New("error pksmams")
	}

	var (
		err              error
		txid             string
		msg              string
		addresses        = make([]*entity.FcGenerateAddressList, 0)
		isColdAddrFrom   bool //from是否是冷地址
		isColdAddrTo     bool // to是否是冷地址
		redisHelper      *util.RedisClient
		redisKeyName     string
		coinName         string
		orderReq         *transfer.KsmOrderRequest
		ksmEof           decimal.Decimal //合并金额的时候保留阈值金额在冷地址
		ksmMergeMin      decimal.Decimal //主链币归集起始金额
		ksmMergeTokenMin decimal.Decimal //代币归集起始金额
	)

	ksmEof = conf.Cfg.Merge[strings.ToLower(pksmams.Coin)].BalanceThreshold
	if ksmEof.IsZero() {
		ksmEof = decimal.NewFromFloat(0.5)
	}

	ksmMergeMin = conf.Cfg.Merge[strings.ToLower(pksmams.Coin)].MergeThreshold
	if ksmMergeMin.IsZero() {
		ksmMergeMin = decimal.NewFromFloat(1)
	}

	ksmMergeTokenMin = conf.Cfg.Merge[strings.ToLower(pksmams.Coin)].MergeTokenThreshold
	if ksmMergeTokenMin.IsZero() {
		ksmMergeTokenMin = decimal.NewFromFloat(1)
	}
	coinName = pksmams.Coin
	if pksmams.Token != "" {
		coinName = pksmams.Token
	}
	redisHelper, err = util.AllocRedisClient()
	if err != nil {
		return "", err
	}

	//批量找出冷地址
	addresses, err = dao.FcGenerateAddressListFindAddressesData(address.AddressTypeCold.ToInt(), int(address.AddressStatusAlloc), pksmams.AppId, pksmams.Coin)
	if err != nil {
		return "", err
	}

	//不允许自己归集自己
	for _, v := range pksmams.Froms {
		if v == pksmams.To {
			return "", errors.New("不允许from中存在to地址")
		}
	}

	//验证to地址是否是冷地址
	for _, v := range addresses {
		if v.Address == pksmams.To {
			isColdAddrTo = true
			break
		}
	}
	if !isColdAddrTo {
		return "", fmt.Errorf("接收地址：%s,非商户冷地址", pksmams.To)
	}

	//验证from地址是否是冷地址,用时添加进入map临时备用
	isColdAddrFrom = false
	errorAddr := ""
	for _, from := range pksmams.Froms {
		errorAddr = from
		for _, v := range addresses {
			if from == v.Address {
				isColdAddrFrom = true
				errorAddr = ""
				break
			}
		}
		if errorAddr != "" {
			break
		}
	}
	if !isColdAddrFrom {
		return "", fmt.Errorf("地址：%s,非商户冷地址", errorAddr)
	}

	//验证from地址是否是冷地址
	for _, v := range pksmams.Froms {
		redisKeyName = fmt.Sprintf(rediskey.MERGE_LOCK, v)
		if has, _ := redisHelper.Exists(redisKeyName); has {
			if msg == "" {
				msg = fmt.Sprintf("地址[%s],合并CD尚未冷却", v)
			} else {
				msg = msg + "," + fmt.Sprintf("地址[%s],合并CD尚未冷却", v)
			}
			continue
		} else {
			//过期时间
			redisHelper.Set(redisKeyName, pksmams)
			redisHelper.Expire(redisKeyName, rediskey.MERGE_LOCK_SECOND_TIME)
		}

		amount := decimal.Zero

		//查找这个地址的所有币种金额
		addressAmounts, err := dao.FcAddressAmountFindAddress(address.AddressTypeCold.ToInt(), pksmams.AppId, v)
		if err != nil {
			msg = msg + "," + fmt.Sprintf("地址[%s],查询相关的余额信息异常=[%s]", v, err.Error())
			continue

		}
		if len(addressAmounts) == 0 {
			msg = msg + "," + fmt.Sprintf("地址[%s],无法查询相关的余额信息", v)
			continue
		}

		if pksmams.Token == "" {
			//主链币合并
			for _, am := range addressAmounts {
				if strings.ToLower(am.CoinType) == ksmName {
					amDecimal, _ := decimal.NewFromString(am.Amount)
					mergeDecimal := amDecimal.Sub(ksmEof)
					log.Infof("amDecimal:%s", amDecimal.String())
					log.Infof("ksmEof:%s", ksmEof.String())
					log.Infof("mergeDecimal:%s", mergeDecimal.String())
					if mergeDecimal.LessThan(ksmMergeMin) {
						msg = msg + "," + fmt.Sprintf("地址[%s], 保留余额[%s],归集金额[%s]，小于阈值[%s],抛弃归集",
							v, ksmEof.String(), mergeDecimal.String(), ksmMergeMin)

						continue
					}
					amount = mergeDecimal
					break
				}
			}

		} else {

			var isHasKsm bool
			var isHasToken bool
			//代币合并,此时需要判断主链币的余额
			for _, am := range addressAmounts {
				amDecimal, _ := decimal.NewFromString(am.Amount)
				if strings.ToLower(am.CoinType) == ksmName {
					if amDecimal.GreaterThanOrEqual(ksmMinFee) {
						isHasKsm = true
					}
				}

				//代币
				if strings.ToLower(am.CoinType) == coinName {
					if amDecimal.GreaterThan(ksmMergeTokenMin) {
						isHasToken = true
						amount = amDecimal
					}
				}

			}

			if !isHasKsm {
				msg = msg + "," + fmt.Sprintf("地址[%s],手续费不足[%s],抛弃归集",
					v, ksmMinFee.String())
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
		orderReq, err = b.createOrder(pksmams.AppId, pksmams.Coin, pksmams.Token, v, pksmams.To, amount)
		//然后直接发送给底层交易
		if err != nil {
			log.Errorf("创建交易合并订单错误:[%s]", err.Error())
			msg = msg + "," + fmt.Sprintf("创建交易合并订单错误:[%s]", err.Error())
			continue
		}
		txid, err = b.walletServerCreateHot(orderReq)
		if err != nil {
			log.Errorf("发送交易合并订单错误:[%s]", err.Error())
			msg = msg + "," + fmt.Sprintf("发送交易合并订单错误:[%s]", err.Error())
			continue
		}
		log.Infof("合并订单[%s]已发送\ntxid:%s\n金额：%s", orderReq.OuterOrderNo, txid, amount.String())
		if msg != "" {
			msg = msg + ","
		}
		msg = msg + fmt.Sprintf("合并订单[%s]已发送\ntxid:%s\n金额：%s", orderReq.OuterOrderNo, txid, amount.String())
	}
	return msg, nil

}

//返回订单ID
func (b *KsmMergeService) createOrder(appid int, coin, token, from, to string, amountFloat decimal.Decimal) (*transfer.KsmOrderRequest, error) {
	mch, err := dao.FcMchFindById(appid)
	if err != nil {
		return nil, err
	}
	coinName := strings.ToLower(coin)
	if token != "" {
		coinName = strings.ToLower(token)
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
		Fee:        "0",
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
	orderReq := &transfer.KsmOrderRequest{
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
		Amount:      amountFloat.Shift(ksmDecimal).String(),
	}

	return orderReq, nil

}

//创建交易接口参数
func (b *KsmMergeService) walletServerCreateHot(orderReq *transfer.KsmOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[b.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", b.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, b.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", b.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", b.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Data.(string), nil
}
