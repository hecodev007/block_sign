package service

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/domain/dto"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SendBillOutMsg
// 发送提币消息
func SendBillOutMsg(serialNo string) error {
	log.Infof("%s 发送提币消息", serialNo)
	var billDao = new(bill.BillDetail)
	datas, err := billDao.GetBillBySerialNo(serialNo)
	if err != nil {
		log.Errorf("根据serialNo='%s', 查询账单失败，%s", serialNo, err.Error())
		return err
	}
	b, err := json.Marshal(datas)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	binfo := domain.BillInfo{}
	json.Unmarshal(b, &binfo)
	log.Infof("发送提币消息 %v", binfo)
	err = MerchantWithdrawal(binfo, 3)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	return nil
}

func MerchantWithdrawal(billData domain.BillInfo, billStatus int) error {

	var (
		billDao = new(bill.BillDetail)
	)
	// 审核确认时间更新
	err := billDao.UpdateBillBySerialNo(billData.SerialNo, map[string]interface{}{
		"audit_time": time.Now().Local(),
	})

	if err != nil {
		log.Errorf("审核确认时间更新:%s", err.Error())
		return err
	}
	// 拒绝
	if billStatus == 5 {
		err := RollbackBillAssets(&billData)
		if err != nil {
			log.Errorf("拒绝:%s", err.Error())
			return err
		}
	}
	// 通过，发送
	if billStatus == 3 {
		log.Infof(" 通过，发送3%v", billData)
		billData, err = FindBillInfoBySerialNo(billData.SerialNo)
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}
		pInfo := serviceSecurity.NewEntity()
		err = pInfo.GetBindInfoByBid(billData.ServiceId)
		if err != nil {
			log.Errorf("GetBindInfoByBid：%v", err)
			return err
		}
		if pInfo == nil && pInfo.Id == 0 {
			log.Errorf("查询业务线失败：%d,业务线为空", billData.ServiceId)
			return fmt.Errorf("查询业务线失败：%d,业务线为空", billData.ServiceId)
		}
		// 先入库，以防失败，新增链上账单
		cbDao := chainBill.NewEntity()
		err = cbDao.FindChainBillBySerialNo(billData.SerialNo)
		if err != nil {
			log.Errorf("创建CreateChainBill：%v", err)
			return err
		}
		if cbDao.Id == 0 {
			cbDao = dto.GetChainBillInfo(billData)
			err = CreateChainBill(cbDao)
			if err != nil {
				log.Errorf("创建CreateChainBill：%v", err)
				return err
			}
		}

		// 查询链上订单表，待上链数据
		//chain, err := FindChainBillNoUpChainSerialNo(cbDao.SerialNo)
		//if err != nil {
		//	return false
		//}
		var chainName string
		var tokenName string
		var contractAddress string
		chainName = FindChainName(billData.CoinName)
		upC := strings.ToUpper(billData.CoinName)
		if upC != chainName {
			tokenName = billData.CoinName
			cInfo, _ := base.FindCoinsByName(billData.CoinName)
			contractAddress = cInfo.Token
		}
		// TODO 发送给钱包,在这里写上链逻辑
		req := domain.BCWithDrawReq{
			ApiKey:          pInfo.ClientId,
			OutOrderId:      cbDao.SerialNo,
			CoinName:        chainName,
			Amount:          billData.Nums,
			ToAddress:       billData.TxToAddr,
			TokenName:       tokenName,
			ContractAddress: contractAddress,
			Memo:            billData.Memo,
		}
		log.Infof("发送给钱包：%v,业务线为空", req)
		err = blockChainsApi.BlockChainWithdrawCoin(req, Conf.BlockchainCustody.ClientId, Conf.BlockchainCustody.ApiSecret)
		if err != nil {
			fmt.Printf("\n WalletWithdraw req=%+v \n err = %v\n", req, err)
			log.Errorf("\n WalletWithdraw req=%+v \n err = %v\n", req, err)
			return err
		}

		// TODO 发送成功，更改链上订单表
		//upMap := map[string]interface{}{
		//	"is_wallet_deal": 1,
		//	"updated_at":     time.Now().Local(),
		//	"height":         0,
		//	"confirm_nums":   0,
		//}
		//err = UpdateChainBillByMap(chain.Id, upMap)
		//if err != nil {
		//	return false
		//}
	}
	return nil
}

// PushDataByUrl
// 推送数据给商户提供的接口
// http数据发送
// method post
// id 订单Id
func PushDataByUrl(str string) error {

	dao := new(bill.BillDetail)
	no, err := dao.FindBillDetailBySerialNo(str)
	if err != nil {
		return err
	}
	cf := ""
	tx := ""
	at := ""
	if !no.TxTime.IsZero() {
		tx = no.TxTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !no.AuditTime.IsZero() {
		at = no.AuditTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !no.ConfirmTime.IsZero() {
		cf = no.ConfirmTime.Format(global.YyyyMmDdHhMmSs)
	}
	statusName := dict.BillStateList[no.BillStatus]
	resultName := dict.OrderResult[no.OrderResult]
	billList := map[string]interface{}{
		"id":             no.Id,
		"txId":           no.TxId,
		"phone":          no.Phone,
		"serialNo":       no.SerialNo,
		"nums":           no.Nums,
		"fee":            no.Fee,
		"upChainFee":     no.UpChainFee,
		"destroyFee":     no.DestroyFee,
		"realNums":       no.RealNums,
		"resultName":     resultName,
		"coinName":       no.CoinName,
		"chainName":      no.ChainName,
		"serviceName":    no.ServiceName,
		"txTypeName":     dict.TxTypeNameList[no.TxType],
		"billStatusName": statusName,
		"txFromAddr":     no.TxFromAddr,
		"txToAddr":       no.TxToAddr,
		"remark":         no.Remark,
		"memo":           no.Memo,
		"txTime":         tx,
		"auditTime":      at,
		"confirmTime":    cf,
	}
	security := serviceSecurity.NewEntity()
	err = security.FindItemByBusinessId(int64(no.ServiceId))
	if err != nil {
		return err
	}
	if security == nil {
		return errors.New(fmt.Sprintf("该业务线%s,没有回调信息", no.ServiceName))
	}
	_, err = xkutils.PostJson(security.CallbackUrl, billList)
	if err != nil {
		return err
	}
	return nil
}
