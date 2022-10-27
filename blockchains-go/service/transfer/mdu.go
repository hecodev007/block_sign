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
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

//各个币种对接文档：https://shimo.im/docs/UtQattVFYnEcIkuv

type MduTransferService struct {
	CoinName string
}

func (srv *MduTransferService) VaildAddr(address string) error {
	if !strings.HasPrefix(address, "mdu") {
		return errors.New("验证地址错误")
	}
	return nil
}

func (srv *MduTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	//无需实现
	return errors.New("implement me")
}

func NewMduTransferService() service.TransferService {
	return &MduTransferService{CoinName: "mdu"}
}

//由于交易端没有写入order_hot表  因此此方法需要写入order_hot表
func (srv *MduTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {

	mch, err := dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}

	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  0,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  "",
		ToAddress:    "",
		Token:        "",
		Amount:       0, //转换整型
		Quantity:     "",
		Memo:         "",
		Fee:          0,
		Decimal:      int64(global.CoinDecimal[ta.CoinName].Decimal),
		CreateData:   "",
		ErrorMsg:     "",
		ErrorCount:   0,
		Status:       int(status.CreateErrorStatus),
		IsRetry:      0,
		TxId:         "",
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
		Worker:       "",
	}

	orderReq, err := srv.buildOrder(ta)
	if err != nil {
		//改变表状态 外层已经定义失败状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,mdu 构建异常:%s", ta.Id, err.Error())
		//	return err
		//}
		log.Errorf("下单表订单id：%d,mdu 构建异常:%s", ta.Id, err.Error())

		orderHot.Status = int(status.CreateErrorStatus)
		orderHot.ErrorMsg = err.Error()
		//写入热钱包表，创建失败
		dao.FcOrderHotInsert(orderHot)
		return "", err
	}
	orderHot.FromAddress = orderReq.FromAddress
	orderHot.ToAddress = orderReq.ToAddress
	orderHot.Token = orderReq.Token
	//orderHot.Amount = orderReq.ToAmount.Mul(decimal.New(1, int32(orderReq.Decimal))).IntPart()
	orderHot.Amount = orderReq.ToAmountFloat.Shift(int32(orderReq.Decimal)).IntPart()
	orderHot.Quantity = orderReq.ToAmountFloat.String()
	orderHot.Memo = orderReq.Memo

	txid, err = srv.WalletServerCreate(orderReq)
	if err != nil {
		//改变表状态 外层已经统一处理
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,mdu 获取发送交易异常:%s", ta.Id, err.Error())
		//	return err
		//}

		log.Errorf("下单表订单id：%d,mdu 获取发送交易异常:%s", ta.Id, err.Error())

		orderHot.Status = int(status.UnknowErrorStatus)
		orderHot.ErrorMsg = err.Error()
		//写入热钱包表，创建失败
		dao.FcOrderHotInsert(orderHot)
		return "", err
	}

	orderHot.Status = int(status.BroadcastStatus)
	orderHot.TxId = txid
	//写入热钱包表，广播成功
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		log.Errorf("热钱包交易，数据库写入交易失败,订单outOrderId：%s，txid:%s,err:%s", orderReq.OuterOrderNo, txid, err.Error())
	}
	return txid, nil
}

//======================私有方法==================
//私有方法 构建mdu订单
func (srv *MduTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.MduOrderRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)

	// 查找from地址和金额
	fromAddrs, err := dao.FcGenerateAddressListFindAddresses(int(address.AddressTypeCold), int(address.AddressStatusAlloc), ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(fromAddrs) == 0 {
		return nil, errors.New("查找出账地址失败")
	}
	fromAddr = fromAddrs[0] // 这里mdu只有一个

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
	orderReq := &transfer.MduOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAddress = toAddr
	orderReq.ToAmountFloat = toAmount
	orderReq.Memo = ta.Memo
	orderReq.Decimal = global.CoinDecimal[ta.CoinName].Decimal
	return orderReq, nil
}

//创建交易接口参数
func (srv *MduTransferService) WalletServerCreate(orderReq *transfer.MduOrderRequest) (txid string, err error) {
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[srv.CoinName].Url+"/transfer", conf.Cfg.HotServers[srv.CoinName].User, conf.Cfg.HotServers[srv.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("mdu 交易发送内容 :%s", string(dd))
	log.Infof("mdu 交易返回内容 :%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常：%s，outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}
