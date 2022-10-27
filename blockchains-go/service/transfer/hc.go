package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HcashOrg/hcd/hcutil"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"sync"
)

type HcTransferService struct {
	CoinName string
	Lock     *sync.Mutex
}

func NewHcTransferService() service.TransferService {
	return &HcTransferService{
		CoinName: "hc",
		Lock:     &sync.Mutex{},
	}
}

func (srv *HcTransferService) VaildAddr(address string) error {
	_, err := hcutil.DecodeAddress(address)
	if err != nil {
		return fmt.Errorf("hc valid address error: %v", err)
	}
	return nil
}

func (srv *HcTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	return "", errors.New("do not support hot transfer")
}

func (srv *HcTransferService) TransferCold(ta *entity.FcTransfersApply) error {

	//随机选择可用机器
	workerId := service.GetWorker(srv.CoinName)
	orderReq, err := srv.getEstimateTpl(ta, workerId)
	if err != nil {
		return err
	}
	err = srv.walletServerCreate(orderReq)
	if err != nil {
		//改变表状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		return err
	}
	return nil
}

func (srv *HcTransferService) getEstimateTpl(ta *entity.FcTransfersApply, worker string) (*transfer.HcOrderRequest, error) {
	var (
		changeAddress string
		toAddr        string
		toAmount      string
		coinSet       *entity.FcCoinSet //db币种配置
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
		return nil, fmt.Errorf("商户=[%d],查询%s找零地址失败", ta.AppId, srv.CoinName)
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
	//需要金额整型
	toAmount = toAddrAmount.Shift(int32(coinSet.Decimal)).String()
	var toList []*transfer.HcOrderToAddressList
	tl := &transfer.HcOrderToAddressList{
		Address:  toAddr,
		Quantity: toAmount,
	}
	toList = append(toList, tl)
	//构建订单
	orderReq := &transfer.HcOrderRequest{
		OrderRequestHead: transfer.OrderRequestHead{
			ApplyId:      int64(ta.Id),
			OuterOrderNo: ta.OutOrderid,
			OrderNo:      ta.OrderId,
			MchId:        int64(ta.AppId),
			MchName:      ta.Applicant,
			CoinName:     srv.CoinName,
			Worker:       worker,
		},
		ChangeAddress: changeAddress,
		ToList:        toList,
	}
	return orderReq, nil

}

//创建交易接口参数
func (srv *HcTransferService) walletServerCreate(orderReq *transfer.HcOrderRequest) error {
	d, _ := json.Marshal(orderReq)
	log.Infof("请求参数为： ", string(d))
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/hc/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
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
