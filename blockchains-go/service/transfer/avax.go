package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/utils/formatting"
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

type AvaxTransferService struct {
	CoinName   string
	DecimalBit int32
}

func NewAvaxTransferService() service.TransferService {
	return &AvaxTransferService{
		CoinName:   "avax",
		DecimalBit: 9,
	}
}

func (srv *AvaxTransferService) VaildAddr(address string) error {
	chainId, hrp, addrByte, err := formatting.ParseAddress(address)
	if err != nil {
		return fmt.Errorf("验证地址错误，err:%s", err.Error())
	}
	if strings.ToLower(chainId) != "x" || hrp != "avax" {
		return errors.New("error address version")
	}

	//目前测试网
	//if strings.ToLower(chainId) != "x" || hrp != "everest" {
	//	return errors.New("error address version")
	//}

	addr, err := formatting.FormatAddress(chainId, hrp, addrByte)
	if err != nil {
		return err
	}
	if addr != address {
		return errors.New("error address")
	}
	return err
}

func (srv *AvaxTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {

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
		return "", fmt.Errorf("getEstimateTpl error:%s", err.Error())
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

func (srv *AvaxTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//无需实现
	return errors.New("implement me")
}

//params worker 指定机器
func (srv *AvaxTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.AvaxTxTpl, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      int64
		fee           int64
		coinSet       *entity.FcCoinSet //db币种配置
		tpl           *transfer.AvaxTxTpl
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
		return nil, fmt.Errorf("Avax 商户=[%d],查询Avax找零地址失败", ta.AppId)
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
	//Avax api需要金额整型
	toAmount = toAddrAmount.Shift(srv.DecimalBit).IntPart()

	//手续费转换
	feeDecimal, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		log.Errorf("下单表订单id：%d,Avax 交易手续费转换异常:%s", ta.Id, err.Error())
		return nil, err
	}
	//Avax api需要金额整型
	fee = feeDecimal.Shift(srv.DecimalBit).IntPart()

	//todo Worker读取数据库
	orderReq := &transfer.AvaxOrderRequest{
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
		return nil, fmt.Errorf("setUtxoData error :%s", err.Error())
	}
	return tpl, nil
}

//创建交易接口参数
func (srv *AvaxTransferService) walletServerCreate(orderReq *transfer.AvaxTxTpl) (string, error) {
	log.Infof("avax 发送url：%s", conf.Cfg.HotServers[srv.CoinName].Url+"/v1/avax/transfer")
	dd, _ := json.Marshal(orderReq)
	log.Infof("avax 发送结构：%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/v1/avax/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("avax 发送返回：%s", string(data))
	result := transfer.DecodRespTranfer(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OrderNo)
	}
	if result.Code != 0 || result.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OrderNo)
	}
	txid := result.Txid
	return txid, nil
}

//为orderReq的 OrderAddress组装交易内容
//toAddr 接收地址
//changeAddr 找零地址
//appid 商户ID
//orderReq walletsever交易结构
//fee 手续费
func (srv *AvaxTransferService) setUtxoData(appid int64, orderReq *transfer.AvaxOrderRequest, toAddr, changeAddr string) (*transfer.AvaxTxTpl, error) {
	var (
		mchAmount       decimal.Decimal //商户余额
		fromAmountFloat decimal.Decimal //发送总金额
		toAmountFloat   decimal.Decimal //接收总金额
		fee             int64
		feeAmountFloat  decimal.Decimal   //临时估算手续费
		feeTmp          int64             //临时估算手续费
		AvaxCfg         *conf.CoinServers //币种数据服务
		coinSet         *entity.FcCoinSet //db币种配置
		//unspents     *transfer.AvaxUnspents
		err          error
		utxos        = make([]string, 0)
		tplUtxos     = make([]string, 0)
		sortUtxoAsc  util.AvaxUnspentAsc
		sortUtxoDesc util.AvaxUnspentDesc
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
		return nil, fmt.Errorf("发送金额不能小于0.00000546,目前金额：%s", decimal.New(orderReq.Amount, -1*srv.DecimalBit).String())
	}
	fee = orderReq.Fee

	toAmountFloat = decimal.New(orderReq.Amount, -1*srv.DecimalBit)

	mchAmountResult, err := dao.FcMchAmountGetInfo(int(appid), srv.CoinName)
	if err != nil {
		return nil, fmt.Errorf("查询商户余额异常：%s", err.Error())
	}
	mchAmount, _ = decimal.NewFromString(mchAmountResult.Amount)
	if mchAmount.LessThanOrEqual(toAmountFloat) {
		return nil, fmt.Errorf("商户:%d,币种：%s，余额不足,存量（包含发送中冻结）：%s，需要发送：%s,手续费未计算",
			appid,
			srv.CoinName,
			mchAmount.String(),
			toAmountFloat.String(),
		)
	}
	AvaxCfg = conf.Cfg.CoinServers[srv.CoinName]
	if AvaxCfg == nil {
		return nil, errors.New("配置文件缺少Avax coinservers设置")
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

	if fee != 0 {
		if fee < 1000000 || fee > 5000000000 {
			//使用指定手续费
			return nil, errors.New("指定的手续费不在合理范围值[[0.0001-5]")
		}
		feeTmp = fee
	} else {
		//先提前预估手续费
		//feeTmp = 1000000
		feeTmp, _ = util.AvaxGetTxFee(conf.Cfg.CoinServers[srv.CoinName].Url)
		if feeTmp == 0 {
			feeTmp = 1000000
		}
	}
	feeAmountFloat = decimal.New(feeTmp, -1*srv.DecimalBit)

	//查询utxo数量
	utxos, err = util.AvaxGetUtxos(conf.Cfg.CoinServers[srv.CoinName].Url, addrs...)
	if err != nil {
		return nil, err
	}

	//排序unspent，先进行降序，找出大额的数值
	//临时使用，小金额排序
	sortUtxoAsc, err = util.ParseUtxosBySortAsc(utxos)
	if err != nil {
		return nil, err
	}

	dd, _ := json.Marshal(sortUtxoAsc)
	log.Info(string(dd))

	//第一次遍历查询最优出账金额utxo
	for _, uv := range sortUtxoAsc {

		rediskeyName := fmt.Sprintf("%s_%s", rediskey.AVAX_UTXO_LOCK, uv.UtxoStr)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		amFloat := uv.AmountFolat
		if amFloat.GreaterThanOrEqual(toAmountFloat.Add(feeAmountFloat)) {
			log.Infof("订单：%s，查询到最符合出账utxo金额：%s,address:%s,刷新utxo列表", orderReq.OuterOrderNo, amFloat.String(), uv.Address)
			sortUtxoDesc = append(sortUtxoDesc, uv)
			break
		}
	}
	if len(sortUtxoDesc) == 0 {
		sortUtxoDesc = append(sortUtxoDesc, sortUtxoAsc...)
	}
	sort.Sort(sortUtxoDesc)

	dd, _ = json.Marshal(sortUtxoDesc)
	log.Info(string(dd))

	//组装from
	for _, v := range sortUtxoDesc {
		rediskeyName := fmt.Sprintf("%s_%s", rediskey.AVAX_UTXO_LOCK, v.UtxoStr)
		if has, _ := redisHelper.Exists(rediskeyName); has {
			//已经占用utxo，跳过
			continue
		}
		famountFloat := v.AmountFolat
		fromAmountFloat = fromAmountFloat.Add(famountFloat)
		tplUtxos = append(tplUtxos, v.UtxoStr)
		//临时存储进入redis 锁定2分钟
		redisHelper.Set(rediskeyName, orderReq.OuterOrderNo)
		redisHelper.Expire(rediskeyName, rediskey.AVAX_UTXO_LOCK_SECOND_TIME)

		if fromAmountFloat.GreaterThan(toAmountFloat.Add(feeAmountFloat)) {
			//满足出账
			break
		}
		//if len(utxoTpl) == conf.Cfg.UtxoScan.Num {
		//	//为了保证扫码稳定性 只使用15个utxo
		//	break
		//}
		//if len(txInTpl) == 100 {
		//	//为了保证扫码稳定性 只使用15个utxo
		//	break
		//}
	}
	if fromAmountFloat.LessThan(toAmountFloat) {
		return nil, fmt.Errorf("使用的utxo数量金额不足出账金额，请等待归集或者入账，商户余额(包含冻结)：%s，限量utxo使用金额：%s,出账金额：%s，预估手续费：%s",
			mchAmount.String(),
			fromAmountFloat.String(),
			toAmountFloat.String(),
			feeAmountFloat.String(),
		)
	}
	orderReq.Fee = feeAmountFloat.Shift(srv.DecimalBit).IntPart()
	//tpl := &transfer.AvaxTxTpl{
	//	MchId:    orderReq.MchName,
	//	OrderId:  orderReq.OuterOrderNo,
	//	CoinName: orderReq.CoinName,
	//	TxIns:    txInTpl,
	//	TxOuts:   txOutTpl,
	//}

	tpl := &transfer.AvaxTxTpl{
		CoinName:   srv.CoinName,
		OrderNo:    orderReq.OrderNo,
		MchName:    orderReq.MchName,
		FromAddr:   "",
		ToAddr:     toAddr,
		ChangeAddr: changeAddr,
		Amount:     toAmountFloat.Shift(srv.DecimalBit).IntPart(),
		Fee:        feeAmountFloat.Shift(srv.DecimalBit).IntPart(),
		Utxos:      tplUtxos,
	}
	return tpl, nil

}
