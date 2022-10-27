package dingding

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

type EthDingService struct {
	*BaseDingService
}

func (bds *BaseDingService) NewEthDingService() *EthDingService {
	eds := new(EthDingService)
	eds.BaseDingService = bds
	return eds
}

func (eds *EthDingService) TransferFee(feeAddr, toAddr string, appId int64, feeApply *entity.FcTransfersApply, fee decimal.Decimal) error {

	orderReq := &transfer.EthTransferFeeReq{}
	orderReq.ApplyId = appId
	orderReq.OuterOrderNo = feeApply.OutOrderid
	orderReq.OrderNo = feeApply.OrderId
	orderReq.MchId = int64(feeApply.AppId)
	orderReq.MchName = feeApply.Applicant
	orderReq.CoinName = "eth"
	orderReq.FromAddr = feeAddr
	orderReq.ToAddrs = []string{toAddr}
	orderReq.NeedFee = fee.Shift(18).String() //eth -> wei
	if err := walletServerFee(feeApply.CoinName, orderReq); err != nil {
		return fmt.Errorf("[%s] 地址大手续费错误，Err=[%v]", toAddr, err)
	}
	return nil
}

func (eds *EthDingService) CollectToken(name, to string, mch *entity.FcMch, fromAddresses []string, tokenCoinSet *entity.FcCoinSet) error {
	//构建订单
	cltApply := &entity.FcTransfersApply{
		Username:   "Robot",
		CoinName:   name,
		Department: "blockchains-go",
		OutOrderid: fmt.Sprintf("COLLECT_%d", time.Now().Nanosecond()),
		OrderId:    util.GetUUID(),
		Applicant:  mch.Platform,
		Operator:   "Robot",
		AppId:      mch.Id,
		Type:       "gj",
		Purpose:    fmt.Sprintf("%s自动归集", name),
		Status:     int(entity.ApplyStatus_Merge), //因为是即时归集，所以直接把状态置为构建成功
		Createtime: time.Now().Unix(),
		Lastmodify: util.GetChinaTimeNow(),
		Source:     1,
		Eostoken:   tokenCoinSet.Token,
		Eoskey:     tokenCoinSet.Name,
	}

	applyAddresses := make([]*entity.FcTransfersApplyCoinAddress, 0)
	applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
		Address:     to,
		AddressFlag: "to",
		Status:      0,
		Lastmodify:  cltApply.Lastmodify,
	})
	collectAddrs := make([]string, 0)
	for _, from := range fromAddresses {
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     from,
			AddressFlag: "from",
			Status:      0,
			Lastmodify:  cltApply.Lastmodify,
		})
		collectAddrs = append(collectAddrs, from)
	}

	appId, err := cltApply.TransactionAdd(applyAddresses)
	if err != nil {
		return fmt.Errorf("build app id error,%v", err)
	}
	//开始请求钱包服务归集
	orderReq := &transfer.EthCollectReq{}
	orderReq.ApplyId = appId
	orderReq.OuterOrderNo = cltApply.OutOrderid
	orderReq.OrderNo = cltApply.OrderId
	orderReq.MchId = int64(mch.Id)
	orderReq.MchName = mch.Platform
	orderReq.CoinName = name
	orderReq.FromAddrs = collectAddrs
	orderReq.ToAddr = to
	orderReq.ContractAddr = tokenCoinSet.Token
	orderReq.Decimal = tokenCoinSet.Decimal
	if err := walletServerCollect(orderReq, name); err != nil {
		return fmt.Errorf("%s 归集失败，Err： %v", name, err)
	}
	return nil
}

func walletServerFee(coinName string, orderReq *transfer.EthTransferFeeReq) error {
	cfg := conf.Cfg.Walletserver
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/fee", cfg.Url, coinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s fee send :%s", coinName, string(dd))
	log.Infof("%s fee resp :%s", coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerFee 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerFee 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

//创建交易接口参数
func walletServerCollect(orderReq *transfer.EthCollectReq, coinName string) error {
	cfg := conf.Cfg.Walletserver
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/%s/collect", cfg.Url, coinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s Collect send :%s", coinName, string(dd))
	log.Infof("%s Collect resp :%s", coinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("walletServerCollect 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 {
		log.Error(result)
		return fmt.Errorf("walletServerCollect 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return nil
}

func (eds *EthDingService) FindCoinFee(mainName, address string, mch *entity.FcMch) (chainAmount string, err error) {
	//2. 获取链上余额
	balance, err := eds.getTokenBalance("", address)
	if err != nil {
		return "", err
	}
	chainAmount = balance.Shift(-18).String()
	return
}

func (eds *EthDingService) getTokenBalance(contractaddress, addr string) (decimal.Decimal, error) {
	apikey := "MXKM5DKHND1KUGKF3PPIDQQJXC2IRIDVUV"
	url := "https://api.etherscan.io/api?module=account&action=balance&address=%s&tag=latest&apikey=%s"
	if strings.TrimSpace(addr) == "" {
		return decimal.Zero, errors.New("empty addr blanance")
	}
	if contractaddress == "" {
		url = fmt.Sprintf(url, addr, apikey)
	} else {
		url = "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=%s&address=%s&tag=latest&apikey=%s"
		url = fmt.Sprintf(url, contractaddress, addr, apikey)
	}

	resultData, err := util.Get(url)
	if err != nil {
		log.Error(string(resultData))
		return decimal.Zero, err
	}
	result := new(BalanceStruct)
	err = json.Unmarshal(resultData, result)
	if err != nil {
		log.Error(string(resultData))
		return decimal.Zero, err
	}

	if result.Status != "1" {
		return decimal.Zero, errors.New(string(resultData))
	}
	return result.Result, nil
}

type BalanceStruct struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  decimal.Decimal `json:"result"`
	//"status": "1",
	//"message": "OK",
	//"result": "4009811415661147191"
}
