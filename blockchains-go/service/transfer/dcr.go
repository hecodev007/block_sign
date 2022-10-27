package transfer

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/txscript"
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
	"sort"
	"strings"
	"time"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type DcrTransferService struct {
	CoinName string
}

func NewDcrTransferService() service.TransferService {
	return &DcrTransferService{
		CoinName: "dcr",
	}
}

func (srv *DcrTransferService) VaildAddr(address string) error {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/dcr/validateaddress?address=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	btcResp := transfer.DecodeDcrAddressResult(data)
	if btcResp != nil && btcResp.Data != nil {
		if btcResp.Data.Isvalid {
			return nil
		}
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	return err
}

func (srv *DcrTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		txHex string
		mch   *entity.FcMch
	)
	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, vins, err := srv.getEstimateTpl(ta, workerId)
	if err != nil {
		return "", err
	}
	txHex, err = srv.walletServerCreateHot(orderReq)
	//发送交易
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(txHex, "0x") {
		txHex = strings.TrimPrefix(txHex, "0x")
	}
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	//查询Dcr的coinSet
	coinSet := global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	//查询出账地址和金额
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
	//写入orderHot表
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
		Amount:       toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:     toAddrAmount.Shift(int32(coinSet.Decimal)).String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	txid, err = srv.sendRawTransaction(orderReq.MchName, txHex)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
	}
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	//保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		//err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	//发送交易成功，冻结utxo
	freezeUtxo(vins)
	return txid, nil
}

func (srv *DcrTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, _, err := srv.getEstimateTpl(ta, workerId)
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

//创建交易接口参数
func (srv *DcrTransferService) walletServerCreate(orderReq *transfer.DcrOrderRequest) error {

	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"dcr/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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

//创建交易接口参数
func (srv *DcrTransferService) walletServerCreateHot(orderReq *transfer.DcrOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[srv.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", srv.CoinName)
	}
	data, err := util.PostJsonByAuth(cfg.Url+"/v1/dcr/sign", cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return result.Data.(string), nil
}

//params worker 指定机器
func (srv *DcrTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.DcrOrderRequest, []transfer.DcrTxInTpl, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      int64
		fee           int64
		coinSet       *entity.FcCoinSet //db币种配置
	)
	coinSet = global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, nil, errors.New("读取 coinSet 设置异常")
	}

	//查询找零地址
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, nil, err
	}
	if len(changes) == 0 {
		return nil, nil, fmt.Errorf("dcr 商户=[%d],查询dcr找零地址失败", ta.AppId)
	}
	//随机选择
	randIndex := util.RandInt64(0, int64(len(changes)))
	changeAddress = changes[randIndex]

	//查询出账地址和金额
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return nil, nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	if toAddrAmount.LessThan(decimal.NewFromFloat(0.00000546)) {
		return nil, nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额0.00000546", ta.Id, ta.OutOrderid)
	}
	toAmount = toAddrAmount.Shift(int32(coinSet.Decimal)).IntPart()

	//手续费转换
	feeDecimal, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		log.Errorf("下单表订单id：%d,Dcr 交易手续费转换异常:%s", ta.Id, err.Error())
		return nil, nil, err
	}
	//
	fee = feeDecimal.Shift(int32(coinSet.Decimal)).IntPart()
	orderReq := &transfer.DcrOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     srv.CoinName,
			Worker:       worker,
		},
	}
	signData, vins, errBt := srv.buildDcrTx(int64(ta.AppId), toAddr, changeAddress, toAmount, fee, orderReq)
	if errBt != nil {
		return nil, nil, fmt.Errorf("build tx error: %v", errBt)
	}
	orderReq.Data = signData
	return orderReq, vins, nil
}

func (srv *DcrTransferService) buildDcrTx(appid int64, toAddr, changeAddr string, toAmount, fee int64, orderReq *transfer.DcrOrderRequest) (*transfer.DcrSignReq, []transfer.DcrTxInTpl, error) {

	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		return nil, nil, err
	}
	if toAddr == "" {
		return nil, nil, errors.New("结构缺少to地址")
	}
	if changeAddr == "" {
		return nil, nil, errors.New("结构缺少change地址")
	}
	if toAmount < 546 {
		return nil, nil, fmt.Errorf("发送金额不能小于0.00000546,目前金额：%s", decimal.New(toAmount, -8).String())
	}
	toAmt := decimal.New(toAmount, 0)

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
	if err != nil {
		return nil, nil, fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ := decimal.NewFromString(mchAmountResult.Amount)
	coinSet := global.CoinDecimal[srv.CoinName]
	if coinSet == nil {
		return nil, nil, errors.New("读取 coinSet 设置异常")
	}
	if mchAmount.Shift(int32(coinSet.Decimal)).LessThanOrEqual(toAmt) {
		return nil, nil, fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			srv.CoinName,
			mchAmount.String(),
			toAmt.Shift(-int32(coinSet.Decimal)).String(),
		)
	}
	//由于dcr币种无法使用unspent接口，只能从数据库维护
	tpush, err := dao.FcTransPushFindVaildUtxo(int(appid), srv.CoinName)
	if err != nil {
		return nil, nil, fmt.Errorf("find dcr utxo error:%s", err.Error())
	}
	if len(tpush) == 0 {
		return nil, nil, errors.New("dcr:No utxo available ")
	}
	var feeTmp int64 //临时估算手续费
	if fee != 0 {
		if fee < 1000 || fee > 10000000 {
			//使用指定手续费
			return nil, nil, errors.New("指定的手续费不在合理范围值[[0.00000546-0.1]")
		}
		feeTmp = fee
	} else {
		//先提前预估手续费
		feeTmp = 100000
	}
	//排序unspent，先进行降序，找出大额的数值
	var sortUtxo entity.DBUnspentDesc
	feeAmountTmp := decimal.New(feeTmp, 0)
	var sortUtxoTmp entity.DBUnspentDesc //临时使用，小金额排序
	sortUtxoTmp = append(sortUtxoTmp, tpush...)
	sort.Sort(sortUtxoTmp)

	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoTmp {
		if uv.Confirmations == 0 {
			continue
		}
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DCR_UTXO_LOCK, uv.TransactionId, uv.TrxN)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		amFloat, _ := decimal.NewFromString(uv.Amount)
		if amFloat.Shift(int32(coinSet.Decimal)).GreaterThanOrEqual(toAmt.Add(feeAmountTmp)) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, amFloat.String(), uv.Address)
			sortUtxo = append(sortUtxo, uv)
			break
		}
	}
	if len(sortUtxo) == 0 {
		sortUtxo = append(sortUtxo, tpush...)
	}
	sort.Sort(sortUtxo)

	var fromAmount decimal.Decimal //发送总金额
	var txInTpl = make([]transfer.DcrTxInTpl, 0)
	var txOutTpl = make([]transfer.DcrTxOutTpl, 0)
	//组装from
	for _, v := range sortUtxo {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DCR_UTXO_LOCK, v.TransactionId, v.TrxN)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		if v.Confirmations == 0 {
			//暂时过滤
			continue
		}
		famount, _ := decimal.NewFromString(v.Amount)
		from_amount, _ := famount.Float64()
		oar := transfer.DcrTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.TransactionId,
			FromIndex:  uint32(v.TrxN),
			FromAmount: from_amount,
		}
		fromAmount = fromAmount.Add(famount.Shift(int32(coinSet.Decimal)))
		txInTpl = append(txInTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.DCR_UTXO_LOCK_SECOND_TIME)

		if fromAmount.GreaterThan(toAmt.Add(feeAmountTmp)) {
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
	if fromAmount.LessThan(toAmt) {
		return nil, nil, fmt.Errorf("使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,出账金额：%s，预估手续费：%s",
			mchAmount.String(),
			fromAmount.Shift(-int32(coinSet.Decimal)).String(),
			toAmt.Shift(-int32(coinSet.Decimal)).String(),
			feeAmountTmp.Shift(-int32(coinSet.Decimal)).String(),
		)
	}

	//实际使用手续费
	if fee == 0 {
		fee, err = srv.getFee(len(txInTpl), 2)
		if err != nil {
			return nil, nil, err
		}
	}
	feeAmount := decimal.New(fee, 0)

	//组装to
	to_amount, _ := toAmt.Shift(0 - int32(coinSet.Decimal)).Float64()
	txOutTpl = append(txOutTpl, transfer.DcrTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: to_amount,
	})

	//计算找零金额
	changeAmount := fromAmount.Sub(toAmt).Sub(feeAmount)
	if changeAmount.LessThan(decimal.Zero) {
		return nil, nil, fmt.Errorf("找零金额异常使用金额：%s,出账金额：%s，手续费：%s，找零：%s",
			fromAmount.Shift(-8).String(),
			toAmt.Shift(-int32(coinSet.Decimal)).String(),
			feeAmount.Shift(-8).String(),
			changeAmount.Shift(-8).String(),
		)
	}
	if changeAmount.LessThanOrEqual(decimal.New(10000, 0)) {
		//如果找零小于10000，那么附加在手续费上
		feeAmount = feeAmount.Add(changeAmount)
	} else {
		//组装找零
		change_amount, _ := changeAmount.Shift(0 - int32(coinSet.Decimal)).Float64()
		txOutTpl = append(txOutTpl, transfer.DcrTxOutTpl{
			ToAddr:   changeAddr,
			ToAmount: change_amount,
		})
	}
	var ctReq transfer.DcrCreateTxReq
	ctReq.Vin = txInTpl
	ctReq.Vout = txOutTpl

	//发送请求去构建交易
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/dcr/create"
	data, errCT := util.PostJson(url, ctReq)
	if errCT != nil {
		return nil, nil, fmt.Errorf("create raw tx error: %v", errCT)
	}
	ct, errTx := transfer.DecodeCreateTxResp(data)
	if errTx != nil {
		return nil, nil, fmt.Errorf("decode createrawtransaction error: %v", errTx)
	}
	if ct.Code == 0 && ct.Message == "ok" && ct.Data != nil {
		var sigData transfer.DcrSignReq
		rawTx := ct.Data.(string)
		var addresses []string
		var inputs []*transfer.RawTxInput
		for _, vin := range txInTpl {
			pkScript, err := parseToScript(vin.FromAddr)
			if err != nil {
				return nil, nil, fmt.Errorf("parse %s to pkScript error: %v", vin.FromAddr, err)
			}
			addresses = append(addresses, vin.FromAddr)
			inputs = append(inputs, &transfer.RawTxInput{
				Txid:         vin.FromTxid,
				Vout:         vin.FromIndex,
				Tree:         0,
				ScriptPubKey: pkScript,
				RedeemScript: "",
			})
		}
		sigData.RawTx = rawTx
		sigData.Addresses = addresses
		sigData.Inputs = inputs
		return &sigData, txInTpl, nil
	}
	return nil, nil, fmt.Errorf("create raw transaction error: %v", ct.Data)
}

//手续费计算
func (srv *DcrTransferService) getFee(inNum, outNum int) (int64, error) {

	var (
		rate int64 = 100
	)
	//redisHelper, err := util.AllocRedisClient()
	//if err != nil {
	//	return 0, err
	//}

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
	byteNum := int64((inNum)*148 + 34*outNum + 10) //相差有点悬殊

	//if has, _ := redisHelper.Exists(rediskey.UCA_RATE); has {
	//	rateStr, _ := redisHelper.Get(rediskey.UCA_RATE)
	//	rate, _ = strconv.ParseInt(rateStr, 10, 64)
	//	log.Infof("Uca 读取缓存的rate:%d", rate)
	//} else {
	//	//respData, err := util.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
	//	//if err != nil {
	//	//	log.Errorf("Uca获取在线费率失败，将会使用默认费率：%d", rate)
	//	//} else {
	//	//	result := &transfer.UsdtGasResult{}
	//	//	result, err = transfer.DecodeUsdtGasResult(respData)
	//	//	if err != nil {
	//	//		log.Errorf("Uca解析在线费率，将会使用默认费率：%d", rate)
	//	//	} else {
	//	//		rate = result.HalfHourFee
	//	//		redisHelper.Set(rediskey.UCA_RATE, rate)
	//	//		redisHelper.Expire(rediskey.UCA_RATE, 600) //10分钟过期
	//	//	}
	//	//}
	//
	//}
	if rate == 0 {
		rate = 100
	}
	fee := rate * byteNum
	//限定最小值
	if fee < 10000 {
		fee = 10000
	}
	//限制最大
	if fee > 10000000 {
		fee = 10000000
	}
	return fee, nil
}

func parseToScript(address string) (string, error) {
	addr, err := dcrutil.DecodeAddress(address)
	if err != nil {
		return "", err
	}
	pks, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(pks), nil
}

func (srv *DcrTransferService) sendRawTransaction(mchId, txHex string) (string, error) {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/dcr/send"
	params := map[string]string{
		"hex":   txHex,
		"mchId": mchId,
	}
	data, err := util.PostJson(url, params)
	if err != nil {
		return "", err
	}
	//	解析data
	if len(data) == 0 {
		return "", errors.New("send tx response data is null")
	}
	var res map[string]interface{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}
	ct, errTx := transfer.DecodeCreateTxResp(data)
	if errTx != nil {
		return "", fmt.Errorf("decode sendrawtransaction error: %v", errTx)
	}
	if ct.Code == 0 && ct.Message == "ok" && ct.Data != nil {
		rawTx := ct.Data.(map[string]interface{})
		if rawTx["txid"] == nil {
			return "", errors.New("send tx response txid is null")
		} else {
			return rawTx["txid"].(string), nil
		}
	} else {
		return "", fmt.Errorf("send tx error,Err=[%s]", string(data))
	}
}

func unFreezeUtxo(vin []transfer.DcrTxInTpl) {
	if len(vin) == 0 {
		return
	}
	for _, v := range vin {
		err := dao.FcTransPushUnFreezeUtxo("dcr", v.FromTxid, int(v.FromIndex), v.FromAddr)
		if err != nil {
			log.Errorf("Dcr 解冻utxo失败,%+v", v)
		}
	}
}

func freezeUtxo(vin []transfer.DcrTxInTpl) {
	if len(vin) == 0 {
		return
	}
	for _, v := range vin {
		err := dao.FcTransPushFreezeUtxo(v.FromTxid, int(v.FromIndex), v.FromAddr)
		if err != nil {
			log.Errorf("Dcr 冻结utxo失败,%+v", v)
		}
	}
}
