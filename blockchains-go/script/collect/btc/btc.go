package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/rediskey"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/script/collect/btc/base"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

var configFile = flag.String("c", "../../../conf/application.conf", "Configuration TOML File")
var coinConfigFile = flag.String("m", "./btc.toml", "Configuration TOML File")

var g_task_mark = 1

func main() {
	//初始化配置
	base.InitConf(*configFile, *coinConfigFile)

	second := base.CoinCfg.Second
	//spec := fmt.Sprintf("*/%d * * * * ?", second) //cron表达式，每?秒一次
	log.Infof("任务执行计划%d秒", second)
	//crontab := cron.New(cron.WithSeconds())
	//crontab.AddFunc(spec, MergeBtc)
	//crontab.Start()

	spec := fmt.Sprintf("@every %ds", second) //cron表达式，每?秒一次
	cronTab := cron.New()
	//定时任务发布,为了说明问题定义了5s,时间间隔很短
	cronTab.AddFunc(spec, func() { //cron的时间间隔规则这里不再陈述,有很多介绍
		//设置全局变量如果该定时任务没有执行完毕,不允许执行下一个定时任务
		if g_task_mark == 1 {
			g_task_mark = 2
			//数据批处理函数
			MergeBtc()
			g_task_mark = 1
		}
	})
	cronTab.Start()
	// 定时任务是另起协程执行的,这里使用 select 简单阻塞
	select {}

}

func MergeBtc() {
	log.Info("归集扫描")
	mchIds, err := base.FindMchIds(base.BtcCoinName)
	if err != nil {
		base.ErrDingBot.NotifyStr("归集BTC异常")
		return
	}
	if len(mchIds) == 0 {
		log.Infof("无可用归集商户")
		return
	}

	for _, v := range mchIds {
		//暂时只允许hoo进来
		if v.MchId != 1 {
			continue
		}

		orderReq, toAddr, toAmount, err := buildOrder(int64(v.MchId))
		if err != nil {
			log.Errorf("商户归集异常，ID：%d，error：%s", v.MchId, err.Error())
			continue
		}
		log.Infof("执行归集，商户ID:%d，toAddr：%s", v.MchId, toAddr)
		err = saveOrder(int64(v.MchId), orderReq, toAddr, toAmount)
		if err != nil {
			log.Errorf("商户归集保存异常，ID：%d，error：%s", v.MchId, err.Error())
			continue
		}
		//发送交易
		err = walletServerCreate(orderReq)
		if err != nil {
			log.Errorf("商户归集发送交易异常，ID：%d，error：%s", v.MchId, err.Error())
			continue
		}
	}

}

//由于归集的特殊性，接收地址只有一个，因此可以同时返回接收地址和接收金额
func buildOrder(appid int64) (orderReq *transfer.BtcOrderRequest, toAddr string, toAmount string, err error) {

	mchInfo, err := base.FindMch(appid)
	if err != nil {
		log.Errorf("无法查询商户信息 ID：%d", appid)
		return nil, "", "", err
	}
	log.Info("appid:", appid)
	toAddrs := make([]string, 0)
	if appid == 1 {
		toAddrs = []string{"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"}
	} else {
		//查询这个币种的归集地址
		toAddrs, err = base.FindBTCMergeAddrStrs(appid)
		if err != nil {
			return nil, "", "", fmt.Errorf("商户归集异常,无法查询归集地址，商户：%d,coinName:%s,err:%s", appid, base.BtcCoinName, err.Error())
		}
	}
	if len(toAddrs) == 0 {
		return nil, "", "", fmt.Errorf("商户归集异常,缺少归集地址，商户：%d,coinName:%s", appid, base.BtcCoinName)
	}
	//随机选一个
	//index := util.RandInt64(0, int64(len(toAddrs)))
	//toAddr = toAddrs[index]
	//暂时固定第一个吧，好观察
	toAddr = toAddrs[0]
	log.Info("toAddr:", toAddr)

	//随机选择可用机器
	workerId, err := getWorker(base.BtcCoinName)
	if err != nil {
		return nil, "", "", fmt.Errorf("商户归集异常,无法指定机器，商户：%d,coinName:%s,err:%s", appid, base.BtcCoinName, err.Error())
	}
	//填充参数
	orderReq = &transfer.BtcOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      1,
			ApplyCoinId:  4,
			OuterOrderNo: "COLLECT_" + util.GetUUID(),
			MchId:        appid,
			MchName:      mchInfo.Platform,
			CoinName:     base.BtcCoinName,
			Worker:       workerId,
		},
	}

	err = makeUtxoData(appid, orderReq, toAddr)
	if err != nil {
		return nil, "", "", err
	}
	toA := decimal.Zero
	for _, v := range orderReq.OrderAddress {
		if v.Dir == transfer.DirTypeTo {
			toA = toA.Add(decimal.New(v.Amount, 0))
		}
	}
	toAmount = toA.String()
	return orderReq, toAddr, "", nil
}

func saveOrder(appid int64, orderReq *transfer.BtcOrderRequest, toAddr, toAmount string) error {
	ta := &entity.FcTransfersApply{
		Username:   "robot",
		OrderId:    orderReq.OrderNo,
		Applicant:  orderReq.MchName,
		AppId:      int(appid),
		CallBack:   "robot",
		OutOrderid: orderReq.OuterOrderNo,
		CoinName:   orderReq.CoinName,
		Type:       "gj",
		Memo:       "",
		Eoskey:     "",
		Eostoken:   "",
		Fee:        strconv.FormatInt(orderReq.Fee, 10),
		Status:     int(entity.ApplyStatus_Merge),
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		ErrorNum:   100,
	}

	tacTo := &entity.FcTransfersApplyCoinAddress{
		ApplyCoinId: base.CoinSet.Id,
		Address:     toAddr,
		AddressFlag: "to",
		ToAmount:    toAddr,
	}

	//保存数据库
	applyId, err := base.SaveApplyOrder(ta, []*entity.FcTransfersApplyCoinAddress{tacTo})
	if err != nil {
		return err
	}
	orderReq.ApplyId = applyId
	return nil

}

//为orderReq的 OrderAddress组装交易内容
//toAddr 接收地址
//changeAddr 找零地址
//appid 商户ID
//orderReq walletsever交易结构
//fee 手续费
func makeUtxoData(appid int64, orderReq *transfer.BtcOrderRequest, toAddr string) error {
	log.Infof("扫描商户：%d", appid)
	var (
		fromAmount   decimal.Decimal //发送总金额
		toAmount     decimal.Decimal //接收总金额
		feeAmount    decimal.Decimal //手续费
		fee          int64
		minFeeAmount decimal.Decimal   //最小手续费
		maxFeeAmount decimal.Decimal   //最大手续费
		btcCfg       *conf.CoinServers //币种数据服务
		unspents     *transfer.BtcUnspents
		err          error
	)
	minFeeAmount = base.CoinCfg.MinFee.Shift(base.BtcDecimalBit) //变成整数
	maxFeeAmount = base.CoinCfg.MaxFee.Shift(base.BtcDecimalBit) //变成整数

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return err
	}

	utxoTpl := make([]*transfer.BtcOrderAddrRequest, 0) //utxo模板

	if toAddr == "" {
		return errors.New("结构缺少to地址")
	}

	btcCfg = base.Cfg.CoinServers[base.BtcCoinName]
	if btcCfg == nil {
		return errors.New("配置文件缺少btc coinservers设置")
	}
	//排序找金额15个地址
	if base.CoinSet == nil {
		return errors.New("读取 coinSet 设置异常")
	}
	//归集先使用前面15个金额地址
	if base.CoinCfg.AddrNumber == 0 {
		base.CoinCfg.AddrNumber = 15
	}
	addrInfos := make([]*entity.FcAddressAmount, 0)
	if base.CoinCfg.HasCold {
		addrInfos, err = base.FindBTCMergeAddr(appid, base.CoinCfg.AddrNumber)
	} else {
		addrInfos, err = base.FindMergeAddr(appid, base.CoinCfg.AddrNumber)
	}
	if err != nil {
		return err
	}
	if len(addrInfos) == 0 {
		return errors.New("暂无归集地址，忽略")
	}
	addrs := make([]string, 0)
	for _, v := range addrInfos {
		addrs = append(addrs, v.Address)
	}

	//查询utxo数量
	byteData, err := util.PostJson(btcCfg.Url+"/api/v1/btc/unspents", addrs)
	if err != nil {
		return fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	//log.Infof("uxto 返回内容：%s", string(byteData))
	unspents = new(transfer.BtcUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return errors.New("btc empty unspents")
	}

	minUtxoAmount := base.CoinCfg.MinUtxoAmount
	maxUtxoAmount := base.CoinCfg.MaxUtxoAmount
	if minUtxoAmount.IsZero() {
		minUtxoAmount = decimal.NewFromFloat(0.0001)
	}
	if maxUtxoAmount.IsZero() {
		maxUtxoAmount = decimal.NewFromFloat(1)
	}
	log.Infof("utxo数量：%d", len(unspents.Data))
	log.Infof("utxo限制最小金额：%s，最大金额：%s", minUtxoAmount.String(), maxUtxoAmount.String())

	//排序unspent，先进行降序，找出大额的数值
	var sortUtxo transfer.BtcUnspentDesc
	sortUtxo = append(sortUtxo, unspents.Data...)
	sort.Sort(sortUtxo)

	//组装from
	for _, v := range sortUtxo {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.BTC_UTXO_LOCK, v.Txid, v.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo 跳过
			continue
		}
		if v.Confirmations == 0 {
			//过滤不使用
			continue
		}

		if v.Amount < minUtxoAmount.Shift(base.BtcDecimalBit).IntPart() {
			//过滤不使用
			continue
		}
		if v.Amount > maxUtxoAmount.Shift(base.BtcDecimalBit).IntPart() {
			//过滤不使用
			continue
		}

		oar := &transfer.BtcOrderAddrRequest{
			Dir:     transfer.DirTypeFrom,
			Address: v.Address,
			Amount:  v.Amount,
			TxID:    v.Txid,
			Vout:    v.Vout,
		}
		fromAmount = fromAmount.Add(decimal.New(v.Amount, 0))
		utxoTpl = append(utxoTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.BTC_UTXO_LOCK_SECOND_TIME)

		if len(utxoTpl) == int(base.CoinCfg.UtxoNumber) {
			//为了保证扫码稳定性 只使用15个utxo
			break
		}
	}

	//实际使用手续费
	fee, err = getFee(len(utxoTpl), 1)
	if err != nil {
		return err
	}
	feeAmount = decimal.New(fee, 0)

	if feeAmount.GreaterThan(maxFeeAmount) {
		//不允许使用超过最大手续费限制
		feeAmount = maxFeeAmount
	}
	if feeAmount.LessThan(minFeeAmount) {
		feeAmount = minFeeAmount
	}

	toAmount = fromAmount.Sub(feeAmount)
	log.Infof(fmt.Sprintf("from金额：%s,手续费金额:%s，发送金额:%s",
		fromAmount.Shift(-1*base.BtcDecimalBit).String(),
		feeAmount.Shift(-1*base.BtcDecimalBit).String(),
		toAmount.Shift(-1*base.BtcDecimalBit).String(),
	))
	if toAmount.LessThan(base.CoinCfg.MinAmount.Shift(base.BtcDecimalBit)) {
		//如果归集金额 扣除手续费后小于minAmount,则放弃归集
		return fmt.Errorf("放弃归集，from金额：%s,手续费金额:%s，预计归集金额：%s,最小期望金额：%s",
			fromAmount.Shift(-1*base.BtcDecimalBit).String(),
			feeAmount.Shift(-1*base.BtcDecimalBit).String(),
			toAmount.Shift(-1*base.BtcDecimalBit).String(),
			base.CoinCfg.MinAmount.String(),
		)
	}
	//组装to
	utxoTpl = append(utxoTpl, &transfer.BtcOrderAddrRequest{
		Dir:     transfer.DirTypeTo,
		Address: toAddr,
		Amount:  toAmount.IntPart(),
	})

	orderReq.OrderAddress = utxoTpl
	orderReq.Fee = feeAmount.IntPart()
	orderReq.OrderNo, _ = util.GetOrderId(orderReq.OuterOrderNo, toAddr, toAmount.String())
	return nil

}

//手续费计算
func getFee(inNum, outNum int) (int64, error) {

	//默认费率
	rate := int64(40)
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

	respData, err := util.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
	if err != nil {
		log.Errorf("获取在线费率失败，将会使用默认费率：%d", rate)
	} else {
		result := &transfer.BtcGasResult{}
		result, err = transfer.DecodeBtcGasResult(respData)
		if err != nil {
			log.Errorf("解析在线费率，将会使用默认费率：%d", rate)
		} else {
			rate = result.HalfHourFee
		}

	}
	//使用最快优先级手续费
	fee := rate * byteNum
	feeD := decimal.New(fee, 0)
	if feeD.GreaterThan(decimal.NewFromFloat(0.01).Shift(8)) {
		return 0, fmt.Errorf("手续费过高：%s", feeD.String())
	}
	//限定最小值
	if fee < 1000 {
		fee = 1000
	}
	return fee, nil
}

//发送交易请求
func walletServerCreate(orderReq *transfer.BtcOrderRequest) error {
	data, err := util.PostJsonByAuth(base.Cfg.Walletserver.Url+"/btc/create", base.Cfg.Walletserver.User, base.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		base.ErrDingBot.NotifyStr(fmt.Sprintf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo))
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		base.ErrDingBot.NotifyStr(fmt.Sprintf("order表 请求下单接口返回值失败,服务器返回异常:%s，outOrderId：%s", string(data), orderReq.OuterOrderNo))
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常:%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil
}

func getWorker(coinName string) (string, error) {
	workerData, err := base.FindWorkers()
	if err != nil {
		return "", err
	}
	workers := make([]string, 0)
	for _, v := range workerData {
		arr := strings.Split(v.CoinName, ",")
		for _, av := range arr {
			if strings.ToLower(av) == strings.ToLower(coinName) {
				//如果有限定，使用限定机器
				workers = append(workers, v.WorkerCode)
			}
		}
	}
	if len(workers) == 0 {
		//如果找不到限定机器，使用随机机器
		for _, v := range workerData {
			workers = append(workers, v.WorkerCode)
		}
	}
	return workers[rand.Intn(len(workers))], nil

}

//CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o=btc-merge
