package service

import (
	"bytes"
	"custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/domain/dto"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/model/business"
	"custody-merchant-admin/model/businessPackage"
	"custody-merchant-admin/model/comboUse"
	"custody-merchant-admin/model/order"
	"custody-merchant-admin/model/orm"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/model/white"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
	"strconv"
	"time"
)

func CreateWithdrawBill(info *domain.BillInfo, chainName, coinName string) (string, error) {
	var (
		isAudit = false
		//inAddr          = false
		withdrawalState = true
		serviceDao      = business.NewEntity()
		//uDao            = userAddr.NewEntity()
		security = serviceSecurity.NewEntity()
	)
	// 平台审核判断
	err := security.FindItemByBusinessId(int64(info.ServiceId))
	if err != nil {
		return "", err
	}
	// 查询白名单限制
	addrs, err := white.GetWhiteListUse(info.ServiceId, info.CoinId, info.TxToAddr)
	if err != nil {
		return "", err
	}

	// 是允许进行外部提币
	// security.IsWithdrawal = 0 关闭了白名单，无需限制直接出帐
	// security.IsWithdrawal = 1 开启了白名单，需1、限判断白名单地址是否允许出帐，2、需要进行是否需要人工审核
	//if security.IsWithdrawal != 0 {
	//	// 开启提币限制 security.IsWithdrawal = 1
	//	err = uDao.FindAddressByAddr(info.TxToAddr)
	//	if err != nil {
	//		return "", err
	//	}
	//	if uDao.Id != 0 {
	//		inAddr = true
	//		info.ToId = fmt.Sprintf("%d", serviceDao.AccountId)
	//	} else {
	//		bl := config.Conf.Blockchain
	//		// 判断内外部地址
	//		inAddr, err = blockChainsApi.ValidInsideAddress(bl.ClientId, bl.ApiSecret, info.TxToAddr)
	//		if err != nil {
	//			return "", err
	//		}
	//	}
	//	// 外部地址
	//	if !inAddr {
	//		// 白名单地址
	//		if addrs == nil || addrs.Use == 0 {
	//			return "", errors.New("请加入白名单")
	//		}
	//	}
	//
	//	if addrs == nil || addrs.Use == 0 {
	//		return "", errors.New("请加入白名单")
	//	}
	//}

	// 提币手续费
	withdrawalFee, err := deals.GetWithdrawalFee(chainName, coinName, info.ChainId, info.CoinId)
	if err != nil {
		return "", err
	}
	info.Fee = withdrawalFee
	err = serviceDao.FindBusinessItemById(int64(info.ServiceId))
	if err != nil {
		return "", err
	}
	if serviceDao == nil || serviceDao.Id == 0 || serviceDao.State == 2 {
		return "", errors.New("没有业务线")
	}
	if serviceDao.State == 1 {
		return "", errors.New("业务线被冻结")
	}
	info.MerchantId = serviceDao.AccountId
	info.Phone = serviceDao.Phone
	serialNo := xkutils.Generate("HF", time.Now())
	info.SerialNo = serialNo
	info.TxType = 3
	msg := ""

	// 1. 判断提币限制，判断限制提币数额，是否关闭提币
	if serviceDao.WithdrawalStatus == 1 {
		// 0 开启提币，1 关闭提币
		withdrawalState = false
		msg = "该业务线已关闭提币，请先开启"
	}

	if serviceDao.LimitTransfer == 1 {
		// 0关闭限制转帐,1开启限制转帐，
		err = deals.TransferLimit(info.ServiceId, info.TxToAddr, info.Nums)
		if err != nil {
			log.Error(err.Error())
			withdrawalState = false
			msg = err.Error()
		}
	}
	if serviceDao.LimitSameWithdrawal == 1 {
		// 0 开启限制同地址提币，1 关闭限制同地址提币
		err = deals.WithdrawalLimit(info.ServiceId, info.TxToAddr, info.Nums)
		if err != nil {
			log.Error(err.Error())
			withdrawalState = false
			msg = err.Error()
		}
	}

	// 1. 可以提币，创建提币订单审核
	if withdrawalState {
		// 2. 创建账单
		// 2.1 判断 主链币/代币 够不够出,主链币够不够扣除手续费
		err = deals.CheckAssetNum(info.ServiceId, info.CoinId, chainName, info.Nums, info.Fee)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}

		info.TxType = 3
		info.BillStatus = 3
		// security.IsWithdrawal == 0 关闭白名单
		// 需要平台或者商家自行审核
		if security.IsWhitelist != 0 {
			if addrs != nil && addrs.Use == 0 {
				// 直接出帐
				log.Infof("%d,%s security.IsWithdrawal == 0 关闭白名单,需要平台或者商家自行审核", security.BusinessId, serialNo)
				isAudit = false
			} else {
				return "", errors.New(fmt.Sprintf("地址：%s,未开启白名单", info.TxToAddr))
			}
		}
		// 关闭了白名单，需要平台或者商家审核
		if security.IsWhitelist == 0 && (security.IsAccountCheck == 1 || security.IsPlatformCheck == 1) {
			log.Infof("%d,%s 关闭了白名单，需要平台或者商家审核", security.BusinessId, serialNo)
			isAudit = true
		}
		// 关闭了白名单，不需要平台或者商家审核
		if security.IsWhitelist == 0 && security.IsAccountCheck == 0 && security.IsPlatformCheck == 0 {
			// 直接出帐
			log.Infof("%d,%s 关闭了白名单，不需要平台或者商家审核", security.BusinessId, serialNo)
			isAudit = false
		}
		err = CreateBill(info, isAudit)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}
		if !isAudit {
			// 发起上链
			log.Infof("%d,%s 发起上链", security.BusinessId, serialNo)
			err = SendBillOutMsg(serialNo)
			if err != nil {
				return "", err
			}
		}
	} else {
		// 提币订单限制不通过
		return "", errors.New(msg)
	}
	return info.SerialNo, nil
}

func CreateBill(info *domain.BillInfo, isAudit bool) error {

	var billDao = new(bill.BillDetail)
	db := model.DB().Begin()
	_, err := billDao.CreateBillDetail(db, info)
	if err != nil {
		db.Rollback()
		return err
	}

	if isAudit {
		// 创建提现进度审核
		orders := &domain.OrderInfo{
			CoinId:      info.CoinId,
			ChainId:     info.ChainId,
			ServiceId:   info.ServiceId,
			Type:        0, // 提现
			SerialNo:    info.SerialNo,
			TxId:        info.TxId,
			Memo:        info.Remark,
			ReceiveAddr: info.TxToAddr,
			FromAddr:    info.TxFromAddr,
			Nums:        info.Nums,
			Fee:         info.Fee,
			BurnFee:     info.BurnFee,
			DestroyFee:  info.DestroyFee,
			RealNums:    info.RealNums,
			MerchantId:  info.MerchantId,
			Phone:       info.Phone,
		}
		err := CreateOrderInfo(orders, info.CreateByUser)
		if err != nil {
			db.Rollback()
			return err
		}
	}
	db.Commit()
	// 商户资产冻结和增加
	// 先查找主链币的代币ID
	var chainName string
	if info.ChainId != 0 {
		chain, err1 := base.FindChainsById(info.ChainId)
		if err1 != nil {
			return err
		}
		if chain.Name != "" {
			chainName = chain.Name
		}
	} else {
		chainName = info.CoinName
	}
	if chainName == "" {
		return errors.New("业务线主链币为空")
	}

	coin, err := base.FindCoinsByName(chainName)
	if err != nil {
		return err
	}
	bs := serviceChains.NewEntity()
	err = bs.FindServiceChainsInfo(info.ServiceId, chainName)
	if err != nil {
		return err
	}
	// 先扣除手续费
	err = deals.WithdrawalFreezeAssets(info.ServiceId, int(coin.Id), decimal.Zero, info.Fee)
	if err != nil {
		return err
	}
	// 再冻结代币
	err = deals.WithdrawalFreezeAssets(info.ServiceId, info.CoinId, info.Nums, decimal.Zero)
	if err != nil {
		return err
	}

	return nil
}
func UpdateConfirmNums(mqData domain.MqWalletInfo) error {
	billDao := new(bill.BillDetail)
	od := new(order.Orders)
	chainBill, err := FindChainBillChainSerialNo(mqData.SerialNo)
	if err != nil {
		return err
	}
	if chainBill.Id == 0 {
		return errors.New(fmt.Sprintf("%s链上账单为空", chainBill.SerialNo))
	}
	err = chainBill.UpdatesChainBill(chainBill.Id, map[string]interface{}{
		"cold_wallet_state": 1,
		"is_cold_wallet":    1,
		"is_wallet_deal":    1,
		"tx_id":             mqData.TxId,
		"confirm_nums":      mqData.ConfirmNums,
		"height":            mqData.Height,
		"updated_at":        time.Now().Local(),
	})
	if err != nil {
		return err
	}
	err = billDao.UpdateBillBySerialNo(mqData.SerialNo, map[string]interface{}{
		"tx_id":        mqData.TxId,
		"real_nums":    mqData.RealNums,
		"up_chain_fee": mqData.MinerFee,
		"updated_at":   time.Now().Local(),
	})
	err = od.UpdateOrdersInfoBySerialNo(mqData.SerialNo, map[string]interface{}{
		"tx_id":        mqData.TxId,
		"up_chain_fee": mqData.MinerFee,
		"update_time":  time.Now().Local(),
	})
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return err
}

// FreezeBillDetailState
// 提币确认更新
func FreezeBillDetailState(mqData domain.MqWalletInfo, state int) error {
	//判断txid 是否已经处理
	key := fmt.Sprintf("%s-custody:out-txid:%v", config.Conf.Mod, mqData.TxId)
	var v string
	err := cache.GetRedisClientConn().Get(key, &v)
	if v != "" {
		//txid 已经处理
		log.Error("交易TXID已经处理3\n")
		return nil
	}
	//判断流水记录有没有处理
	billDao := new(bill.BillDetail)
	detail, err := billDao.GetBillBySerialNo(mqData.SerialNo)
	if err != nil {
		return err
	}
	if detail == nil || detail.Id == 0 {
		return errors.New("没有账单")
	}
	if detail.BillStatus != 3 {
		return errors.New(fmt.Sprintf("%s已经出过帐", detail.SerialNo))
	}
	if detail.TxId == "" {
		err = UpdateConfirmNums(mqData)
		if err != nil {
			return err
		}
	}
	cache.GetRedisClientConn().Set(key, "processing", 5*time.Minute) //处理中
	// 1. 根据地址查询业务线和币种
	chainBill, err := FindChainBillChainSerialNo(mqData.SerialNo)
	if err != nil {
		return err
	}
	if chainBill.Id == 0 {
		return errors.New(fmt.Sprintf("%s链上账单为空", chainBill.SerialNo))
	}
	chains, err := base.FindChainsById(chainBill.ChainId)
	if err != nil {
		return err
	}
	bs := serviceChains.NewEntity()
	err = bs.FindServiceChainsInfo(chainBill.ServiceId, chains.Name)
	if err != nil {
		return err
	}
	asset, err := GetDateAssetsBySIdAndCId(chainBill.ServiceId, bs.CoinId)
	if err != nil {
		return err
	}
	log.Errorf("GetDateAssetsBySIdAndCId assetinfo %+v\n", asset)
	// 提现确认
	if state == 1 && detail.BillStatus == 3 {
		// 更新账单
		err = billDao.UpdateBillBySerialNo(mqData.SerialNo, map[string]interface{}{
			"bill_status":  4,
			"tx_type":      4,
			"up_chain_fee": mqData.MinerFee,
			"real_nums":    mqData.RealNums,
			"confirm_time": time.Now().Local(),
			"updated_at":   time.Now().Local(),
		})
		// 更新资产
		err = deals.WithdrawalConfirmAssets(detail.ServiceId, detail.CoinId, detail.Nums)
		if err != nil {
			return err
		}
		detail, err = billDao.GetBillByTxId(mqData.TxId)
		if err != nil {
			return err
		}
		detail.UpChainFee = mqData.MinerFee
		detail.RealNums = mqData.RealNums
		detail.BillStatus = 4
		detail.TxType = 4
		cbDao := dto.BillDetailToChainBill(detail)
		cbDao.Height = mqData.Height
		cbDao.ConfirmNums = mqData.ConfirmNums
		cbDao.ColdWalletResult = 1
		cbDao.IsColdWallet = 1
		cbDao.ColdWalletState = 1
		// 新增或者修改收益户信息
		err = SaveIncome(cbDao, 0)
		if err != nil {
			return err
		}
		if chainBill.Id != 0 {
			upMap := map[string]interface{}{
				"cold_wallet_state":  1,
				"cold_wallet_result": 2,
				"bill_status":        cbDao.BillStatus,
				"tx_type":            cbDao.TxType,
				"chain_id":           detail.ChainId,
				"updated_at":         time.Now().Local(),
				"up_chain_fee":       mqData.MinerFee,
				"destroy_fee":        mqData.Destroy,
				"height":             mqData.Height,
				"confirm_nums":       mqData.ConfirmNums,
				"memo":               mqData.Memo,
			}
			// 更新链上订单
			err = UpdateChainBillByMap(chainBill.Id, upMap)
			if err != nil {
				return err
			}
			// 提现，增加财务流水
			// 财务流水-主链币
			db := orm.Cache(model.DB())
			_, err = SyncToFinanceAssetsByWithdraw(db,
				int64(chainBill.ChainId), billDao.Fee,
				chainBill.SerialNo, int64(bs.ServiceId))
			if err != nil {
				return err
			}
		}
	}
	cache.GetRedisClientConn().Set(key, "success", 5*time.Minute) //处理中
	return nil
}

// ReceiveBillDetail
// 接收币种
func ReceiveBillDetail(mqData domain.MqWalletInfo) error {
	key := fmt.Sprintf("%s-custody:in-txid:%v", config.Conf.Mod, mqData.TxId)
	cache.GetRedisClientConn().Set(key, "processing", 5*time.Second) //处理中

	log.Errorf("ReceiveBillDetail mqData %+v\n", mqData)
	var (
		billDao    = new(bill.BillDetail)
		serviceDao = business.NewEntity()
		//merchantDao = merchant.NewEntity()
	)
	log.Errorf("mqData.ToAddress:%s", mqData.ToAddress)
	// 1. 根据地址查询业务线和币种
	_, fromId, _ := deals.FindAddrAndSId(mqData.FromAddress)
	sid, toId, err := deals.FindAddrAndSId(mqData.ToAddress)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	billDao.ServiceId = sid
	billDao.FromId = fromId
	billDao.ToId = toId
	// 2. 根据业务线查询商户信息
	err = serviceDao.FindBusinessItemById(int64(billDao.ServiceId))
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	if serviceDao.AccountId == 0 {
		log.Error("接收币种,业务线没有商户")
		return errors.New("接收币种,业务线没有商户")
	}
	coin, err := base.FindCoinsByName(mqData.CoinName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	var chainName string
	if coin.ChainId != 0 {
		chain, err := base.FindChainsById(coin.ChainId)
		if err != nil {
			log.Error("接收币种," + err.Error())
			return err
		}
		if chain.Name != "" {
			chainName = chain.Name
		}
	} else {
		chainName = coin.Name
	}
	if chainName == "" {
		return errors.New("接收币种,业务线主链币为空")
	}
	serialNo := xkutils.Generate("HF", time.Now())
	// 3. 整理新增的参数
	billDao.SerialNo = serialNo
	billDao.BillStatus = 1
	// 钱包传回的参数
	billDao.TxId = mqData.TxId
	billDao.Nums = mqData.Nums
	// 充值手续费用
	billDao.Fee = decimal.Zero
	// 记录
	// 添加USDT使用量
	err = UpdateUseComboByType(int64(sid), mqData)
	if err != nil {
		return err
	}
	log.Errorf("ReceiveBillDetail 添加使用量\n")

	// 新增订单
	billDao.DestroyFee = mqData.Destroy
	billDao.BurnFee = mqData.BurnFee
	billDao.UpChainFee = mqData.MinerFee
	billDao.RealNums = mqData.RealNums      // 实际到账
	billDao.Memo = mqData.Memo              // 备注
	billDao.TxFromAddr = mqData.FromAddress // 发送地址
	billDao.TxToAddr = mqData.ToAddress     // 接收地址
	// 托管这边查询的参数
	billDao.MerchantId = serviceDao.AccountId
	billDao.Phone = serviceDao.Phone
	billDao.CoinId = int(coin.Id)
	billDao.ChainId = coin.ChainId
	billDao.TxType = 1                       // 链上确认
	billDao.ConfirmTime = mqData.ConfirmTime // 链上确认时间
	billDao.TxTime = time.Now().Local()      // 交易时间
	billDao.AuditTime = time.Now().Local()   // 审核时间
	billDao.WithdrawalFee = decimal.Zero     // 提现费用
	billDao.TopUpFee = decimal.Zero          // 充值费用
	billDao.Remark = ""
	billDao.CreateByUser = 0
	//  新增订单
	orderNo, err := billDao.InsertBillDetail()
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	fmt.Printf("账单ID:%s", orderNo)

	// TODO 4. 商户资产冻结
	// 4.1 商户资产主链币手续费冻结
	// 第一期不涉及充值收取手续费，暂时注释
	//coinId, err := base.FindCoinsByName(chainName)
	//if err != nil {
	//	return err
	//}

	// TODO 查询商户的链路地址，用于判断链路地址资产
	// 第一期不涉及充值收取手续费，暂时注释
	//bs := serviceChains.NewEntity()
	//err = bs.FindServiceChainsInfo(billDao.ServiceId, chainName)
	//if err != nil {
	//	return err
	//}

	// TODO 冻结主链币手续费
	// 收费是手续费
	//err = deals.ReceiveFreezeAssets(billDao.ServiceId, int(coinId.Id), decimal.Zero, billDao.Fee, bs.ChainAddr)
	//if err != nil {
	//	return err
	//}

	// TODO 冻结币种
	// 增加充值费用
	err = deals.ReceiveFreezeAssets(billDao.ServiceId, billDao.CoinId, billDao.Nums, decimal.Zero)
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	// TODO 商户资产解冻和增加
	// 4.1 已经接收
	err = deals.ReceiveConfirmAssets(billDao.ServiceId, billDao.CoinId, billDao.RealNums)
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	log.Errorf("ReceiveBillDetail 新增链上订单\n")
	// 5. 新增链上订单
	cbDao := dto.BillDetailToChainBill(billDao)
	cbDao.Height = mqData.Height
	cbDao.ConfirmNums = mqData.ConfirmNums
	cbDao.ColdWalletResult = 2
	cbDao.IsColdWallet = 1
	cbDao.ColdWalletState = 1
	err = CreateChainBill(cbDao)
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	// 6. 新增收益户
	err = SaveIncome(cbDao, 1)
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}

	// TODO 7. 添加充值的流水
	// 取商户的链路地址进行手续费扣减记录
	// 第一期不涉及充值收取手续费，暂时注释
	//asset, err := deals.FindAssetBySIdAndCId(billDao.ServiceId, int(coinId.Id), bs.ChainAddr)
	//if err != nil {
	//	return err
	//}
	//or := orm.Cache(model.DB())
	//_, err = SyncToFinanceAssetsByRecharge(or, coinId.Id, billDao.Fee, orderNo, asset.ChainAddress)
	//if err != nil {
	//	return err
	//}

	cache.GetRedisClientConn().Set(key, "ssuccess", 5*time.Minute) //处理成功

	return nil
}

// UpdateReceiveBillDetail
// 接收更新
func UpdateReceiveBillDetail(txId string, state int) error {
	var (
		billDao = new(bill.BillDetail)
	)
	if txId == "" {
		return errors.New(fmt.Sprintf("txId:%s", txId))
	}
	bil, err := billDao.GetBillByTxId(txId)
	if err != nil {
		return err
	}
	if bil == nil || bil.Id == 0 {
		return errors.New(fmt.Sprintf("%s找不到账单", txId))
	}
	// 账单更新 - 0
	bmp := map[string]interface{}{
		"bill_status":  state,
		"tx_type":      state,
		"confirm_time": time.Now().Local(),
		"updated_at":   time.Now().Local(),
	}
	// 更新账单
	_, err = billDao.UpdateBill(bil.TxId, bmp)
	if err != nil {
		return err
	}
	// 4. 商户资产冻结和增加
	// 4.1 已经接收
	err = deals.ReceiveConfirmAssets(bil.ServiceId, bil.CoinId, bil.RealNums)
	if err != nil {
		return err
	}
	// 5. 新增链上订单
	incomeDao := dto.BillDetailToChainBill(bil)
	err = CreateChainBill(incomeDao)
	if err != nil {
		return err
	}
	// 6. 新增收益户
	err = SaveIncome(incomeDao, 1)
	if err != nil {
		return err
	}
	return nil
}

func FindBillService(info *domain.BillSelect) ([]domain.BillInfo, int64, error) {
	list, err := deals.FindBillDetailList(info)
	if err != nil {
		return nil, 0, err
	}
	count, err := deals.CountBillDetailList(info)
	if err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func FindBillBalanceService(info *domain.BillBalance, id int64) (*domain.BillBalance, error) {
	balance, err := deals.FindBillBalance(info, id)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func FindBillInfoBySerialNo(serialNo string) (domain.BillInfo, error) {
	return deals.FindBillInfoBySerialNo(serialNo)
}
func FindOutBillInfoBySerialNo(serialNo string) (map[string]interface{}, error) {
	return deals.FindOutBillInfoBySerialNo(serialNo)
}

func ExportBillService(info *domain.BillSelect) (bytes.Buffer, error) {
	info.Limit = 99999

	bl, err := deals.FindBillDetailList(info)
	if err != nil {
		return bytes.Buffer{}, err
	}

	xFile := xlsx.NewFile()
	sheet, err := xFile.AddSheet("Sheet1")
	if err != nil {
		return bytes.Buffer{}, err
	}
	info.Title = []string{"序号", "账单ID", "业务线ID", "业务线名称", "商户ID", "手机号", "主链币", "代币", "数量", "手续费", "矿工费", "销毁数量",
		"实际到账数量", "账单类型", "账单状态", "发送地址", "接收地址", "MEMO", "交易时间", "审核时间", "确认时间", "备注"}
	r := sheet.AddRow()
	var ce *xlsx.Cell
	for _, v := range info.Title {
		ce = r.AddCell()
		ce.Value = v
	}
	for i := 0; i < len(bl); i++ {
		r = sheet.AddRow()
		// 序号
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(bl[i].Id, 10)
		// 账单ID
		ce = r.AddCell()
		ce.Value = bl[i].SerialNo
		// 业务线ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bl[i].ServiceId)
		// 业务线
		ce = r.AddCell()
		ce.Value = bl[i].ServiceName
		// 商户ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bl[i].MerchantId)
		// 手机号
		ce = r.AddCell()
		ce.Value = bl[i].Phone
		// 主链币
		ce = r.AddCell()
		ce.Value = bl[i].ChainName
		// 代币
		ce = r.AddCell()
		ce.Value = bl[i].CoinName
		// 数量
		ce = r.AddCell()
		ce.Value = bl[i].Nums.String()
		// 手续费
		ce = r.AddCell()
		ce.Value = bl[i].Fee.String()
		// 矿工费
		ce = r.AddCell()
		ce.Value = bl[i].UpChainFee.String()
		// 销毁费
		ce = r.AddCell()
		ce.Value = bl[i].DestroyFee.String()
		// 实际到账
		ce = r.AddCell()
		ce.Value = bl[i].RealNums.String()
		// 账单类型
		ce = r.AddCell()
		ce.Value = dict.TxTypeNameList[bl[i].TxType]
		// 账单状态
		ce = r.AddCell()
		ce.Value = dict.BillStateList[bl[i].BillStatus]
		// 发送地址
		ce = r.AddCell()
		ce.Value = bl[i].TxFromAddr
		// 接收地址
		ce = r.AddCell()
		ce.Value = bl[i].TxToAddr
		//// TXID
		//ce = r.AddCell()
		//ce.Value = bill[i].TxId
		// Memo
		ce = r.AddCell()
		ce.Value = bl[i].Memo
		// 交易时间
		ce = r.AddCell()
		ce.Value = bl[i].TxTime
		// 审核时间
		ce = r.AddCell()
		ce.Value = bl[i].AuditTime
		// 确认时间
		ce = r.AddCell()
		ce.Value = bl[i].ConfirmTime
		// 备注
		ce = r.AddCell()
		ce.Value = bl[i].Remark
	}
	//将数据存入buff中
	var buff bytes.Buffer
	if err = xFile.Write(&buff); err != nil {
		return bytes.Buffer{}, err
	}
	return buff, nil
}

func PushBillService(info *domain.BillInfo) error {
	return deals.PushBill(info)
}

func UpdateBillDetailService(info *domain.BillInfo) (int, error) {
	return deals.UpdateBillDetail(info)
}

func RollbackBillAssets(billData *domain.BillInfo) error {
	err := deals.RollbackBillAssets(billData.SerialNo)
	if err != nil {
		fmt.Println(err.Error())
		log.Error(err.Error())
		return err
	}
	// TODO 由于链上已经确认才有流水，所以不会有流水回滚
	// 财务流水回滚
	//ffInfo := financeFlow.NewEntity()
	//err = ffInfo.FindItemByOrderId(billData.SerialNo)
	//if err != nil {
	//	return err
	//}
	//err = RollbackFinanceAssetsByFlowId(ffInfo.Db, int(ffInfo.Id))
	return nil
}

func FindBillByTxId(txid string) (*bill.BillDetail, error) {
	var billDao = new(bill.BillDetail)
	return billDao.GetBillByTxId(txid)
}

func FindBillByOrderId(orderId string) (*bill.BillDetail, error) {
	var billDao = new(bill.BillDetail)
	return billDao.GetBillBySerialNo(orderId)
}

func ComboUsd(aId, packageId int64, usedLine decimal.Decimal) error {
	comboUseDao := comboUse.NewEntity()
comboLoop:
	comboInfo, err := comboUseDao.FindComboUserDayByCId(packageId, aId, time.Now().Local().Format(global.YyyyMmDd))
	if err != nil {
		return err
	}
	if comboInfo != nil {
		up, err := comboUseDao.UpDateComboUserDayByCId(packageId, aId,
			time.Now().Local().Format(global.YyyyMmDd), comboInfo.Version,
			map[string]interface{}{
				"used_line_day": usedLine,
				"version":       comboInfo.Version + 1,
			})
		if err != nil {
			return err
		}
		if up == 0 {
			goto comboLoop
		}
	} else {
		comboUseDao.CreateComboUserDay(comboUse.Entity{
			ComboUserId: packageId,
			UsedAddrDay: 0,
			UsedLineDay: usedLine,
			CreateTime:  time.Now().Local(),
		})
	}
	return nil
}

func UpdateUseComboByType(sid int64, mqData domain.MqWalletInfo) error {

	bpDao := businessPackage.NewEntity()
	coins, err := base.FindCoinsByName(mqData.CoinName)
	if err != nil {
		return err
	}
	err = bpDao.FindBPItemByBusinessId(sid)
	if err != nil {
		log.Error("接收币种," + err.Error())
		return err
	}
	if bpDao.Id == 0 {
		log.Error("没有套餐，请联系管理员")
		return errors.New("没有套餐，请联系管理员")
	}
	usedtNums := coins.PriceUsd.Mul(mqData.RealNums)
	hadUsed := bpDao.HadUsed
	log.Errorf("ReceiveBillDetail loopUsed up, %v, %v, %v\n", sid, bpDao.HadUsed, usedtNums)
	if bpDao.TypeName == "地址收费套餐" {
		log.Info("地址收费套餐 无需更新流量使用")
		return nil
	} else {
		// 记录
		// 今日使用量
		err = ComboUsd(bpDao.AccountId, bpDao.PackageId, hadUsed.Add(usedtNums))
		if err != nil {
			log.Error("接收币种," + err.Error())
			return err
		}
		_, err = bpDao.UpdateBPItemUsdBySIdByMap(sid, map[string]interface{}{
			"had_used": hadUsed.Add(usedtNums),
		})
		if err != nil {
			log.Error("接收币种," + err.Error())
			return err
		}
	}
	return nil
}
