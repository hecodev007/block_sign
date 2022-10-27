package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/rediskey"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type AdaTransferService struct {
	CoinName  string
	minAmount int64
}

func NewAdaTransferService() service.TransferService {
	return &AdaTransferService{
		CoinName:  "ada",
		minAmount: 1000000,
	}
}

func (srv *AdaTransferService) VaildAddr(address string) error {
	url := conf.Cfg.HotServers[srv.CoinName].Url + "/api/v1/ada/validAddress?address=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	AdaResp := transfer.DecodeAdaAddressResult(data)
	if AdaResp != nil && AdaResp.Data != nil {
		if AdaResp.Data.Isvalid {
			return nil
		}
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	return err
}

func (srv *AdaTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {

	coinSet := global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	mch, err := dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	workerId := service.GetWorker(srv.CoinName)
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return "", err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return "", fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr := toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	orderReq, err := srv.getEstimateTpl(ta, workerId)
	if err != nil {
		return "", err
	}
	createData, _ := json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  "",
		ToAddress:    toAddr,
		Token:        ta.Eostoken,
		Amount:       toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:     toAddrAmount.Shift(int32(coinSet.Decimal)).String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}
	txid, err = srv.walletServerCreate(orderReq)
	if err != nil || txid == "" {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		//写入热钱包表，创建失败
		return "", err
	}
	orderHot.Status = int(status.BroadcastStatus)
	orderHot.TxId = txid
	//保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		//发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}

func (srv *AdaTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//无需实现
	return errors.New("implement me")
}

//params worker 指定机器
func (srv *AdaTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.AdaTxTpl, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      int64
		fee           int64
		coinSet       *entity.FcCoinSet //db币种配置
		tokencoinSet  *entity.FcCoinSet //db币种配置
		tpl           *transfer.AdaTxTpl
	)

	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常")
	}
	if ta.Eoskey != "" {
		tokencoinSet = global.CoinDecimal[ta.Eoskey]
		if tokencoinSet == nil {
			return nil, errors.New("读取 coinSet 设置异常")
		}
	}

	//查询找零地址
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("Ada 商户=[%d],查询Ada找零地址失败", ta.AppId)
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
	//Ada api需要金额整型
	toAmount = toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart()
	if ta.Eoskey != "" {
		toAmount = toAddrAmount.Shift(int32(tokencoinSet.Decimal)).IntPart()
	}
	if ta.Eostoken == "" && toAmount < 1000000 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额1.000000", ta.Id, ta.OutOrderid)
	}
	//手续费转换
	feeDecimal, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		log.Errorf("下单表订单id：%d,Ada 交易手续费转换异常:%s", ta.Id, err.Error())
		return nil, err
	}
	//Ada api需要金额整型
	fee = feeDecimal.Shift(int32(coinSet.Decimal)).IntPart()

	//todo Worker读取数据库
	orderReq := &transfer.AdaOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     srv.CoinName,
			Worker:       worker,
		},
		Amount: toAmount,
		Fee:    fee,
	}
	//填充参数
	if ta.Eoskey == "" {
		tpl, err = srv.setUtxoData(int64(ta.AppId), orderReq, toAddr, changeAddress)
		if err != nil {
			return nil, err
		}
	} else {
		tpl, err = srv.setTokenUtxoData(int64(ta.AppId), orderReq, toAddr, changeAddress, ta.Eostoken, ta.Eoskey)
		if err != nil {
			return nil, err
		}
	}
	return tpl, nil
}

//创建交易接口参数
func (srv *AdaTransferService) walletServerCreate(orderReq *transfer.AdaTxTpl) (string, error) {

	log.Infof("Ada 发送url：%s", conf.Cfg.HotServers[srv.CoinName].Url+"/v1/ada/transfer")
	log.Infof("Ada 发送结构：%v", String(orderReq))
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/v1/ada/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("Ada 发送返回：%s", string(data))
	result, err := transfer.DecodeTransferHotResp(data)
	if err != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,error:%s", orderReq.OrderId, err.Error())
	}
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderId)
	}
	if result.Code != 0 || result.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OrderId)
	}
	txid := result.Txid
	//冻结utxo
	for _, v := range orderReq.TxIns {
		dao.FcTransPushFreezeUtxo(v.FromTxid, int(v.FromIndex), v.FromAddr)
	}
	return txid, nil
}

//为orderReq的 OrderAddress组装交易内容
//toAddr 接收地址
//changeAddr 找零地址
//appid 商户ID
//orderReq walletsever交易结构
//fee 手续费
func (srv *AdaTransferService) setUtxoData(appid int64, orderReq *transfer.AdaOrderRequest, toAddr, changeAddr string) (*transfer.AdaTxTpl, error) {
	var (
		mchAmount       decimal.Decimal //商户余额
		fromAmountInt64 decimal.Decimal //发送总金额
		toAmountInt64   decimal.Decimal //接收总金额
		feeAmount       decimal.Decimal //手续费
		fee             int64
		feeAmountTmp    decimal.Decimal   //临时估算手续费
		feeTmp          int64             //临时估算手续费
		changeAmount    decimal.Decimal   //找零金额
		AdaCfg          *conf.HotServers  //币种数据服务
		coinSet         *entity.FcCoinSet //db币种配置
		//unspents     *transfer.BtcUnspents
		err      error
		txInTpl  = make([]transfer.AdaTxInTpl, 0)
		txOutTpl = make([]transfer.AdaTxOutTpl, 0)
		tpl      *transfer.AdaTxTpl
		unspents *transfer.AdaUnspents
	)
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return nil, err
	}

	if toAddr == "" {
		return nil, errors.New("结构缺少to地址")
	}
	if changeAddr == "" {
		return nil, errors.New("结构缺少change地址")
	}

	fee = orderReq.Fee

	toAmountInt64 = decimal.New(orderReq.Amount, 0)

	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常:" + srv.CoinName)
	}

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
	if err != nil {
		return nil, fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
	if mchAmount.Shift(int32(coinSet.Decimal)).LessThanOrEqual(toAmountInt64) {
		return nil, fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			srv.CoinName,
			mchAmount.String(),
			toAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
		)
	}
	AdaCfg = conf.Cfg.HotServers[srv.CoinName]
	if AdaCfg == nil {
		return nil, errors.New("配置文件缺少Ada" + srv.CoinName + " coinservers设置")
	}

	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常")
	}

	//出账的话优先使用前面50个大金额地址
	addrInfos, err := dao.FcAddressAmountFindTransfer(appid, srv.CoinName, 50, "desc")
	if err != nil {
		return nil, err
	}
	if len(addrInfos) == 0 {
		return nil, fmt.Errorf("订单：%s，暂无可用地址出账", orderReq.OuterOrderNo)
	}

	addrs := make([]string, 0)
	for _, v := range addrInfos {
		addrs = append(addrs, v.Address)
	}
	//查询utxo数量
	byteData, err := util.PostJson(AdaCfg.Url+"/api/v1/ada/unspents", addrs)
	if err != nil {
		return nil, fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.AdaUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return nil, fmt.Errorf("获取utxo解析json异常，:%s, data:%v", err.Error(), string(byteData))
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return nil, errors.New("Ada empty unspents")
	}

	if len(unspents.Data) < 50 {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("商户=[%d]，币种=[%s]，utxo数量：%d", appid, srv.CoinName, len(unspents.Data)))
	}

	if fee != 0 {
		if fee < 150000 || fee > 10000000 {
			//使用指定手续费
			return nil, errors.New("指定的手续费不在合理范围值[[0.150000-10.000000]")
		}
		feeTmp = fee
	} else {
		//先提前预估手续费
		feeTmp = 500000
	}

	feeAmountTmp = decimal.New(feeTmp, 0)

	//排序unspent，先进行降序，找出大额的数值
	var sortUtxo transfer.AdaUnspentDesc
	var sortUtxoTmp transfer.AdaUnspentAsc //临时使用，小金额排序
	sortUtxoTmp = append(sortUtxoTmp, unspents.Data...)
	srv.Sort(sortUtxoTmp)
	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoTmp {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DOGE_UTXO_LOCK, uv.Txid, uv.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		if len(uv.Tokens) > 1 { //只筛选ada的utxo
			continue
		}
		amInt64 := decimal.NewFromInt(uv.Tokens["ADA"])
		if amInt64.GreaterThanOrEqual(toAmountInt64.Add(feeAmountTmp)) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, amInt64.Shift(0-int32(coinSet.Decimal)).String(), uv.Address)
			sortUtxo = append(sortUtxo, uv)
			break
		}
	}
	if len(sortUtxo) == 0 {
		sortUtxo = append(sortUtxo, unspents.Data...)
	}
	sort.Sort(sortUtxo)

	//组装from
	for _, v := range sortUtxo {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DOGE_UTXO_LOCK, v.Txid, v.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		if len(v.Tokens) > 1 { //只筛选ada的utxo
			continue
		}
		famountInt64 := decimal.NewFromInt(v.Tokens["ADA"])
		oar := transfer.AdaTxInTpl{
			FromAddr:  v.Address,
			FromTxid:  v.Txid,
			FromIndex: uint32(v.Vout),
			Tokens:    v.Tokens,
		}
		fromAmountInt64 = fromAmountInt64.Add(famountInt64)
		txInTpl = append(txInTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.UCA_UTXO_LOCK_SECOND_TIME)

		if fromAmountInt64.GreaterThan(toAmountInt64.Add(feeAmountTmp)) {
			//满足出账
			break
		}
		//if len(utxoTpl) == conf.Cfg.UtxoScan.Num {
		//	//为了保证扫码稳定性 只使用15个utxo
		//	break
		//}

		if len(txInTpl) == 100 {
			//为了保证扫码稳定性 只使用15个utxo
			break
		}
	}
	if fromAmountInt64.LessThan(toAmountInt64) {
		return nil, fmt.Errorf("使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,出账金额：%s，预估手续费：%s",
			mchAmount.String(),
			fromAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			toAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			feeAmountTmp.Shift(0-int32(coinSet.Decimal)).String(),
		)
	}

	//实际使用手续费
	if fee == 0 {
		fee, err = srv.getFee(len(txInTpl), 2)
		if err != nil {
			return nil, err
		}
	}
	feeAmount = decimal.New(fee, 0)

	//组装to
	txOutTpl = append(txOutTpl, transfer.AdaTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: toAmountInt64.IntPart(),
		Token:    "",
	})

	//计算找零金额
	changeAmount = fromAmountInt64.Sub(toAmountInt64).Sub(feeAmount)
	if changeAmount.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("找零金额异常使用金额：%s,出账金额：%s，手续费：%s，找零：%s",
			fromAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			toAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			feeAmount.Shift(0-int32(coinSet.Decimal)).String(),
			changeAmount.Shift(0-int32(coinSet.Decimal)).String(),
		)
	}

	orderReq.Fee = fee
	tpl = &transfer.AdaTxTpl{
		MchId:    orderReq.MchName,
		OrderId:  orderReq.OuterOrderNo,
		CoinName: orderReq.CoinName,
		TxIns:    txInTpl,
		TxOuts:   txOutTpl,
		Change:   changeAddr,
	}
	return tpl, nil
}

func (srv *AdaTransferService) setTokenUtxoData(appid int64, orderReq *transfer.AdaOrderRequest, toAddr, changeAddr string, Eostoken, Eoskey string) (*transfer.AdaTxTpl, error) {
	var (
		mchAmount       decimal.Decimal //商户余额
		fromAmountInt64 decimal.Decimal //发送总金额
		toAmountInt64   decimal.Decimal //接收总金额
		feeAmount       decimal.Decimal //手续费
		fee             int64
		feeAmountTmp    decimal.Decimal //临时估算手续费
		feeTmp          int64           //临时估算手续费
		//changeAmount    decimal.Decimal   //找零金额
		AdaCfg  *conf.HotServers  //币种数据服务
		coinSet *entity.FcCoinSet //db币种配置
		//unspents     *transfer.BtcUnspents
		err      error
		txInTpl  = make([]transfer.AdaTxInTpl, 0)
		txOutTpl = make([]transfer.AdaTxOutTpl, 0)
		tpl      *transfer.AdaTxTpl
		unspents *transfer.AdaUnspents
	)
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return nil, err
	}

	if toAddr == "" {
		return nil, errors.New("结构缺少to地址")
	}
	if changeAddr == "" {
		return nil, errors.New("结构缺少change地址")
	}

	fee = orderReq.Fee

	toAmountInt64 = decimal.New(orderReq.Amount, 0)

	coinname := Eoskey
	//if Eoskey != "" {
	//	coinname = Eoskey
	//}

	coinSet = global.CoinDecimal[coinname]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常:" + coinname)
	}

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), coinname)
	if err != nil {
		return nil, fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
	if mchAmount.Shift(int32(coinSet.Decimal)).LessThan(toAmountInt64) {
		return nil, fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			Eoskey,
			mchAmount.String(),
			toAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
		)
	}
	AdaCfg = conf.Cfg.HotServers[srv.CoinName]
	if AdaCfg == nil {
		return nil, errors.New("配置文件缺少" + srv.CoinName + " coinservers设置")
	}

	//coinSet = global.CoinDecimal[srv.CoinName]
	//if coinSet == nil {
	//	return nil, errors.New("读取 coinSet 设置异常")
	//}

	//出账的话优先使用前面50个大金额地址
	addrInfos, err := dao.FcAddressAmountFindTransfer(appid, Eoskey, 50, "desc")
	if err != nil {
		return nil, err
	}
	if len(addrInfos) == 0 {
		return nil, fmt.Errorf("订单：%s，暂无可用地址出账", orderReq.OuterOrderNo)
	}

	addrs := make([]string, 0)
	for _, v := range addrInfos {
		addrs = append(addrs, v.Address)
	}
	//查询utxo数量
	byteData, err := util.PostJson(AdaCfg.Url+"/api/v1/ada/unspents", addrs)
	if err != nil {
		return nil, fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.AdaUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return nil, fmt.Errorf("获取utxo解析json异常，:%s, data:%v", err.Error(), string(byteData))
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return nil, errors.New("Ada empty unspents")
	}

	if len(unspents.Data) < 50 {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("商户=[%d]，币种=[%s]，utxo数量：%d", appid, srv.CoinName, len(unspents.Data)))
	}

	//if fee != 0 {
	//	if fee < 150000 || fee > 10000000 {
	//		//使用指定手续费
	//		return nil, errors.New("指定的手续费不在合理范围值[[500.00-100000.00]")
	//	}
	//	feeTmp = fee
	//} else {
	//	//先提前预估手续费
	//	feeTmp = 500000
	//}

	feeAmountTmp = decimal.New(feeTmp, 0)

	//排序unspent，先进行降序，找出大额的数值
	sortUtxo := transfer.AdaTokenUnspentDesc{
		Assertid: coinSet.Token,
	}
	sortUtxoTmp := transfer.AdaTokenUnspentAsc{
		Assertid: coinSet.Token,
	} //临时使用，小金额排序
	sortUtxoTmp.Utxos = append(sortUtxoTmp.Utxos, unspents.Data...)
	sort.Sort(sortUtxoTmp)
	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoTmp.Utxos {

		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DOGE_UTXO_LOCK, uv.Txid, uv.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		amInt64 := decimal.NewFromInt(uv.Tokens[sortUtxoTmp.Assertid])
		if amInt64.GreaterThanOrEqual(toAmountInt64) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, amInt64.Shift(0-int32(coinSet.Decimal)).String(), uv.Address)
			sortUtxo.Utxos = append(sortUtxo.Utxos, uv)
			break
		}
	}

	if len(sortUtxo.Utxos) == 0 {
		sortUtxo.Utxos = append(sortUtxo.Utxos, unspents.Data...)
	}
	sort.Sort(sortUtxo)
	var AdaAmountInt64 decimal.Decimal
	//组装from
	for _, v := range sortUtxo.Utxos {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DOGE_UTXO_LOCK, v.Txid, v.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		if v.Tokens[sortUtxo.Assertid] == 0 {
			continue
		}

		famountInt64 := decimal.NewFromInt(v.Tokens[sortUtxo.Assertid])

		oar := transfer.AdaTxInTpl{
			FromAddr:  v.Address,
			FromTxid:  v.Txid,
			FromIndex: uint32(v.Vout),
			Tokens:    v.Tokens,
			//FromAmount: famountInt64.IntPart(),
		}
		fromAmountInt64 = fromAmountInt64.Add(famountInt64)
		AdaAmountInt64 = AdaAmountInt64.Add(decimal.NewFromInt(v.Tokens["ADA"]))
		txInTpl = append(txInTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.UCA_UTXO_LOCK_SECOND_TIME)

		if fromAmountInt64.Cmp(toAmountInt64) >= 0 {
			//满足出账
			break
		}
		//if len(utxoTpl) == conf.Cfg.UtxoScan.Num {
		//	//为了保证扫码稳定性 只使用15个utxo
		//	break
		//}

		if len(txInTpl) == 100 {
			//为了保证扫码稳定性 只使用15个utxo
			break
		}
	}

	if fromAmountInt64.LessThan(toAmountInt64) {
		return nil, fmt.Errorf("使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,出账金额：%s，预估手续费：%s",
			mchAmount.String(),
			fromAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			toAmountInt64.Shift(0-int32(coinSet.Decimal)).String(),
			feeAmountTmp.Shift(0-int32(coinSet.Decimal)).String(),
		)
	}

	//实际使用手续费
	//if fee == 0 {
	fee, err = srv.getFee(len(txInTpl)+1, 2)
	if err != nil {
		return nil, err
	}

	fee += 2 * 1450000 //每个oututxo 都必须有ada
	//}
	feeAmount = decimal.New(fee, 0)

	//组装to
	txOutTpl = append(txOutTpl, transfer.AdaTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: toAmountInt64.IntPart(),
		Token:    Eostoken,
	})
	log.Info(feeAmount, AdaAmountInt64)
	//计算找零金额还需要的utxo
	if feeAmount.GreaterThan(AdaAmountInt64) {
		feeutxos, err := srv.getFeeAdaUtxo(appid, orderReq, feeAmount.Sub(AdaAmountInt64))
		if err != nil {
			return nil, err
		}
		txInTpl = append(txInTpl, feeutxos...)
	}

	orderReq.Fee = fee
	tpl = &transfer.AdaTxTpl{
		MchId:    orderReq.MchName,
		OrderId:  orderReq.OuterOrderNo,
		CoinName: orderReq.CoinName,
		TxIns:    txInTpl,
		TxOuts:   txOutTpl,
		Change:   changeAddr,
	}
	return tpl, nil
}

func (srv *AdaTransferService) getFeeAdaUtxo(appid int64, orderReq *transfer.AdaOrderRequest, neddFee decimal.Decimal) ([]transfer.AdaTxInTpl, error) {
	addrInfos, err := dao.FcAddressAmountFindTransfer(appid, srv.CoinName, 50, "desc")
	if err != nil {
		return nil, err
	}
	if len(addrInfos) == 0 {
		return nil, fmt.Errorf("订单：%s，暂无可用地址出账", orderReq.OuterOrderNo)
	}

	addrs := make([]string, 0)
	for _, v := range addrInfos {
		addrs = append(addrs, v.Address)
	}
	//查询utxo数量
	byteData, err := util.PostJson(conf.Cfg.HotServers[srv.CoinName].Url+"/api/v1/ada/unspents", addrs)
	if err != nil {
		return nil, fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents := new(transfer.AdaUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return nil, fmt.Errorf("获取utxo解析json异常，:%s, data:%v", err.Error(), string(byteData))
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return nil, errors.New("Ada empty unspents")
	}
	var unspentfee []transfer.AdaUtxo //最多需要两个额外的ada utxo够支付手续费
	for k, _ := range unspents.Data {
		if len(unspents.Data[k].Tokens) != 1 { //必须是纯粹的ada utxo
			continue
		}
		if len(unspentfee) == 0 {
			unspentfee = append(unspentfee, unspents.Data[k])
			continue
		} else if len(unspentfee) == 1 {
			unspentfee = append(unspentfee, unspents.Data[k])
		} else if len(unspentfee) == 2 {
			if unspentfee[1].Tokens["ADA"] > unspents.Data[k].Tokens["ADA"] {
				unspentfee[1] = unspents.Data[k]
			}
		}
		if unspentfee[0].Tokens["ADA"] > unspentfee[1].Tokens["ADA"] {
			unspentfee[0].Tokens["ADA"], unspentfee[1].Tokens["ADA"] = unspentfee[1].Tokens["ADA"], unspentfee[0].Tokens["ADA"]
		}
	}
	if len(unspentfee) == 0 {
		return nil, errors.New("ada 没找到手续费utxo")
	}
	if unspentfee[0].Tokens["ADA"] >= neddFee.IntPart() {
		unspentfee = unspentfee[0:1]
	} else if len(unspentfee) == 1 {
		return nil, errors.New("ada 1没找到足够手续费utxo")
	} else if len(unspentfee) == 2 && unspentfee[0].Tokens["ADA"]+unspentfee[1].Tokens["ADA"] < neddFee.IntPart() {
		return nil, errors.New("ada 2没找到足够手续费utxo")
	}
	var ret []transfer.AdaTxInTpl
	for k, _ := range unspentfee {
		tmptxin := transfer.AdaTxInTpl{
			FromAddr:  unspentfee[k].Address,
			FromTxid:  unspentfee[k].Txid,
			FromIndex: uint32(unspentfee[k].Vout),
			Tokens:    unspentfee[k].Tokens,
		}
		ret = append(ret, tmptxin)
	}
	return ret, nil
}

//手续费计算
func (srv *AdaTransferService) getFee(inNum, outNum int) (int64, error) {
	return int64(155381 + 4400*(inNum+outNum)), nil
}
func (srv *AdaTransferService) Sort(list transfer.AdaUnspentAsc) {
	sort.Sort(list)
	for k, _ := range list {
		if len(list[k].Tokens) > 1 {
			list = list[0:k]
			break
		}
	}
	return
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
