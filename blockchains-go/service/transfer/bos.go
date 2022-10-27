package transfer

import (
	"encoding/json"
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
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"time"
	"xorm.io/builder"
)

type BosTransferService struct {
	CoinName string
	//Lock     *sync.Mutex
}

func NewBosTransferService() service.TransferService {
	return &BosTransferService{
		CoinName: "bos",
		//Lock:     &sync.Mutex{},
	}
}
func (tt *BosTransferService) VaildAddr(address string) error {
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

func (tt *BosTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		orderReq   *transfer.BosOrderRequest
		createData []byte //构造交易信息
		coinSet    *entity.FcCoinSet
	)
	orderReq, err = tt.buildOrder(ta, false)
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

func (tt *BosTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return nil
}
func (tt *BosTransferService) buildOrder(ta *entity.FcTransfersApply, isCold bool) (*transfer.BosOrderRequest, error) {
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
	orderReq := new(transfer.BosOrderRequest)
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
	qcoin := "BOS"
	orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(coin.Decimal)+"f"+" %s", toAmount2, qcoin)
	//if isCold {
	//	//todo 切换冷钱包的时候做处理
	//	reqBosData.BlockId = ""
	//}
	return orderReq, nil
}

func (tt *BosTransferService) walletServiceTransfer(orderReq *transfer.BosOrderRequest) (string, error) {
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
