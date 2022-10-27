package dingding

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
	"strings"
	"time"
	"xorm.io/builder"
)

type BscDingService struct {
	*BaseDingService
}

func (bds *BaseDingService) NewBscDingService() *BscDingService {
	eds := new(BscDingService)
	eds.BaseDingService = bds
	return eds
}

func (eds *BscDingService) TransferFee(feeAddr, toAddr string, appId int64, feeApply *entity.FcTransfersApply, fee decimal.Decimal) error {
	//1. 构建交易数据
	orderReq := &transfer.BscOrderRequest{}
	orderReq.ApplyId = int64(appId)
	orderReq.OuterOrderNo = feeApply.OutOrderid
	orderReq.OrderNo = feeApply.OrderId
	orderReq.MchId = int64(feeApply.AppId)
	orderReq.MchName = feeApply.Applicant
	orderReq.CoinName = feeApply.CoinName

	orderReq.FromAddress = feeAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = fee.Shift(int32(18)).String()

	//2. 构建order_hot
	coinSet := global.CoinDecimal[feeApply.CoinName]

	if coinSet == nil {
		return fmt.Errorf("缺少币种%s的coinSet信息", feeApply.CoinName)
	}
	createData, _ := json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      feeApply.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: feeApply.OutOrderid,
		OrderNo:      feeApply.OrderId,
		MchName:      feeApply.Applicant,
		CoinName:     feeApply.CoinName,
		FromAddress:  orderReq.FromAddress,
		ToAddress:    orderReq.ToAddress,
		Amount:       fee.Shift(18).IntPart(), //转换整型
		Quantity:     fee.String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
	}
	txid, err := eds.walletServerCreateHot(orderReq, strings.ToLower(feeApply.CoinName))
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", feeApply.Id, err.Error())
		return err
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
	return nil
}

func (eds *BscDingService) walletServerCreateHot(orderReq *transfer.BscOrderRequest, coinName string) (string, error) {
	cfg, ok := conf.Cfg.HotServers[coinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", coinName)
	}
	timeNow := fmt.Sprintf("%d", time.Now().Unix())
	//添加参数签名
	sig, err := util.CreateTransferParamsSign(orderReq.FromAddress, orderReq.ToAddress, orderReq.Amount, orderReq.ContractAddress, timeNow)
	if err != nil {
		return "", fmt.Errorf("创建transfer参数签名错误： %s", err)
	}
	orderReq.Sign = sig
	orderReq.CurrentTime = timeNow
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, coinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", coinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", coinName, string(data))
	result := transfer.DecodeBscTransferResp(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result["code"].(float64) != 0 || result["data"] == nil {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result["data"].(string), nil
}

func (eds *BscDingService) CollectToken(name, to string, mch *entity.FcMch, fromAddresses []string, tokenCoinSet *entity.FcCoinSet) error {
	var errMsg = "代币归集有错误："
	var success int
	for _, from := range fromAddresses {
		// 查找amount

		thresh := 0.0
		collectThreshold, _ := decimal.NewFromString(tokenCoinSet.CollectThreshold)
		collectThresholdFloat, _ := collectThreshold.Float64()
		if collectThresholdFloat <= 0 {
			log.Infof("代币：%s,没有设置参数，使用默认金额：%v", tokenCoinSet.Name, thresh)
		} else {
			thresh = collectThresholdFloat
		}
		log.Infof("%s 归集的最小金额为： %f", tokenCoinSet.Name, thresh)
		//4。 获取有余额的地址
		fromAddrs, err := entity.FcAddressAmount{}.FindAddressAndAmount(builder.Eq{"address": from, "coin_type": tokenCoinSet.Name, "app_id": mch.Id}.
			And(builder.Expr("amount >= ? and forzen_amount = 0", thresh)), 10)
		if err != nil {
			return fmt.Errorf("币种%s查询地址%s余额错误：%v", tokenCoinSet.Name, from, err)
		}
		if len(fromAddrs) != 1 {
			return fmt.Errorf("币种%s未查找到地址%s余额", tokenCoinSet.Name, from)
		}
		amount, _ := decimal.NewFromString(fromAddrs[0].Amount)
		//生成订单
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
			Address:     from,
			AddressFlag: "from",
			Status:      0,
		})
		applyAddresses = append(applyAddresses, &entity.FcTransfersApplyCoinAddress{
			Address:     to,
			AddressFlag: "to",
			Status:      0,
		})
		appId, err := cltApply.TransactionAdd(applyAddresses)
		if err == nil {
			orderReq := &transfer.BscOrderRequest{}
			orderReq.ApplyId = int64(appId)
			orderReq.OuterOrderNo = cltApply.OutOrderid
			orderReq.OrderNo = cltApply.OrderId
			orderReq.MchId = int64(cltApply.AppId)
			orderReq.MchName = cltApply.Applicant
			orderReq.CoinName = cltApply.CoinName

			orderReq.FromAddress = from
			orderReq.ToAddress = to
			orderReq.ContractAddress = cltApply.Eostoken
			orderReq.Token = cltApply.Eoskey
			orderReq.Amount = amount.Shift(int32(tokenCoinSet.Decimal)).String()

			//构建order_hot订单
			createData, _ := json.Marshal(orderReq)
			orderHot := &entity.FcOrderHot{
				ApplyId:      cltApply.Id,
				ApplyCoinId:  tokenCoinSet.Id,
				OuterOrderNo: cltApply.OutOrderid,
				OrderNo:      cltApply.OrderId,
				MchName:      cltApply.Applicant,
				CoinName:     cltApply.CoinName,
				FromAddress:  orderReq.FromAddress,
				ToAddress:    orderReq.ToAddress,
				Amount:       amount.Shift(int32(tokenCoinSet.Decimal)).IntPart(), //转换整型
				Quantity:     amount.String(),
				Decimal:      int64(tokenCoinSet.Decimal),
				CreateData:   string(createData),
				Status:       int(status.UnknowErrorStatus),
				CreateAt:     time.Now().Unix(),
				UpdateAt:     time.Now().Unix(),
			}
			txid, err := eds.walletServerCreateHot(orderReq, strings.ToLower(cltApply.CoinName))
			if err != nil {
				orderHot.Status = int(status.BroadcastErrorStatus)
				orderHot.ErrorMsg = err.Error()
				dao.FcOrderHotInsert(orderHot)
				log.Errorf("下单表订单id：%d,获取发送交易异常:%s", cltApply.Id, err.Error())
				errMsg = fmt.Sprintf("%s【%v】", errMsg, err)
				continue
			}
			orderHot.TxId = txid
			orderHot.Status = int(status.BroadcastStatus)
			// 保存热表
			err = dao.FcOrderHotInsert(orderHot)
			if err != nil {
				err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
				// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
				log.Error(err.Error())
				errMsg = fmt.Sprintf("%s【%v】", errMsg, err)
				// 发送给钉钉
				dingding.ErrTransferDingBot.NotifyStr(err.Error())
				continue
			}
			success++
		} else {
			errMsg = fmt.Sprintf("%s【%v】", errMsg, err)
			continue
		}
	}
	if success < len(fromAddresses) {
		return fmt.Errorf("总共需要归集代币数量：%d,成功： %d，错误原因：%s", len(fromAddresses), success, errMsg)
	}
	return nil
}

func (eds *BscDingService) FindCoinFee(mainName, address string, mch *entity.FcMch) (chainAmount string, err error) {

	// 查找链上余额
	balance, err := eds.getChainBalance(mainName, address, "")
	if err != nil {
		return "", fmt.Errorf("获取链上余额错误：%v", err)
	}
	bal, _ := decimal.NewFromString(balance)
	chainAmount = bal.Shift(-18).String()
	return
}

func (eds *BscDingService) getChainBalance(coinName, address, contract string) (string, error) {

	cfg, ok := conf.Cfg.HotServers[coinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", coinName)
	}
	var req transfer.ReqGetBalanceParams
	req.Address = address
	req.CoinName = "bsc"
	req.ContractAddress = contract
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/bsc/getBalance", cfg.Url), cfg.User, cfg.Password, req)
	if err != nil {
		return "", err
	}
	gbr, err := transfer.DecodeGetBalanceResp(data)
	if err != nil {
		return "", err
	}
	balance, ok := gbr.Data.(string)
	if !ok {
		return "", fmt.Errorf("%v is not string", gbr.Data)
	}
	return balance, nil
}
