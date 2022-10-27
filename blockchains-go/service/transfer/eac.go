package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
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

type EacTransferService struct {
	CoinName string
}

func NewEacTransferService() service.TransferService {
	return &EacTransferService{
		CoinName: "eac",
	}
}

var EacNetParams = new(chaincfg.Params)

func init() {
	EacNetParams.PubKeyHashAddrID = 0x5d
	EacNetParams.ScriptHashAddrID = 0x21
	EacNetParams.PrivateKeyID = 0xdd
	EacNetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xAD, 0xE4}
	EacNetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xB2, 0x1E}
}

func (srv *EacTransferService) VaildAddr(address string) error {
	if strings.HasPrefix(address, "e") {
		_, err := btcutil.DecodeAddress(address, EacNetParams)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
}

func (srv *EacTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {

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

func (srv *EacTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//无需实现
	return errors.New("implement me")
}

//params worker 指定机器
func (srv *EacTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.EacTxTpl, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      int64
		fee           int64
		coinSet       *entity.FcCoinSet //db币种配置
		tpl           *transfer.EacTxTpl
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
		return nil, fmt.Errorf("Eac 商户=[%d],查询Eac找零地址失败", ta.AppId)
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
	if toAddrAmount.LessThan(decimal.NewFromFloat(0.00000001)) {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额0.00000001", ta.Id, ta.OutOrderid)
	}
	//Eac api需要金额整型
	toAmount = toAddrAmount.Shift(8).IntPart()

	//手续费转换
	feeDecimal, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		log.Errorf("下单表订单id：%d,Eac 交易手续费转换异常:%s", ta.Id, err.Error())
		return nil, err
	}
	//Eac api需要金额整型
	fee = feeDecimal.Shift(8).IntPart()

	//todo Worker读取数据库
	orderReq := &transfer.EacOrderRequest{
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
	tpl, err = srv.setUtxoData(int64(ta.AppId), orderReq, toAddr, changeAddress)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

//创建交易接口参数
func (srv *EacTransferService) walletServerCreate(orderReq *transfer.EacTxTpl) (string, error) {

	log.Infof("eac 发送url：%s", conf.Cfg.HotServers[srv.CoinName].Url+"/v1/"+strings.ToLower(srv.CoinName)+"/transfer")
	log.Infof("eac 发送结构：%+v", orderReq)
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/v1/"+strings.ToLower(srv.CoinName)+"/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("eac 发送返回：%s", string(data))
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
func (srv *EacTransferService) setUtxoData(appid int64, orderReq *transfer.EacOrderRequest, toAddr, changeAddr string) (*transfer.EacTxTpl, error) {
	var (
		mchAmount       decimal.Decimal //商户余额
		fromAmountInt64 decimal.Decimal //发送总金额
		toAmountInt64   decimal.Decimal //接收总金额
		feeAmount       decimal.Decimal //手续费
		fee             int64
		feeAmountTmp    decimal.Decimal   //临时估算手续费
		feeTmp          int64             //临时估算手续费
		changeAmount    decimal.Decimal   //找零金额
		EacCfg          *conf.CoinServers //币种数据服务
		coinSet         *entity.FcCoinSet //db币种配置
		//unspents     *transfer.EacUnspents
		err      error
		txInTpl  = make([]transfer.EacTxInTpl, 0)
		txOutTpl = make([]transfer.EacTxOutTpl, 0)
		tpl      *transfer.EacTxTpl
		unspents *transfer.EacUnspents
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
	if orderReq.Amount < 546 {
		return nil, fmt.Errorf("发送金额不能小于0.00000546,目前金额：%s", decimal.New(orderReq.Amount, -8).String())
	}
	fee = orderReq.Fee

	toAmountInt64 = decimal.New(orderReq.Amount, 0)

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
	if err != nil {
		return nil, fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
	if mchAmount.Shift(8).LessThanOrEqual(toAmountInt64) {
		return nil, fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			srv.CoinName,
			mchAmount.String(),
			toAmountInt64.Shift(-8).String(),
		)
	}
	EacCfg = conf.Cfg.CoinServers[srv.CoinName]
	if EacCfg == nil {
		return nil, errors.New("配置文件缺少Eac coinservers设置")
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
	log.Infof("发送地址：%s", EacCfg.Url+"/v1/eac/unspents")
	byteData, err := util.PostJson(EacCfg.Url+"/v1/eac/unspents", addrs)
	if err != nil {
		return nil, fmt.Errorf("获取utxo异常，err:%s", err.Error())
	}
	unspents = new(transfer.EacUnspents)
	err = json.Unmarshal(byteData, unspents)
	if err != nil {
		return nil, fmt.Errorf("获取utxo解析json异常，:%s", err.Error())
	}
	if unspents.Code != 0 {
		fmt.Errorf("获取utxo异常，err:%s", unspents.Message)
	}
	if len(unspents.Data) == 0 {
		return nil, errors.New("eac empty unspents")
	}

	if len(unspents.Data) < 50 {
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("商户=[%d]，币种=[%s]，utxo数量：%d", appid, srv.CoinName, len(unspents.Data)))
	}

	if fee != 0 {
		if fee < 226000 || fee > 100000000 {
			//使用指定手续费
			return nil, errors.New("指定的手续费不在合理范围值[[0.00226-1]")
		}
		feeTmp = fee
	} else {
		//先提前预估手续费
		feeTmp = 226000
	}

	feeAmountTmp = decimal.New(feeTmp, 0)

	//排序unspent，先进行降序，找出大额的数值
	var sortUtxo transfer.EacUnspentDesc
	var sortUtxoTmp transfer.EacUnspentAsc //临时使用，小金额排序
	sortUtxoTmp = append(sortUtxoTmp, unspents.Data...)
	sort.Sort(sortUtxoTmp)
	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoTmp {
		rediskeyName := fmt.Sprintf("%s_%s_%d", rediskey.DOGE_UTXO_LOCK, uv.Txid, uv.Vout)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		amInt64 := uv.AmountInt64
		if amInt64.GreaterThanOrEqual(toAmountInt64.Add(feeAmountTmp)) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, amInt64.Shift(-8).String(), uv.Address)
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
		famountInt64 := v.AmountInt64
		oar := transfer.EacTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromIndex:  uint32(v.Vout),
			FromAmount: famountInt64.IntPart(),
		}
		fromAmountInt64 = fromAmountInt64.Add(famountInt64)
		txInTpl = append(txInTpl, oar)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.BIW_UTXO_LOCK_SECOND_TIME)

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
			fromAmountInt64.Shift(-8).String(),
			toAmountInt64.Shift(-8).String(),
			feeAmountTmp.Shift(-8).String(),
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
	txOutTpl = append(txOutTpl, transfer.EacTxOutTpl{
		ToAddr:   toAddr,
		ToAmount: toAmountInt64.IntPart(),
	})

	//计算找零金额
	changeAmount = fromAmountInt64.Sub(toAmountInt64).Sub(feeAmount)
	if changeAmount.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("找零金额异常使用金额：%s,出账金额：%s，手续费：%s，找零：%s",
			fromAmountInt64.Shift(-8).String(),
			toAmountInt64.Shift(-8).String(),
			feeAmount.Shift(-8).String(),
			changeAmount.Shift(-8).String(),
		)
	}
	if changeAmount.LessThanOrEqual(decimal.New(546, 0)) {
		//如果找零小于1，那么附加在手续费上
		feeAmount = feeAmount.Add(changeAmount)
	} else {
		//组装找零
		txOutTpl = append(txOutTpl, transfer.EacTxOutTpl{
			ToAddr:   changeAddr,
			ToAmount: changeAmount.IntPart(),
		})
	}
	orderReq.Fee = fee
	tpl = &transfer.EacTxTpl{
		MchId:    orderReq.MchName,
		OrderId:  orderReq.OuterOrderNo,
		CoinName: orderReq.CoinName,
		TxIns:    txInTpl,
		TxOuts:   txOutTpl,
	}
	return tpl, nil
}

//手续费计算
func (srv *EacTransferService) getFee(inNum, outNum int) (int64, error) {

	var (
		rate int64 = 1000
	)

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

	if rate == 0 {
		rate = 1000
	}

	fee := rate * byteNum
	//限定最小值
	if fee < 226000 {
		fee = 226000
	}
	//限制最大
	if fee > 150000000 {
		fee = 150000000
	}
	return fee, nil
}
