package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/rediskey"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"sort"
	"strconv"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type UsdtTransferService struct {
	CoinName string
}

func NewUsdtTransferService() service.TransferService {

	return &UsdtTransferService{
		CoinName: "usdt",
	}
}

func (srv *UsdtTransferService) VaildAddr(address string) error {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/usdt/validateaddress?address=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	usdtResp := transfer.DecodeUsdtAddressResult(data)
	if usdtResp != nil && usdtResp.Data != nil {
		if usdtResp.Data.Isvalid {
			return nil
		}
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	return err
}

func (srv *UsdtTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}

func (srv *UsdtTransferService) TransferCold(ta *entity.FcTransfersApply) error {
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

//======================私有方法==================

//params worker 指定机器
func (srv *UsdtTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.UsdtOrderRequest, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      int64
		fee           int64
		coinSet       *entity.FcCoinSet //db币种配置
	)

	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, errors.New("读取 coinSet 设置异常")
	}

	//查询找零地址，目前只有一个
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("usdt 商户=[%d],查询btc找零地址失败", ta.AppId)
	}
	//随机选择
	//randIndex := util.RandInt64(0, int64(len(changes)))
	//查询找零地址，目前只有一个
	changeAddress = changes[0]

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
	if toAddrAmount.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额0.00000001", ta.Id, ta.OutOrderid)
	}
	//usdt api需要金额整型
	toAmount = toAddrAmount.Shift(8).IntPart()

	//手续费转换
	feeDecimal, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		log.Errorf("下单表订单id：%d,usdt 交易手续费转换异常:%s", ta.Id, err.Error())
		return nil, err
	}
	//usdt api需要金额整型
	fee = feeDecimal.Shift(8).IntPart()

	//todo Worker读取数据库
	orderReq := &transfer.UsdtOrderRequest{
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
	err = srv.setUtxoData(int64(ta.AppId), orderReq, toAddr, changeAddress)
	if err != nil {
		return nil, err
	}
	return orderReq, nil
}

//创建交易接口参数
func (srv *UsdtTransferService) walletServerCreate(orderReq *transfer.UsdtOrderRequest) error {

	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/usdt/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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
func (srv *UsdtTransferService) setUtxoData(appid int64, orderReq *transfer.UsdtOrderRequest, toAddr, changeAddr string) error {
	var (
		mchAmount     decimal.Decimal   //商户余额
		fromUsdtInt64 decimal.Decimal   //发送总金额
		fromBtcInt64  decimal.Decimal   //发送总金额
		toUsdtInt64   decimal.Decimal   //接收总金额
		feeInt64      decimal.Decimal   //手续费
		feeTmp        int64             //临时估算手续费
		changeInt64   decimal.Decimal   //找零金额
		usdtCfg       *conf.CoinServers //币种数据服务
		coinSet       *entity.FcCoinSet //db币种配置
		unspents      *transfer.BtcUnspents
		err           error
		usdtBalance   *transfer.UsdtBalanceData
		toBtcInt64    = decimal.New(546, 0)
	)
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return err
	}

	utxoTpl := make([]*transfer.UsdtOrderAddrRequest, 0) //utxo模板

	if toAddr == "" {
		return errors.New("结构缺少to地址")
	}
	if changeAddr == "" {
		return errors.New("结构缺少change地址")
	}
	if orderReq.Amount < 1 {
		return fmt.Errorf("发送金额不能小于0.00000001,目前金额：%s", decimal.New(orderReq.Amount, -8).String())
	}

	toUsdtInt64 = decimal.New(orderReq.Amount, 0)

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
	if err != nil {
		return fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
	if mchAmount.Shift(8).LessThan(toUsdtInt64) {
		return fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			srv.CoinName,
			mchAmount.String(),
			toUsdtInt64.Shift(-8).String(),
		)
	}
	usdtCfg = conf.Cfg.CoinServers[srv.CoinName]
	if usdtCfg == nil {
		return errors.New("配置文件缺少usdt coinservers设置")
	}
	//排序找金额15个地址
	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return errors.New("读取 coinSet 设置异常")
	}
	//查询满足出账的冷地址
	coldddr, err := dao.FcAddressAmountFindTransferToAccount(srv.CoinName, appid)
	if err != nil {
		log.Infof("币种=[%s]，查询出账地址失败，err=[%s]", srv.CoinName, err.Error())
	}
	usdtBalance, err = srv.getRealBalance(coldddr.Address)
	if err != nil {
		//查询线上数据异常
		return fmt.Errorf("地址：%s,查询线上数据异常：%s", coldddr.Address, err.Error())
	}
	orderReq.FromAddress = coldddr.Address

	if usdtBalance.RealBalanceFloat.Shift(8).LessThan(toUsdtInt64) {
		err = fmt.Errorf("订单：%s，usdt出账 ，接收地址：%s 需要接收金额：%s, 冷地址：%s,实际金额：%s,冻结金额：%s,接收中金额：%s，节点金额：%s",
			orderReq.OuterOrderNo, toAddr, toUsdtInt64.Shift(-8).String(), coldddr.Address, usdtBalance.RealBalanceFloat.String(),
			usdtBalance.LockFloat.String(), usdtBalance.PendingBalanceFloat.String(), usdtBalance.BalanceFloat.String())

		//钉钉通知
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
		return err
	}

	fromUsdtInt64, _ = decimal.NewFromString(coldddr.Amount)
	if fromUsdtInt64.Shift(8).LessThan(toUsdtInt64) {
		err = fmt.Errorf("订单=[%s]，出账金额=[%s]，冷地址金额=[%s]",
			orderReq.OuterOrderNo,
			toUsdtInt64.Shift(-8).String(),
			fromUsdtInt64.Shift(-8).String())
		return err
	}

	addrs := []string{coldddr.Address}
	//查询utxo数量
	byteData, err := util.PostJson(usdtCfg.Url+"/api/v1/btc/unspents", addrs)
	if err != nil {
		return fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.BtcUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return errors.New("usdt empty unspents")
	}
	//utxo告警
	srv.countUtxoToDing(unspents.Data)

	if orderReq.Fee != 0 {
		if orderReq.Fee < 5000 || orderReq.Fee > 1000000 {
			//使用指定手续费
			return errors.New("指定的手续费不在合理范围值[[0.00005000-0.1]")
		}
		feeInt64 = decimal.New(orderReq.Fee, 0)
	} else {
		//一般归集是两个utxo，出账是一个utxo
		feeTmp, err = srv.getFee(2, 3)
		if err != nil {
			return err
		}
		feeInt64 = decimal.New(feeTmp, 0)
	}

	//如果找不到合适的utxo 排序unspent，先进行降序，找出大额的数值
	var sortUtxo transfer.BtcUnspentDesc

	var sortUtxoTmp transfer.BtcUnspentAsc //临时使用，小金额排序
	sortUtxoTmp = append(sortUtxoTmp, unspents.Data...)
	sort.Sort(sortUtxoTmp)
	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoTmp {
		if uv.Confirmations == 0 {
			continue
		}
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.USDT_UTXO_LOCK, uv.Txid, uv.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		am := decimal.New(uv.Amount, 0)
		if am.GreaterThanOrEqual(toBtcInt64.Add(feeInt64)) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, am.String(), uv.Address)
			sortUtxo = append(sortUtxo, uv)
			break
		}
	}

	//如果没有合适的，重新计算
	if len(sortUtxo) == 0 {
		sortUtxo = append(sortUtxo, unspents.Data...)
		if feeInt64.IsZero() {
			//手续费模拟15个
			feeTmp, err = srv.getFee(conf.Cfg.UtxoScan.Num/3, 3)
			if err != nil {
				return err
			}
			feeInt64 = decimal.New(feeTmp, 0)
		}

	}
	sort.Sort(sortUtxo)

	//组装from
	for _, v := range sortUtxo {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.USDT_UTXO_LOCK, v.Txid, v.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		if v.Confirmations == 0 {
			//暂时过滤
			continue
		}
		oar := &transfer.UsdtOrderAddrRequest{
			Dir:          transfer.DirTypeFrom,
			Address:      v.Address,
			Amount:       v.Amount,
			TxID:         v.Txid,
			Vout:         v.Vout,
			ScriptPubKey: v.ScriptPubKey,
		}
		fromBtcInt64 = fromBtcInt64.Add(decimal.New(v.Amount, 0))
		utxoTpl = append(utxoTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.USDT_UTXO_LOCK_SECOND_TIME)

		if fromBtcInt64.GreaterThan(toBtcInt64.Add(feeInt64)) {
			//满足出账
			break
		}
		if len(utxoTpl) == conf.Cfg.UtxoScan.Num {
			//为了保证扫码稳定性 只使用15个utxo
			break
		}
	}
	if fromBtcInt64.LessThan(toBtcInt64.Add(feeInt64)) {
		return fmt.Errorf("usdt 使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,需要发送BTC金额：%s，预估限定数量utxo手续费：%s",
			mchAmount.String(),
			fromBtcInt64.Shift(-8).String(),
			toBtcInt64.Shift(-8).String(),
			feeInt64.Shift(-8).String(),
		)
	}

	if len(utxoTpl) == 0 {
		return fmt.Errorf("地址：%s,暂无可用UTXO", coldddr.Address)
	}

	//第一个附加usdt金额
	utxoTpl[0].TokenAmount = toUsdtInt64.IntPart()

	//实际使用手续费
	if len(utxoTpl) < conf.Cfg.UtxoScan.Num {
		feeTmp, err = srv.getFee(len(utxoTpl), 3)
		if err != nil {
			return err
		}
		if feeInt64.GreaterThan(decimal.New(feeTmp, 0)) {
			feeInt64 = decimal.New(feeTmp, 0)
		}
	}

	//组装to
	utxoTpl = append(utxoTpl, &transfer.UsdtOrderAddrRequest{
		Dir:         transfer.DirTypeTo,
		Address:     toAddr,
		Amount:      toBtcInt64.IntPart(),
		TokenAmount: toUsdtInt64.IntPart(),
	})

	//计算找零金额
	changeInt64 = fromBtcInt64.Sub(toBtcInt64).Sub(feeInt64)
	if changeInt64.LessThan(decimal.Zero) {
		return fmt.Errorf("usdt utxo找零金额异常使用金额：%s,出账BTC金额：%s，手续费：%s，找零：%s",
			fromBtcInt64.Shift(-8).String(),
			toBtcInt64.Shift(-8).String(),
			feeInt64.Shift(-8).String(),
			changeInt64.Shift(-8).String(),
		)
	}
	if changeInt64.LessThanOrEqual(decimal.New(546, 0)) {
		//如果找零小于546，那么附加在手续费上
		feeInt64 = feeInt64.Add(changeInt64)
	} else {
		//组装找零
		utxoTpl = append(utxoTpl, &transfer.UsdtOrderAddrRequest{
			Dir:     transfer.DirTypeChange,
			Address: changeAddr,
			Amount:  changeInt64.IntPart(),
		})
	}
	orderReq.OrderAddress = utxoTpl
	orderReq.Fee = feeInt64.IntPart()
	return nil

}

//手续费计算
func (srv *UsdtTransferService) getFee(inNum, outNum int) (int64, error) {

	var (
		rate int64
	)

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return 0, err
	}

	//默认费率
	if inNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error InNum"))
	}
	if outNum <= 0 {
		return 0, errors.New(fmt.Sprintf("Error OutNum"))
	}
	//近似值字节数
	//byteNum := int64(inNum*148 + 34*outNum + 10)
	//提高输出字节，加速出块
	byteNum := int64((inNum)*148 + 34*outNum + 10)

	if has, _ := redisHelper.Exists(rediskey.BTC_RATE); has {
		rateStr, _ := redisHelper.Get(rediskey.BTC_RATE)
		rate, _ = strconv.ParseInt(rateStr, 10, 64)
	} else {
		respData, err := util.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
		if err != nil {
			log.Errorf("USDT获取在线费率失败，将会使用默认费率：%d", rate)
		} else {
			result := &transfer.UsdtGasResult{}
			result, err = transfer.DecodeUsdtGasResult(respData)
			if err != nil {
				log.Errorf("USDT解析在线费率，将会使用默认费率：%d", rate)
			} else {
				rate = result.FastestFee
				redisHelper.Set(rediskey.BTC_RATE, rate)
				redisHelper.Expire(rediskey.BTC_RATE, 600) //10分钟过期
			}
		}

	}
	if rate == 0 {
		rate = 50
	}
	fee := rate * byteNum
	//限定最小值
	if fee < 50000 {
		fee = 50000
	}
	//限制最大
	if fee > 500000 {
		fee = 500000
	}
	return fee, nil
}

//钉钉告警
func (srv *UsdtTransferService) countUtxoToDing(utxo []transfer.BtcUtxo) {
	var (
		countBof  int //低于1000计数
		countHalf int //大于1000 小于 10000计数
		countEof  int //大于10000计数

		bof = decimal.New(1000, 0)
		eof = decimal.New(10000, 0)

		bofWarnNum = 200
		maxWarnNum = 30
	)
	for _, v := range utxo {
		if v.Amount < bof.IntPart() {
			countBof = countBof + 1
			continue
		} else if v.Amount > bof.IntPart() && v.Amount < eof.IntPart() {
			countHalf = countHalf + 1
			continue
		} else {
			countEof = countEof + 1
			continue
		}
	}
	if (countHalf + countEof) < maxWarnNum {

		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf(
			"usdt utxo预警,及时充值打散,金额小于[%s],utxo数量=[%d],其他信息,范围=[%s-%s]的数量为[%d],高于=[%s]的数量为[%d]",
			bof.Shift(-8).String(),
			countBof,
			bof.Shift(-8).String(),
			eof.Shift(-8).String(),
			countHalf,
			eof.Shift(-8).String(),
			countEof))
		return
	}

	if countBof >= bofWarnNum {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf(
			"usdt utxo预警,建议合并零散,金额小于[%s],utxo数量=[%d],其他信息,范围=[%s-%s]的数量为[%d],高于=[%s]的数量为[%d]",
			bof.Shift(8).String(),
			countBof,
			bof.Shift(8).String(),
			eof.Shift(8).String(),
			countHalf,
			eof.Shift(8).String(),
			countEof))
		return
	}

}

func (srv *UsdtTransferService) getRealBalance(address string) (*transfer.UsdtBalanceData, error) {
	usdtCfg := conf.Cfg.CoinServers[srv.CoinName]
	byteData, err := util.Get(fmt.Sprintf(usdtCfg.Url+"/api/v1/usdt/usdtbalance?address=%s", address))
	if err != nil {
		return nil, err
	}
	result, err := transfer.DecodeUsdtBalanceResp(byteData)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 || result.Data == nil {
		return nil, fmt.Errorf("获取线上余额失败，内容：%s", string(byteData))
	}
	return result.Data, nil
}
