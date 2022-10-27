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
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type FoTransferService struct {
	CoinName string
}

func (srv *FoTransferService) VaildAddr(address string) error {
	return nil
}

func (srv *FoTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//无需实现
	return "", errors.New("implement me")
}

func NewFoTransferService() service.TransferService {
	return &FoTransferService{CoinName: "fo"}
}

func (srv *FoTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := srv.buildOrder(ta)
	if err != nil {
		//改变表状态 外层已经定义失败状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		if err != nil {
			log.Errorf("下单表订单id：%d,fo 构建异常:%s", ta.Id, err.Error())
			return err
		}
	}

	err = srv.WalletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,fo 获取发送交易异常:%s", ta.Id, err.Error())
		return err
	}

	return nil
}

//======================私有方法==================
type PubKey struct {
	Key string `json:"key"`
}

//私有方法 构建fo订单
func (srv *FoTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.FoOrderRequest, error) {
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
	orderReq := &transfer.FoOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.SignPubkey = pubKey
	if strings.ToLower(ta.CoinName) == "fo" && (strings.ToLower(ta.Eoskey) == "fo" || ta.Eoskey == "") {
		// 主链币
		orderReq.Token = fmt.Sprintf("eosio.token@eosio")
		orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(global.CoinDecimal[ta.CoinName].Decimal)+"f"+" %s", toAmount2, strings.ToUpper(ta.CoinName))
	} else {
		// 代币
		if ta.Eostoken != global.CoinDecimal[ta.Eoskey].Token {
			return nil, errors.New("fo 请求参数错误")
		}
		orderReq.Token = fmt.Sprintf("eosio.token@%s", global.CoinDecimal[ta.Eoskey].Token)
		orderReq.Quantity = fmt.Sprintf("%."+strconv.Itoa(global.CoinDecimal[ta.Eoskey].Decimal)+"f"+" %s", toAmount2, strings.ToUpper(ta.Eoskey))
	}

	return orderReq, nil
}

//创建交易接口参数
func (srv *FoTransferService) WalletServerCreate(orderReq *transfer.FoOrderRequest) error {
	reqdata, _ := json.Marshal(orderReq)
	log.Debug("post to walletserver data:", string(reqdata))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/fo/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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
