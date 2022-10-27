package transfer

//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//
//	"github.com/group-coldwallet/blockchains-go/conf"
//	"github.com/group-coldwallet/blockchains-go/dao"
//	"github.com/group-coldwallet/blockchains-go/entity"
//	"github.com/group-coldwallet/blockchains-go/model/transfer"
//	"github.com/group-coldwallet/blockchains-go/pkg/util"
//	"github.com/group-coldwallet/blockchains-go/runtime/global"
//	"github.com/group-coldwallet/blockchains-go/service"
//	"github.com/group-coldwallet/blockchains-go/log"
//	"github.com/shopspring/decimal"
//
//	"xorm.io/builder"
//)
//
//type WbcTransferService struct {
//	CoinName string
//}
//
//func NewWbcTransferService() service.TransferService {
//	return &WbcTransferService{CoinName: "wbc"}
//}
//
//func (srv *WbcTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
//	var (
//		orderReq *transfer.WbcOrderReq
//	)
//	orderReq, err = srv.buildHotOrder(ta)
//	if err != nil {
//		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
//		return "", err
//	}
//	return srv.walletServerCreateHot(orderReq)
//}
//func (srv *WbcTransferService) TransferCold(ta *entity.FcTransfersApply) error {
//
//	return errors.New("this is a hot wallet")
//}
//func (srv *WbcTransferService) VaildAddr(address string) error {
//	return RubValidAddress(address)
//}
//
//func (srv *WbcTransferService) walletServerCreateHot(orderReq *transfer.WbcOrderReq) (string, error) {
//	cfg, ok := conf.Cfg.HotServers[srv.CoinName]
//	if !ok {
//		return "", fmt.Errorf("don't find %s config", srv.CoinName)
//	}
//	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/ruby-wbc/transfer", cfg.Url), cfg.User, cfg.Password, orderReq)
//	if err != nil {
//		return "", err
//	}
//	dd, _ := json.Marshal(orderReq)
//	log.Infof("%s 交易发送内容 :%s", srv.CoinName, string(dd))
//	log.Infof("%s 交易返回内容 :%s", srv.CoinName, string(data))
//	result := transfer.DecodeWalletServerRespOrder(data)
//	if result == nil {
//		return "", fmt.Errorf("order表请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
//	}
//	if result.Code != 0 || result.Data == nil {
//		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
//	}
//	txid := result.Data.(string)
//	return txid, nil
//}
//
//func (srv *WbcTransferService) buildHotOrder(ta *entity.FcTransfersApply) (*transfer.WbcOrderReq, error) {
//	var (
//		changeAddress string
//		toAddress     string
//		fee           int64           = 0 //默认等于0
//		toAmount      decimal.Decimal     //发送金额
//	)
//	//查询找零地址
//	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
//	if err != nil {
//		return nil, err
//	}
//	if len(changes) == 0 {
//		return nil, fmt.Errorf("wbc 商户=[%d],查询wbc找零地址失败", ta.AppId)
//	}
//	//随机选择
//	randIndex := util.RandInt64(0, int64(len(changes)))
//	changeAddress = changes[randIndex]
//	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
//	if err != nil {
//		return nil, err
//	}
//	//一般出账地址只有一个
//	if len(toAddrs) != 1 {
//		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
//	}
//	toAddress = toAddrs[0].Address
//	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
//	if toAmount.IsZero() {
//		return nil, errors.New("wbc toAmount  is zero")
//	}
//	coinSet := global.CoinDecimal[ta.CoinName]
//	if coinSet == nil {
//		return nil, fmt.Errorf("缺少币种信息")
//	}
//	tos := transfer.WbcChainParamsTransferTos{
//		ToAddress: toAddress,
//		ToAmount:  toAmount.Shift(int32(coinSet.Decimal)).IntPart(),
//	}
//	var toses []transfer.WbcChainParamsTransferTos
//	toses = append(toses, tos)
//	orderReq := &transfer.WbcOrderReq{}
//	orderReq.ApplyId = int64(ta.Id)
//	orderReq.OuterOrderNo = ta.OutOrderid
//	orderReq.OrderNo = ta.OrderId
//	orderReq.MchName = ta.Applicant
//	orderReq.CoinName = ta.CoinName
//	orderReq.Worker = service.GetWorker(srv.CoinName)
//
//	orderReq.ChangeAddress = changeAddress
//	orderReq.Fee = fee
//	orderReq.Tos = toses
//
//	return orderReq, nil
//}
