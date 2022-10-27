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
	"strconv"
	"strings"
)

var (
	mtrName    = "mtr"
	mtrDecimal = int32(18)
	mtrMinFee  = decimal.NewFromFloat(0.02)
)

type MtrMergeService struct {
	CoinName string
}

func NewMtrMergeService() service.MergeService {
	return &MtrMergeService{CoinName: "mtr"}
}

func (b *MtrMergeService) MergeCoin(pmtrams merge.MergeParams) (string, error) {
	if _, ok := conf.Cfg.Merge[strings.ToLower(pmtrams.Coin)]; !ok {
		return "", fmt.Errorf("配置文件缺少币种[%s]合并配置", pmtrams.Coin)
	}

	log.Infof("合并配置：%+v", conf.Cfg.Merge[strings.ToLower(pmtrams.Coin)])

	if pmtrams.AppId == 0 || strings.ToLower(pmtrams.Coin) != b.CoinName || len(pmtrams.Froms) == 0 || pmtrams.To == "" {
		return "", errors.New("error pmtrams")
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
		orderReq         *transfer.MtrOrderRequest
		mtrEof           decimal.Decimal //合并金额的时候保留阈值金额在冷地址
		mtrMergeMin      decimal.Decimal //主链币归集起始金额
		mtrMergeTokenMin decimal.Decimal //代币归集起始金额
	)

	mtrEof = conf.Cfg.Merge[strings.ToLower(pmtrams.Coin)].BalanceThreshold
	if mtrEof.IsZero() {
		mtrEof = decimal.NewFromFloat(0.5)
	}

	mtrMergeMin = conf.Cfg.Merge[strings.ToLower(pmtrams.Coin)].MergeThreshold
	if mtrMergeMin.IsZero() {
		mtrMergeMin = decimal.NewFromFloat(1)
	}

	mtrMergeTokenMin = conf.Cfg.Merge[strings.ToLower(pmtrams.Coin)].MergeTokenThreshold
	if mtrMergeTokenMin.IsZero() {
		mtrMergeTokenMin = decimal.NewFromFloat(1)
	}
	coinName = pmtrams.Coin
	if pmtrams.Token != "" {
		coinName = pmtrams.Token
	}
	redisHelper, err = util.AllocRedisClient()
	if err != nil {
		return "", err
	}

	//批量找出冷地址
	addresses, err = dao.FcGenerateAddressListFindAddressesData(address.AddressTypeCold.ToInt(), int(address.AddressStatusAlloc), pmtrams.AppId, pmtrams.Coin)
	if err != nil {
		return "", err
	}

	//不允许自己归集自己
	for _, v := range pmtrams.Froms {
		if v == pmtrams.To {
			return "", errors.New("不允许from中存在to地址")
		}
	}

	//验证to地址是否是冷地址
	for _, v := range addresses {
		if v.Address == pmtrams.To {
			isColdAddrTo = true
			break
		}
	}
	if !isColdAddrTo {
		return "", fmt.Errorf("接收地址：%s,非商户冷地址", pmtrams.To)
	}

	//验证from地址是否是冷地址,用时添加进入map临时备用
	for _, from := range pmtrams.Froms {
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
	for _, v := range pmtrams.Froms {
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
			redisHelper.Set(redisKeyName, pmtrams)
			redisHelper.Expire(redisKeyName, rediskey.MERGE_LOCK_SECOND_TIME)
		}

		amount := decimal.Zero

		//查找这个地址的所有币种金额
		addressAmounts, err := dao.FcAddressAmountFindAddress(address.AddressTypeCold.ToInt(), pmtrams.AppId, v)
		if err != nil {
			msg = msg + "," + fmt.Sprintf("地址[%s],查询相关的余额信息异常=[%s]", v, err.Error())
			continue

		}
		if len(addressAmounts) == 0 {
			msg = msg + "," + fmt.Sprintf("地址[%s],无法查询相关的余额信息", v)
			continue
		}

		if pmtrams.Token == "" {
			//主链币合并
			for _, am := range addressAmounts {
				if strings.ToLower(am.CoinType) == mtrName {
					amDecimal, _ := decimal.NewFromString(am.Amount)
					mergeDecimal := amDecimal.Sub(mtrEof)
					log.Infof("amDecimal:%s", amDecimal.String())
					log.Infof("mtrEof:%s", mtrEof.String())
					log.Infof("mergeDecimal:%s", mergeDecimal.String())
					if mergeDecimal.LessThan(mtrMergeMin) {
						msg = msg + "," + fmt.Sprintf("地址[%s], 保留余额[%s],归集金额[%s]，小于阈值[%s],抛弃归集",
							v, mtrEof.String(), mergeDecimal.String(), mtrMergeMin)

						continue
					}
					amount = mergeDecimal
					break
				}
			}

		} else {

			var isHasMtr bool
			var isHasToken bool
			//代币合并,此时需要判断主链币的余额
			for _, am := range addressAmounts {
				log.Infof("am.CoinType:%s,coinName:%s", am.CoinType, coinName)
				amDecimal, _ := decimal.NewFromString(am.Amount)
				if strings.ToLower(am.CoinType) == mtrName {
					if amDecimal.GreaterThanOrEqual(mtrMinFee) {
						isHasMtr = true
					}
				}

				//代币
				if strings.ToLower(am.CoinType) == coinName {
					if amDecimal.GreaterThan(mtrMergeTokenMin) {
						isHasToken = true
						amount = amDecimal
					}
				}

			}

			if !isHasMtr {
				if msg == "" {
					msg = msg + ","
				}
				msg = msg + fmt.Sprintf("地址[%s],手续费不足[%s],抛弃归集",
					v, mtrMinFee.String())
				continue
			}
			if !isHasToken {
				if msg == "" {
					msg = msg + ","
				}
				msg = msg + fmt.Sprintf("地址[%s],没有代币[%s],抛弃归集",
					v, coinName)
				continue
			}
		}

		if amount.LessThanOrEqual(decimal.Zero) {
			log.Infof("地址:[%s]，抛弃归集", v)
			continue
		}
		//构建交易
		orderReq, err = b.createOrder(pmtrams.AppId, pmtrams.Coin, pmtrams.Token, v, pmtrams.To, amount)
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
func (b *MtrMergeService) createOrder(appid int, coin, token, from, to string, amountFloat decimal.Decimal) (*transfer.MtrOrderRequest, error) {
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

	if coinSet.Token == "" {
		coinSet.Token = "0"
	}
	tokenId, err := strconv.ParseInt(coinSet.Token, 10, 64)
	if err != nil {
		return nil, err
	}
	orderReq := &transfer.MtrOrderRequest{
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
		FromAddr:      from,
		ToAddr:        to,
		ToAmountInt64: amountFloat.Shift(mtrDecimal).String(),
		Token:         tokenId,
	}

	return orderReq, nil

}

//创建交易接口参数
func (b *MtrMergeService) walletServerCreateHot(orderReq *transfer.MtrOrderRequest) (string, error) {
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
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,err:%s", orderReq.OuterOrderNo, err.Error())

	}
	if resp["data"] == nil {
		return "", fmt.Errorf("mtr transfer error,Err=%s", string(data))
	}
	return resp["data"].(string), nil

}
