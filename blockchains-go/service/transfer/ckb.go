package transfer

import (
	"encoding/json"
	_ "errors"
	"fmt"
	_ "fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	_ "github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"sync"

	"github.com/nervosnetwork/ckb-sdk-go/address"
	//model "github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
)

type CkbTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewCkbTransferService() service.TransferService {
	return &CkbTransferService{
		CoinName: "ckb",
		Lock:     &sync.Mutex{},
	}
}

func (s *CkbTransferService) VaildAddr(addr string) error {
	_, e := address.Parse(addr)
	if e != nil {
		return fmt.Errorf("验证地址错误，%s,address:%s, error:%s", s.CoinName, addr, e.Error())
	}
	return nil
	//url := conf.Cfg.CoinServers[s.CoinName].Url + "/vaildaddress?address=%s"
	//url = fmt.Sprintf(url, address)
	//data, err := util.Get(url)
	//if err != nil {
	//	err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", s.CoinName, address, err.Error())
	//	return err
	//}
	//log.Infof("验证地址返回结果：%s", string(data))
	//btcResp := transfer.DecodeCkbAddressResult(data)
	//if btcResp != nil && btcResp.Data != nil {
	//	if btcResp.Data.Vaild {
	//		return nil
	//	}
	//}
	//err = fmt.Errorf("验证地址错误，%s,address:%s", s.CoinName, address)
	//return err
}

func (s *CkbTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	panic("implement me")
}

func (s *CkbTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := s.buildOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return err
	}
	err = s.walletServerCreate(orderReq)
	if err != nil {
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		return err
	}
	return nil
}

//=================私有方法=================
//创建交易接口参数
func (s *CkbTransferService) walletServerCreate(orderReq *transfer.CkbOrderRequest) error {
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/create", conf.Cfg.Walletserver.Url, s.CoinName), conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", s.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", s.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//私有方法 构建ckb订单
func (s *CkbTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.CkbOrderRequest, error) {
	////获取币种的配置
	//ckbCoins, err := entity.FcCoinSet{}.Find(builder.Eq{"status": 1}.And(builder.Eq{"name": s.CoinName}))
	//if err != nil {
	//	return nil,err
	//}
	//if len(ckbCoins) == 0 {
	//	return nil,errors.New("do not find dot coin")
	//}
	////只有一个coin
	//coin := ckbCoins[0]

	//查找冷地址
	coldAddress, err1 := dao.FcGenerateAddressListFindAddressesData(1, 2, ta.AppId, s.CoinName)
	if err1 != nil || len(coldAddress) == 0 {
		return nil, fmt.Errorf("无法获取币种%s冷地址", s.CoinName)
	}
	from := coldAddress[0].Address
	//查找to地址
	//查询出账地址和金额
	toAddrs, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(ta.Id, "to")
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,查找接收地址异常", ta.Id, ta.OutOrderid)
	}
	to := toAddrs[0].Address
	toAddrAmount, _ := decimal.NewFromString(toAddrs[0].ToAmount)
	if toAddrAmount.LessThan(decimal.NewFromFloat(61)) {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接收地址金额异常,最小金额61", ta.Id, ta.OutOrderid)
	}
	//	构建订单
	orderReq := &transfer.CkbOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     s.CoinName,
			Worker:       service.GetWorker(s.CoinName),
		},
	}
	var orderAddreses []map[string]interface{}
	fromOrder := map[string]interface{}{
		"dir":     0,
		"address": from,
	}
	toOrder := map[string]interface{}{
		"dir":      1,
		"address":  to,
		"quantity": toAddrAmount.String(),
	}
	changeOrder := map[string]interface{}{
		"dir":     2,
		"address": from,
	}
	orderAddreses = append(orderAddreses, fromOrder)
	orderAddreses = append(orderAddreses, toOrder)
	orderAddreses = append(orderAddreses, changeOrder)

	orderReq.OrderAddress = orderAddreses
	orderReq.FeeString = ta.Fee
	return orderReq, nil
}
