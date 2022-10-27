package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/group-coldwallet/blockchains-go/script/collect/btc/base"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"time"
)

type GxcTransferService struct {
	CoinName string
}

func NewGxcTransferService() service.TransferService {
	return &GxcTransferService{
		CoinName: "gxc",
	}
}

func (srv *GxcTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.GxcOrderRequestHot
		amount     decimal.Decimal             //发送金额
		createData []byte                      //构造交易信息
		result     *transfer.GxcTransferResult //出账信息
		feeInt64   decimal.Decimal
	)

	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinSet = global.CoinDecimal[ta.CoinName]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err = srv.buildOrderHot(ta)
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
		FromAddress:  orderReq.FromAccount,
		ToAddress:    orderReq.ToAccount,
		Amount:       amount.Shift(int32(coinSet.Decimal)).IntPart(), //转换整型
		Quantity:     amount.String(),
		Memo:         orderReq.Memo,
		Decimal:      int64(global.CoinDecimal[ta.CoinName].Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}

	result, err = srv.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		//写入热钱包表，创建失败
		return "", err
	}
	feeInt64, _ = decimal.NewFromString(result.Fee)
	feeInt64 = feeInt64.Shift(int32(coinSet.Decimal))
	orderHot.Fee = feeInt64.IntPart()
	orderHot.TxId = result.Txid
	orderHot.MemoEncrypt = result.Memo
	orderHot.Status = int(status.BroadcastStatus)
	//保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		//保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		//发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	txid = result.Txid
	return txid, nil
}

func (srv *GxcTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//随机选择可用机器
	workerId := service.GetWorker(base.BtcCoinName)
	orderReq, err := srv.buildOrder(ta, workerId)
	if err != nil {
		return err
	}
	err = srv.walletServerCreateCold(orderReq)
	if err != nil {
		//改变表状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		return err
	}
	return nil
}

func (srv *GxcTransferService) VaildAddr(address string) error {
	url := conf.Cfg.CoinServers[srv.CoinName].Url + "/api/v1/gxc/account?account=%s"
	url = fmt.Sprintf(url, address)
	data, err := util.Get(url)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", srv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	btcResp, err := transfer.DecodeGxcAddressResp(data)
	if err != nil {
		log.Errorf("验证地址错误，%s,address:%s,err=[%s]", srv.CoinName, address, err.Error())
		return fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
	}
	if btcResp.Data == "true" {
		return nil
	} else {
		err = fmt.Errorf("验证地址错误，%s,address:%s", srv.CoinName, address)
		return err
	}
}

//======================私有方法==================
//私有方法 构建mdu订单
func (srv *GxcTransferService) buildOrder(ta *entity.FcTransfersApply, worker string) (*transfer.GxcOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
		fee      decimal.Decimal
	)

	coinName := ta.CoinName
	if ta.Eoskey != "" {
		coinName = ta.Eostoken
	}
	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("gxc 缺少币种设置：%s", coinName)
	}
	// 查找from地址和金额
	fromAddrs, err := dao.FcGenerateAddressListFindAddresses(int(address.AddressTypeCold), int(address.AddressStatusAlloc), ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(fromAddrs) == 0 {
		return nil, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0] // 这里gxc只有一个

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
		return nil, errors.New("mdu toAmount  is zero")
	}

	//填充参数
	orderReq := &transfer.GxcOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAddress = toAddr
	orderReq.ToAmount = toAmount.Shift(int32(coinSet.Decimal)).IntPart()
	orderReq.Memo = ta.Memo
	orderReq.Worker = worker
	fee, _ = decimal.NewFromString(ta.Fee)
	if !fee.IsZero() {
		orderReq.Fee = fee.Shift(int32(coinSet.Decimal)).IntPart()
	}
	return orderReq, nil
}

//创建交易接口参数
func (srv *GxcTransferService) walletServerCreateCold(orderReq *transfer.GxcOrderRequest) error {
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/v1/gxc/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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

//私有方法 构建mdu订单
func (srv *GxcTransferService) buildOrderHot(ta *entity.FcTransfersApply) (*transfer.GxcOrderRequestHot, error) {
	type key struct {
		Key string `json:"key"`
	}

	var (
		fromAddr string
		toAddr   string
		pubkey   string
		toAmount decimal.Decimal
		fee      decimal.Decimal
	)
	pubKeyStruct := new(key)

	coinName := ta.CoinName
	if ta.Eoskey != "" {
		coinName = ta.Eostoken
	}
	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("gxc 缺少币种设置：%s", coinName)
	}
	// 查找from地址和金额
	fromAddrs, err := dao.FcGenerateAddressListFindAddressesData(int(address.AddressTypeCold), int(address.AddressStatusAlloc), ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(fromAddrs) == 0 {
		return nil, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0].Address // 这里gxc只有一个
	if fromAddrs[0].Json == "" {
		return nil, errors.New("缺少设置出账地址公钥失败")
	}
	json.Unmarshal([]byte(fromAddrs[0].Json), pubKeyStruct)

	pubkey = pubKeyStruct.Key
	if pubkey == "" {
		return nil, errors.New("查询出账地址公钥失败")
	}

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
		return nil, errors.New("mdu toAmount  is zero")
	}

	//填充参数
	orderReq := &transfer.GxcOrderRequestHot{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAccount = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAccount = toAddr
	orderReq.Amount = toAmount.String()
	orderReq.Memo = ta.Memo
	orderReq.PublicKey = pubkey
	fee, _ = decimal.NewFromString(ta.Fee)
	if !fee.IsZero() {
		orderReq.Fee = fee.String()
	}
	return orderReq, nil
}

//创建交易接口参数
func (s *GxcTransferService) walletServerCreateHot(orderReq *transfer.GxcOrderRequestHot) (*transfer.GxcTransferResult, error) {
	cfg, ok := conf.Cfg.HotServers[s.CoinName]
	if !ok {
		return nil, fmt.Errorf("don't find %s config", s.CoinName)
	}
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, s.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return nil, err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result := transfer.DecodeGxcTransferResp(data)
	if result == nil {
		return nil, fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return nil, fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}

	return result.Data, nil

}
