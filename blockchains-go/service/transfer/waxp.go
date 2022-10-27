package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"xorm.io/builder"

	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type WaxpTransferService struct {
	CoinName string
}

func (tt *WaxpTransferService) VaildAddr(address string) error {
	cfg, ok := conf.Cfg.HotServers[tt.CoinName]
	if !ok {
		return fmt.Errorf("don't find %s config", tt.CoinName)
	}
	params := make(map[string]string)
	params["address"] = address
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/validAddress", cfg.Url, tt.CoinName), cfg.User, cfg.Password, params)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(params)
	log.Infof("验证地址%s 交易发送内容 :%s", tt.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", tt.CoinName, string(data))
	return transfer.DecodeValidAddressResp(data)
}

func (tt *WaxpTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		orderReq   *transfer.WaxpHotOrderRequest
		createData []byte //构造交易信息
		coinSet    *entity.FcCoinSet
	)
	orderReq, err = tt.buildHotOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
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
		Amount:       orderReq.Amount, //转换整型
		Quantity:     orderReq.Quantity,
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}
	//发送交易
	txid, err = tt.walletServiceTransfer(orderReq)
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

func NewWaxpTransferService() service.TransferService {
	return &WaxpTransferService{CoinName: "waxp"}
}

func (srv *WaxpTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("implement me")
	orderReq, err := srv.buildOrder(ta)
	if err != nil {
		//改变表状态 外层已经定义失败状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		if err != nil {
			log.Errorf("下单表订单id：%d,waxp 构建异常:%s", ta.Id, err.Error())
			return err
		}
	}

	err = srv.WalletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,waxp 获取发送交易异常:%s", ta.Id, err.Error())
		return err
	}

	return nil
}

//私有方法 构建wax订单
func (srv *WaxpTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.WaxpOrderRequest, error) {
	var (
		fromAddr string
		pubKey   string
		toAddr   string
		toAmount decimal.Decimal
	)

	if global.CoinDecimal[ta.CoinName] == nil {
		return nil, errors.New("币种不支持")
	}

	// 查找from地址和金额
	fromAddrs, err := dao.FcGenerateAddressListFindAddressesData(int(address.AddressTypeCold), int(address.AddressStatusAlloc), ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(fromAddrs) == 0 {
		return nil, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0].Address // 这里fo只有一个

	pub := new(PubKey)
	json.Unmarshal([]byte(fromAddrs[0].Json), pub)
	pubKey = pub.Key

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
	//toAmount = toAddrs[0].ToAmount
	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
	if toAmount.IsZero() {
		return nil, errors.New("fo toAmount  is zero")
	}
	toAmount2, _ := toAmount.Float64()

	//填充参数
	orderReq := &transfer.WaxpOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.SignPubkey = pubKey
	if strings.ToLower(ta.CoinName) == "waxp" && (strings.ToLower(ta.Eoskey) == "waxp" || ta.Eoskey == "") {
		// 主链币
		orderReq.Token = fmt.Sprintf("eosio.token")
		orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(global.CoinDecimal[ta.CoinName].Decimal)+"f"+" %s", toAmount2, strings.ToUpper(ta.CoinName))
	} else {
		return nil, errors.New("unknow waxp token")
	}

	return orderReq, nil
}

//创建交易接口参数
func (srv *WaxpTransferService) WalletServerCreate(orderReq *transfer.WaxpOrderRequest) error {
	reqdata, _ := json.Marshal(orderReq)
	log.Debug("post to walletserver data:", string(reqdata))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/waxp/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	if data == nil {
		log.Error("请求失败")
		return errors.New("请求失败")
	}
	log.Debug(string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s", orderReq.OuterOrderNo)
	}
	return nil
}

func (tt *WaxpTransferService) buildHotOrder(ta *entity.FcTransfersApply) (*transfer.WaxpHotOrderRequest, error) {
	var (
		fromAddr string
		pubKey   string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.Find(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	//查询出账地址和金额
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
	coinType := strings.ToLower(ta.CoinName)
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coin := global.CoinDecimal[coinType]
	if coin == nil {
		return nil, fmt.Errorf("读取 %s coinSet 设置异常", coinType)
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs[0].Address)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]
	pub := new(PubKey)
	json.Unmarshal([]byte(coldAddrs[0].Json), pub)
	pubKey = pub.Key
	toAmount2, _ := toAmount.Float64()

	//填充参数
	orderReq := new(transfer.WaxpHotOrderRequest)
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(ta.CoinName)

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.SignPubkey = pubKey
	orderReq.Amount = toAmount.IntPart()
	if ta.Eostoken != "" {
		if coin.Token != ta.Eostoken {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coin.Token, ta.Eostoken)
		}
		orderReq.Token = ta.Eostoken
	} else {
		orderReq.Token = "eosio.token"
	}
	qcoin := strings.ToUpper(coinType)
	if strings.Contains(qcoin, "-") {
		qcoin = strings.Replace(qcoin, "WAXP", "", 1)
		qcoin = strings.Replace(qcoin, "WAX", "", 1)
		qcoin = strings.Replace(qcoin, "-", "", 1)

	}
	//qcoin := "WAXP"
	orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(coin.Decimal)+"f"+" %s", toAmount2, qcoin)
	return orderReq, nil
}

func (tt *WaxpTransferService) walletServiceTransfer(orderReq *transfer.WaxpHotOrderRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[tt.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", tt.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, tt.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", tt.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", tt.CoinName, string(data))
	thr, err1 := transfer.DecodeTransferHotResp(data)
	if err1 != nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if thr.Code != 0 || thr.Txid == "" {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return thr.Txid, nil
}
