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
	"sync"
)

//交易端已经写入order_hot表 ，无需写入
type ZvcTransferService struct {
	Lock     *sync.Mutex
	CoinName string
}

func (zvcSrv *ZvcTransferService) VaildAddr(address string) error {
	url := global.CheckAddressServer[zvcSrv.CoinName]
	//url := conf.Cfg.HotServers[zvcSrv.CoinName].Url
	mapData := make(map[string]string, 0)
	mapData["coinname"] = zvcSrv.CoinName
	mapData["address"] = address
	zvcdata, _ := json.Marshal(mapData)
	data, err := util.PostJsonData(url, zvcdata)
	if err != nil {
		err = fmt.Errorf("验证地址错误，%s,address:%s, error:%s", zvcSrv.CoinName, address, err.Error())
		return err
	}
	log.Infof("验证地址返回结果：%s", string(data))
	zvcResp := decodeZvcAddrResult(data)
	if zvcResp != nil && zvcResp.Code == 0 {
		return nil
	}
	err = fmt.Errorf("验证地址错误，%s,address:%s", zvcSrv.CoinName, address)
	return err
}

func (zvcSrv *ZvcTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	//有并发问题
	zvcSrv.Lock.Lock()
	defer zvcSrv.Lock.Unlock()
	orderReq, err := zvcSrv.buildOrder(ta)
	if err != nil {
		//改变表状态 外层已经定义失败状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,cocos 构建异常:%s", ta.Id, err.Error())
		//	return err
		//}
		log.Errorf("下单表订单id：%d,cocos 构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	txid, err = zvcSrv.walletServerCreate(orderReq)
	if err != nil {
		//改变表状态 外层已经统一处理
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		//err = dao.FcTransfersApplyUpdateStatusById(ta.Id, 8)
		//if err != nil {
		//	log.Errorf("下单表订单id：%d,cocos 获取发送交易异常:%s", ta.Id, err.Error())
		//	return err
		//}
		log.Errorf("下单表订单id：%d,cocos 获取发送交易异常:%s", ta.Id, err.Error())
		return "", err
	}
	return txid, nil
}

func (zvcSrv *ZvcTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	panic("implement me")
}

func NewZvcTransferService() service.TransferService {
	return &ZvcTransferService{
		Lock:     &sync.Mutex{},
		CoinName: "zvc",
	}
}

//======================私有方法==================
//私有方法 构建cocos订单
func (srv *ZvcTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.ZvcOrderRequest, error) {
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
	fromAddr = fromAddrs[0] // 这里zvc只有一个

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
	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
	if toAmount.IsZero() {
		return nil, errors.New("zvc toAmount  is zero")
	}
	//填充参数
	orderReq := &transfer.ZvcOrderRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.FromAddress = fromAddr
	orderReq.CoinName = ta.CoinName
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.ToAmount = toAmount
	return orderReq, nil
}

//创建交易接口参数
func (zvcSrv *ZvcTransferService) walletServerCreate(orderReq *transfer.ZvcOrderRequest) (txid string, err error) {
	dd, _ := json.Marshal(orderReq)
	log.Infof("zvc 交易发送内容 :%s", string(dd))
	data, err := util.PostJsonByAuth(conf.Cfg.HotServers[orderReq.CoinName].Url+"/v1/zvc/Transfer", conf.Cfg.HotServers[orderReq.CoinName].User, conf.Cfg.HotServers[orderReq.CoinName].Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("zvc 交易返回内容 :%s", string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s", orderReq.OuterOrderNo)
	}
	txid = fmt.Sprintf("%v", result.Data)
	return txid, nil
}

type ZvcCheckAddrResp struct {
	Code int `json:"code"`
}

func decodeZvcAddrResult(data []byte) *ZvcCheckAddrResp {
	if len(data) != 0 {
		result := new(ZvcCheckAddrResp)
		//初始化为-1，实际状态 0 1 2  0 是正常  2是合约地址
		result.Code = -1
		err := json.Unmarshal(data, result)
		if err == nil {
			return result
		}
	}
	return nil
}
