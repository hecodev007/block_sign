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
	vetName    = "vet"
	vetFeeName = "vtho"
	vetDecimal = int32(18)
	vetMinFee  = decimal.NewFromFloat(100)
)

type VetMergeService struct {
	CoinName string
}

func NewVetMergeService() service.MergeService {
	return &VetMergeService{CoinName: "vet"}
}

func (b *VetMergeService) MergeCoin(pvetams merge.MergeParams) (string, error) {
	if _, ok := conf.Cfg.Merge[strings.ToLower(pvetams.Coin)]; !ok {
		return "", fmt.Errorf("配置文件缺少币种[%s]合并配置", pvetams.Coin)
	}

	log.Infof("合并配置：%+v", conf.Cfg.Merge[strings.ToLower(pvetams.Coin)])

	if pvetams.AppId == 0 || strings.ToLower(pvetams.Coin) != b.CoinName || len(pvetams.Froms) == 0 || pvetams.To == "" {
		return "", errors.New("error pvetams")
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
		orderReq         *transfer.VetOrderRequest
		vetEof           decimal.Decimal //合并金额的时候保留阈值金额在冷地址
		vetMergeMin      decimal.Decimal //主链币归集起始金额
		vetMergeTokenMin decimal.Decimal //代币归集起始金额
	)

	vetEof = conf.Cfg.Merge[strings.ToLower(pvetams.Coin)].BalanceThreshold
	if vetEof.IsZero() {
		vetEof = decimal.NewFromFloat(0.5)
	}

	vetMergeMin = conf.Cfg.Merge[strings.ToLower(pvetams.Coin)].MergeThreshold
	if vetMergeMin.IsZero() {
		vetMergeMin = decimal.NewFromFloat(1)
	}

	vetMergeTokenMin = conf.Cfg.Merge[strings.ToLower(pvetams.Coin)].MergeTokenThreshold
	if vetMergeTokenMin.IsZero() {
		vetMergeTokenMin = decimal.NewFromFloat(1)
	}
	coinName = pvetams.Coin
	if pvetams.Token != "" {
		coinName = pvetams.Token
	}
	redisHelper, err = util.AllocRedisClient()
	if err != nil {
		return "", err
	}

	//批量找出冷地址
	addresses, err = dao.FcGenerateAddressListFindAddressesData(address.AddressTypeCold.ToInt(), int(address.AddressStatusAlloc), pvetams.AppId, pvetams.Coin)
	if err != nil {
		return "", err
	}

	//不允许自己归集自己
	for _, v := range pvetams.Froms {
		if v == pvetams.To {
			return "", errors.New("不允许from中存在to地址")
		}
	}

	//验证to地址是否是冷地址
	for _, v := range addresses {
		if v.Address == pvetams.To {
			isColdAddrTo = true
			break
		}
	}
	if !isColdAddrTo {
		return "", fmt.Errorf("接收地址：%s,非商户冷地址", pvetams.To)
	}

	//验证from地址是否是冷地址,用时添加进入map临时备用
	for _, from := range pvetams.Froms {
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
	for _, v := range pvetams.Froms {
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
			redisHelper.Set(redisKeyName, pvetams)
			redisHelper.Expire(redisKeyName, rediskey.MERGE_LOCK_SECOND_TIME)
		}

		amount := decimal.Zero

		//查找这个地址的所有币种金额
		addressAmounts, err := dao.FcAddressAmountFindAddress(address.AddressTypeCold.ToInt(), pvetams.AppId, v)
		if err != nil {
			msg = msg + "," + fmt.Sprintf("地址[%s],查询相关的余额信息异常=[%s]", v, err.Error())
			continue

		}
		if len(addressAmounts) == 0 {
			msg = msg + "," + fmt.Sprintf("地址[%s],无法查询相关的余额信息", v)
			continue
		}

		if pvetams.Token == "" {
			//主链币合并
			for _, am := range addressAmounts {
				if strings.ToLower(am.CoinType) == vetName {
					amDecimal, _ := decimal.NewFromString(am.Amount)
					mergeDecimal := amDecimal.Sub(vetEof)
					log.Infof("amDecimal:%s", amDecimal.String())
					log.Infof("vetEof:%s", vetEof.String())
					log.Infof("mergeDecimal:%s", mergeDecimal.String())
					if mergeDecimal.LessThan(vetMergeMin) {
						msg = msg + "," + fmt.Sprintf("地址[%s], 保留余额[%s],归集金额[%s]，小于阈值[%s],抛弃归集",
							v, vetEof.String(), mergeDecimal.String(), vetMergeMin)

						continue
					}
					amount = mergeDecimal
					break
				}
			}

		} else {

			var isHasVtho bool
			var isHasToken bool
			//代币合并,此时需要判断vtho余额
			for _, am := range addressAmounts {
				amDecimal, _ := decimal.NewFromString(am.Amount)
				if strings.ToLower(am.CoinType) == vetFeeName {
					if vetMinFee.GreaterThanOrEqual(decimal.NewFromFloat(100.0)) {
						isHasVtho = true
					}
				}

				//代币
				if strings.ToLower(am.CoinType) == coinName {
					if amDecimal.GreaterThan(vetMergeTokenMin) {
						isHasToken = true
						amount = amDecimal
					}
				}

			}

			if !isHasVtho {
				msg = fmt.Sprintf("地址[%s],手续费不足[%s],抛弃归集",
					v, vetMinFee.String())
				continue
			}
			if !isHasToken {
				msg = fmt.Sprintf("地址[%s],没有代币[%s],抛弃归集",
					v, coinName)
				continue
			}
		}

		if strings.ToLower(coinName) == strings.ToLower(vetFeeName) {
			amount = amount.Sub(decimal.NewFromFloat(100.0))
			if amount.LessThanOrEqual(decimal.NewFromFloat(100.0)) {
				log.Infof("地址:[%s]，抛弃归集,手续费不足", v)
				continue
			}
		}
		if amount.LessThanOrEqual(decimal.Zero) {
			log.Infof("地址:[%s]，抛弃归集", v)
			continue
		}

		//构建交易
		orderReq, err = b.createOrder(pvetams.AppId, pvetams.Coin, pvetams.Token, v, pvetams.To, amount)
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
func (b *VetMergeService) createOrder(appid int, coin, token, from, to string, amountFloat decimal.Decimal) (*transfer.VetOrderRequest, error) {
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
		Eostoken:   coinSet.Token,
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
	orderReq := &transfer.VetOrderRequest{
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
		Data: []transfer.VetData{transfer.VetData{SubData: transfer.VetSubData{
			From:            from,
			ContractAddress: coinSet.Token,
			CoinName:        coinSet.Name,
			Tolist: []transfer.VetToList{transfer.VetToList{
				To:     to,
				Amount: amountFloat.Shift(int32(coinSet.Decimal)).String(),
			}},
		}}}}

	return orderReq, nil

}

//创建交易接口参数
//交易
func (s *VetMergeService) walletServerCreateHot(orderReq *transfer.VetOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", errors.New("json unmarshal response data error")
	}
	if resp["error"] != nil {
		return "", fmt.Errorf("vet transfer error,Err=%v", resp["error"])
	}
	return resp["result"].(string), nil
}
