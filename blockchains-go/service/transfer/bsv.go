package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type BsvTransferService struct {
	CoinName string
}

func NewBsvTransferService() service.TransferService {
	return &BsvTransferService{
		CoinName: "bsv",
	}
}

func (srv *BsvTransferService) VaildAddr(address string) error {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/bsv/validateaddress?address=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	BsvResp := transfer.DecodeBsvAddressResult(data)
	if BsvResp != nil && BsvResp.Data != nil {
		if BsvResp.Data.Isvalid {
			return nil
		}
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	return err
}

func (srv *BsvTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}

func (srv *BsvTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, err := srv.getEstimateTpl(ta, workerId)
	if err != nil {
		return err
	}
	err = srv.walletServerCreate(orderReq)
	if err != nil {
		//改变表状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		return err
	}
	return nil
}

//params worker 指定机器
func (srv *BsvTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.BsvOrderRequest, error) {
	var (
		changeAddress string
		toAddr        string
		//toAmount      int64
		//fee           int64
		coinSet *entity.FcCoinSet //db币种配置
	)

	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常")
	}

	//查询找零地址
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("Bsv 商户=[%d],查询Bsv找零地址失败", ta.AppId)
	}
	//随机选择
	randIndex := util.RandInt64(0, int64(len(changes)))
	changeAddress = changes[randIndex]

	//查询出账地址和金额
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	if toAddrAmount.LessThan(decimal.NewFromFloat(0.00000546)) {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额0.00000546", ta.Id, ta.OutOrderid)
	}
	//Bsv api需要金额整型
	toAmount := toAddrAmount.Shift(8).IntPart()

	oa := transfer.BsvOrderAddressRequest{
		Address: toAddr,
		Amount:  toAmount,
	}
	orderReq := &transfer.BsvOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     srv.CoinName,
			Worker:       worker,
		},
		FromAddress: changeAddress,
		OrderAddress: []*transfer.BsvOrderAddressRequest{
			&oa,
		},
	}

	return orderReq, nil
}

//创建交易接口参数
func (srv *BsvTransferService) walletServerCreate(orderReq *transfer.BsvOrderRequest) error {
	dd, _ := json.Marshal(orderReq)
	log.Infof("bsv 发送：%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/bsv/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil

}

//为orderReq的 OrderAddress组装交易内容
//toAddr 接收地址
//changeAddr 找零地址
//appid 商户ID
//orderReq walletsever交易结构
//fee 手续费
//func (srv *BsvTransferService) setUtxoData(appid int64, orderReq *transfer.BsvOrderRequest, toAddr, changeAddr string) error {
//	var (
//		mchAmount    decimal.Decimal //商户余额
//		fromAmount   decimal.Decimal //发送总金额
//		toAmount     decimal.Decimal //接收总金额
//		feeAmount    decimal.Decimal //手续费
//		fee          int64
//		feeAmountTmp decimal.Decimal   //临时估算手续费
//		feeTmp       int64             //临时估算手续费
//		changeAmount decimal.Decimal   //找零金额
//		BsvCfg       *conf.CoinServers //币种数据服务
//		coinSet      *entity.FcCoinSet //db币种配置
//		unspents     *transfer.BsvUnspents
//		err          error
//	)
//	redisHelper, err := util.AllocRedisClient()
//	if err != nil {
//		return err
//	}
//
//	utxoTpl := make([]*transfer.BsvOrderAddrRequest, 0) //utxo模板
//
//	if toAddr == "" {
//		return errors.New("结构缺少to地址")
//	}
//	if changeAddr == "" {
//		return errors.New("结构缺少change地址")
//	}
//	if orderReq.Amount < 546 {
//		return fmt.Errorf("发送金额不能小于0.00000546,目前金额：%s", decimal.New(orderReq.Amount, -8).String())
//	}
//	fee = orderReq.Fee
//
//	toAmount = decimal.New(orderReq.Amount, 0)
//
//	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
//	if err != nil {
//		return fmt.Errorf("查询商户余额异常：%s", err.Error())
//	}
//	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
//	if mchAmount.Shift(8).LessThanOrEqual(toAmount) {
//		return fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
//			appid,
//			srv.CoinName,
//			mchAmount.String(),
//			toAmount.Shift(-8).String(),
//		)
//	}
//	BsvCfg = conf.Cfg.CoinServers[srv.CoinName]
//	if BsvCfg == nil {
//		return errors.New("配置文件缺少Bsv coinservers设置")
//	}
//	//排序找金额15个地址
//	coinSet = global.CoinDecimal[srv.CoinName]
//	if coinSet == nil {
//		return errors.New("读取 coinSet 设置异常")
//	}
//	//出账的话优先使用前面50个大金额地址
//	addrInfos, err := dao.FcAddressAmountFindTransferToBsv(appid, 50, "desc")
//	if err != nil {
//		return err
//	}
//	if len(addrInfos) == 0 {
//		return fmt.Errorf("订单：%s，暂无可用地址出账", orderReq.OuterOrderNo)
//	}
//
//	addrs := make([]string, 0)
//	for _, v := range addrInfos {
//		addrs = append(addrs, v.Address)
//	}
//	//查询utxo数量
//	byteData, err := util.PostJson(BsvCfg.Url+"/api/v1/Bsv/unspents", addrs)
//	if err != nil {
//		return fmt.Errorf("获取utxo异常，err:%s", err.Error())
//	}
//	unspents = new(transfer.BsvUnspents)
//	err = json.Unmarshal(byteData, unspents)
//	if err != nil {
//		return fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
//	}
//	if unspents.Code != 0 {
//		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
//	}
//	if len(unspents.Data) == 0 {
//		return errors.New("Bsv empty unspents")
//	}
//
//	if len(unspents.Data) < 50 {
//		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("商户=[%d]，币种=[%s]，utxo数量：%d", appid, srv.CoinName, len(unspents.Data)))
//	}
//
//	if fee != 0 {
//		if fee < 546 || fee > 1000000 {
//			//使用指定手续费
//			return errors.New("指定的手续费不在合理范围值[[0.00000546-0.1]")
//		}
//		feeTmp = fee
//	} else {
//		//先提前预估手续费
//		feeTmp, err = srv.getFee(conf.Cfg.UtxoScan.Num, 2)
//		if err != nil {
//			return err
//		}
//	}
//
//	feeAmountTmp = decimal.New(feeTmp, 0)
//
//	//排序unspent，先进行降序，找出大额的数值
//	var sortUtxo transfer.BsvUnspentDesc
//
//	var sortUtxoTmp transfer.BsvUnspentAsc //临时使用，小金额排序
//	sortUtxoTmp = append(sortUtxoTmp, unspents.Data...)
//	sort.Sort(sortUtxoTmp)
//	//第一次遍历查询最优出账金额utxo
//	for _, uv := range sortUtxoTmp {
//		if uv.Confirmations == 0 {
//			continue
//		}
//		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.Bsv_UTXO_LOCK, uv.Txid, uv.Vout)
//		if has, _ := redisHelper.Exists(rediskeyName); has {
//			//已经占用utxo，跳过
//			continue
//		}
//		am := decimal.New(uv.Amount, 0)
//		if am.GreaterThanOrEqual(toAmount.Add(feeAmountTmp)) {
//			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, am.String(), uv.Address)
//			sortUtxo = append(sortUtxo, uv)
//			break
//		}
//	}
//	if len(sortUtxo) == 0 {
//		sortUtxo = append(sortUtxo, unspents.Data...)
//	}
//	sort.Sort(sortUtxo)
//
//	//组装from
//	for _, v := range sortUtxo {
//		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.Bsv_UTXO_LOCK, v.Txid, v.Vout)
//		if has, _ := redisHelper.Exists(rediskeyName); has {
//			//已经占用utxo，跳过
//			continue
//		}
//		if v.Confirmations == 0 {
//			//暂时过滤
//			continue
//		}
//		oar := &transfer.BsvOrderAddrRequest{
//			Dir:     transfer.DirTypeFrom,
//			Address: v.Address,
//			Amount:  v.Amount,
//			TxID:    v.Txid,
//			Vout:    v.Vout,
//		}
//		fromAmount = fromAmount.Add(decimal.New(v.Amount, 0))
//		utxoTpl = append(utxoTpl, oar)
//		//临时存储进入redis 锁定2分钟
//		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
//		redisHelper.Expire(rediskeyName, rediskey.Bsv_UTXO_LOCK_SECOND_TIME)
//
//		if fromAmount.GreaterThan(toAmount.Add(feeAmountTmp)) {
//			//满足出账
//			break
//		}
//		if len(utxoTpl) == conf.Cfg.UtxoScan.Num {
//			//为了保证扫码稳定性 只使用15个utxo
//			break
//		}
//	}
//	if fromAmount.LessThan(toAmount) {
//		return fmt.Errorf("使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,出账金额：%s，预估手续费：%s",
//			mchAmount.String(),
//			fromAmount.Shift(-8).String(),
//			toAmount.Shift(-8).String(),
//			feeAmountTmp.Shift(-8).String(),
//		)
//	}
//
//	//实际使用手续费
//	if fee == 0 {
//		fee, err = srv.getFee(len(utxoTpl), 2)
//		if err != nil {
//			return err
//		}
//	}
//	feeAmount = decimal.New(fee, 0)
//
//	//组装to
//	utxoTpl = append(utxoTpl, &transfer.BsvOrderAddrRequest{
//		Dir:     transfer.DirTypeTo,
//		Address: toAddr,
//		Amount:  toAmount.IntPart(),
//	})
//
//	//计算找零金额
//	changeAmount = fromAmount.Sub(toAmount).Sub(feeAmount)
//	if changeAmount.LessThan(decimal.Zero) {
//		return fmt.Errorf("找零金额异常使用金额：%s,出账金额：%s，手续费：%s，找零：%s",
//			fromAmount.Shift(-8).String(),
//			toAmount.Shift(-8).String(),
//			feeAmount.Shift(-8).String(),
//			changeAmount.Shift(-8).String(),
//		)
//	}
//	if changeAmount.LessThanOrEqual(decimal.New(546, 0)) {
//		//如果找零小于546，那么附加在手续费上
//		feeAmount = feeAmount.Add(changeAmount)
//	} else {
//		//组装找零
//		utxoTpl = append(utxoTpl, &transfer.BsvOrderAddrRequest{
//			Dir:     transfer.DirTypeChange,
//			Address: changeAddr,
//			Amount:  changeAmount.IntPart(),
//		})
//	}
//	orderReq.OrderAddress = utxoTpl
//	orderReq.Fee = fee
//	return nil
//
//}

//手续费计算
//func (srv *BsvTransferService) getFee(inNum, outNum int) (int64, error) {
//
//	//默认费率
//	rate := int64(20)
//	if inNum <= 0 {
//		return 0, errors.New(fmt.Sprintf("Error InNum"))
//	}
//	if outNum <= 0 {
//		return 0, errors.New(fmt.Sprintf("Error OutNum"))
//	}
//	//近似值字节数
//	//byteNum := int64(inNum*148 + 34*outNum + 10)
//	//提高输出字节，加速出块
//	byteNum := int64((inNum)*148 + 34*outNum + 10)
//
//	respData, err := util.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
//	if err != nil {
//		log.Errorf("获取在线费率失败，将会使用默认费率：%d", rate)
//	} else {
//		result := &transfer.BsvGasResult{}
//		result, err = transfer.DecodeBsvGasResult(respData)
//		if err != nil {
//			log.Errorf("解析在线费率，将会使用默认费率：%d", rate)
//		} else {
//			rate = result.HalfHourFee
//		}
//
//	}
//	//使用最快优先级手续费
//	fee := rate * byteNum
//	//限定最小值
//	if fee < 1000 {
//		fee = 1000
//	}
//	//限制最大
//	if fee > 100000 {
//		fee = 100000
//	}
//	return fee, nil
//}

//手续费计算
//func (srv *BsvTransferService) getFee(inNum, outNum int) (int64, error) {
//
//	var (
//		rate int64
//	)
//	redisHelper, err := util.AllocRedisClient()
//	if err != nil {
//		return 0, err
//	}
//
//	//默认费率
//	if inNum <= 0 {
//		return 0, errors.New(fmt.Sprintf("Error InNum"))
//	}
//	if outNum <= 0 {
//		return 0, errors.New(fmt.Sprintf("Error OutNum"))
//	}
//	//近似值字节数
//	//byteNum := int64(inNum*148 + 34*outNum + 10)
//	//提高输出字节，加速出块
//	byteNum := int64((inNum)*148 + 34*outNum + 10)
//
//	if has, _ := redisHelper.Exists(rediskey.Bsv_RATE); has {
//		rateStr, _ := redisHelper.Get(rediskey.Bsv_RATE)
//		rate, _ = strconv.ParseInt(rateStr, 10, 64)
//		log.Infof("Bsv 读取缓存的rate:%d", rate)
//	} else {
//		respData, err := util.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
//		if err != nil {
//			log.Errorf("Bsv获取在线费率失败，将会使用默认费率：%d", rate)
//		} else {
//			result := &transfer.UsdtGasResult{}
//			result, err = transfer.DecodeUsdtGasResult(respData)
//			if err != nil {
//				log.Errorf("Bsv解析在线费率，将会使用默认费率：%d", rate)
//			} else {
//				rate = result.HalfHourFee
//				redisHelper.Set(rediskey.Bsv_RATE, rate)
//				redisHelper.Expire(rediskey.Bsv_RATE, 600) //10分钟过期
//			}
//		}
//
//	}
//	if rate == 0 {
//		rate = 20
//	}
//	fee := rate * byteNum
//	//限定最小值
//	if fee < 1000 {
//		fee = 1000
//	}
//	//限制最大
//	if fee > 200000 {
//		fee = 200000
//	}
//	return fee, nil
//}
