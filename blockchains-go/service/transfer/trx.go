package transfer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"

	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

const (
	codeSuccess      = 0     // 签名成功
	codeNotBroadcast = 10000 // 尚未进行广播
)

type TrxTransferService struct {
	CoinName string
	limitMap sync.Map //限制发送频率
	Lock     *sync.Mutex
}

func NewTrxTransferService() service.TransferService {
	return &TrxTransferService{
		CoinName: "trx",
		limitMap: sync.Map{},
		Lock:     &sync.Mutex{},
	}
}

func (s *TrxTransferService) VaildAddr(address string) error {
	if !strings.HasPrefix(address, "T") {
		return fmt.Errorf("address is not has prefix 'T' :%s ", address)
	}
	decodeCheck := base58.Decode(address)
	if len(decodeCheck) == 0 {
		return fmt.Errorf("b58 decode %s error", address)
	}

	if len(decodeCheck) < 4 {
		return fmt.Errorf("b58 data length is less 4 : %s ", address)
	}

	if "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" == address {
		return fmt.Errorf("invalid address(usdt-trc20) %s", address)
	}

	decodeData := decodeCheck[:len(decodeCheck)-4]

	h256h0 := sha256.New()
	h256h0.Write(decodeData)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)

	if h1[0] == decodeCheck[len(decodeData)] &&
		h1[1] == decodeCheck[len(decodeData)+1] &&
		h1[2] == decodeCheck[len(decodeData)+2] &&
		h1[3] == decodeCheck[len(decodeData)+3] {
		return nil
	}
	return fmt.Errorf("b58 check sum error: %s", address)

}

func (s *TrxTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.TrxOrderRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err = s.buildOrderHot(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount, _ = decimal.NewFromString(orderReq.Amount)

	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.FromAddress,
		ToAddress:    orderReq.ToAddress,
		Amount:       amount.IntPart(), //转换整型
		Quantity:     amount.String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Token:        ta.Eostoken,
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}
	signResult, err := s.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Infof("订单=%s trx签名调用s.walletServerCreateHot出错:%v", ta.OutOrderid, err)
		return "", err
	}

	if signResult.Code != codeSuccess {
		if signResult.Code == codeNotBroadcast {
			// 如果签名服务还没有调用节点广播，直接把状态改为6，值班可以直接进行重推操作
			orderHot.Status = int(status.SignErrorStatus)
		} else {
			// 其他情况，状态改为7，需要开发处理
			orderHot.Status = int(status.BroadcastErrorStatus)
		}
		orderHot.ErrorMsg = signResult.Message
		dao.FcOrderHotInsert(orderHot)
		return "", errors.New(signResult.Message)
	}
	txid = signResult.TxId
	log.Infof("订单=%s trx签名成功，得到txId=%s", ta.OutOrderid, txid)
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}

func (s *TrxTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("do not support cold transfer")
}

//=================私有方法=================

func (s *TrxTransferService) walletServerCreateHot(orderReq *transfer.TrxOrderRequest) (*transfer.TrxSignRes, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return nil, fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return nil, err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", orderReq.OuterOrderNo, string(dd))
	log.Infof("%s 交易返回内容 :%s", orderReq.OuterOrderNo, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return nil, fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	resp := &transfer.TrxSignRes{Code: thr.Code, Message: thr.Message}
	if thr.Data != nil {
		txId, isStr := thr.Data.(string)
		if isStr {
			resp.TxId = txId
		}
	}
	if resp.Code == 0 && resp.TxId == "" {
		// 处理异常情况
		return nil, fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return resp, nil
}
func (s *TrxTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.TrxOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, err = decimal.NewFromString(toAddrs[0].ToAmount)
	if err != nil {
		return nil, err
	}
	coinType := ta.CoinName
	if ta.Eostoken != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet := global.CoinDecimal[coinType]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少%s币种配置", coinType)
	}
	fromAmount := toAmount.Add(decimal.Zero)
	if coinType == s.CoinName {
		fromAmount = toAmount.Add(decimal.NewFromFloat(0.1))
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount > ? and forzen_amount = 0", coinType, fromAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		dao.ReportColdAddrBalanceNotEnough(int64(ta.AppId), ta.OutOrderid, "trx", coinType, toAddrs[0].ToAmount)
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址，大于1.1 \n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]
	//判断地址前一笔交易是否已经过限制时间
	//err = s.isLimit(fromAddr, 10)

	if err != nil {
		return nil, err
	}
	orderReq := &transfer.TrxOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(int32(coinSet.Decimal)).String()
	if coinType != s.CoinName {
		coin, err := dao.FcCoinSetGetByName(coinType, 1)
		if err != nil {
			return nil, fmt.Errorf("没有找到coin_set配置：%s,%v", coinType, err)
		}
		if strings.ToLower(coinSet.Token) != strings.ToLower(ta.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coinSet.Token, ta.Eostoken)
		}

		log.Infof("contractAddress: %s", coin.Token)
		//trc10转账
		if s.isTrc10ContractAddress(coin.Token) {
			orderReq.AssetId = coin.Token
			return orderReq, nil
		}

		//trc20 转账
		orderReq.ContractAddress = coin.Token
		orderReq.FeeLimit = 10000000 //最大手续费 10个
	}
	return orderReq, nil
}

func (s *TrxTransferService) isTrc10ContractAddress(addr string) bool {
	for _, a := range addr {
		if a > 57 || a < 48 {
			return false
		}
	}
	return true
}

/*
判断地址是否被限制出账
second： 限制出账的秒数
*/
func (s *TrxTransferService) isLimit(address string, second int) error {
	v, ok := s.limitMap.Load(address)
	// 如果map中没有这个地址，直接可以出账
	if !ok {
		s.limitMap.Store(address, time.Now().Unix())
		return nil
	}
	lastTime, ok := v.(int64)
	if !ok {
		return fmt.Errorf("value is not int64 type : %v", v)
	}
	// 判断是否 超过限制时间
	limitTime := time.Unix(int64(lastTime), 0).Add(time.Second * time.Duration(second)).Unix()
	now := time.Now().Unix()
	if now >= limitTime {
		s.limitMap.Delete(address)
		s.limitMap.Store(address, time.Now().Unix())
		return nil
	}
	return fmt.Errorf("%s is limit,limit time %ds,now=[%s],last=[%s]",
		address, second, s.timeToStr(now), s.timeToStr(int64(limitTime)))
}

func (s *TrxTransferService) timeToStr(val int64) string {
	if val == 0 {
		return "2006-01-02 15:04:05"
	}
	tm := time.Unix(val, 0)
	return tm.Format("2006-01-02 15:04:05")
}
