package deals

import (
	conf "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/bill"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/unitUsdt"
	"custody-merchant-admin/model/userAddr"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/module/log"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
)

// UpdateBillDetail
// 更新订单
func UpdateBillDetail(info *domain.BillInfo) (int, error) {
	var billDao = new(bill.BillDetail)
	// 账单更新
	return billDao.UpdateBillDetail(info)
}

func PushBill(info *domain.BillInfo) error {
	var billDao = new(bill.BillDetail)
	// 查询该订单是否为处理中
	detail, err := billDao.GetBillByTxId(info.TxId)
	if err != nil {
		return err
	}
	fmt.Printf("%v", detail)
	return nil
}

func CountBillDetailList(info *domain.BillSelect) (int64, error) {
	dao := new(bill.BillDetail)
	return dao.CountBillDetailList(info)
}

func FindBillDetailList(info *domain.BillSelect) ([]domain.BillInfo, error) {
	var (
		billList = []domain.BillInfo{}
		billDao  = new(bill.BillDetail)
	)

	list, err := billDao.FindBillDetailList(info)
	if err != nil {
		return billList, err
	}

	for i := 0; i < len(list); i++ {
		cf := ""
		tx := ""
		at := ""
		if !list[i].TxTime.IsZero() {
			tx = list[i].TxTime.Format(global.YyyyMmDdHhMmSs)
		}
		if !list[i].AuditTime.IsZero() {
			at = list[i].AuditTime.Format(global.YyyyMmDdHhMmSs)
		}
		if !list[i].ConfirmTime.IsZero() {
			cf = list[i].ConfirmTime.Format(global.YyyyMmDdHhMmSs)
		}
		statusName := dict.BillStateList[list[i].BillStatus]
		resultName := dict.OrderResult[list[i].OrderResult]
		if list[i].BillStatus == 1 || list[i].BillStatus == 2 {
			resultName = dict.OrderResult[1]
		}
		billList = append(billList, domain.BillInfo{
			Id:             list[i].Id,
			TxId:           list[i].TxId,
			MerchantId:     list[i].MerchantId,
			Phone:          list[i].Phone,
			SerialNo:       list[i].SerialNo,
			CoinId:         list[i].CoinId,
			ChainId:        list[i].ChainId,
			ServiceId:      list[i].ServiceId,
			Nums:           list[i].Nums,
			Fee:            list[i].Fee,
			UpChainFee:     list[i].UpChainFee,
			BurnFee:        list[i].BurnFee,
			DestroyFee:     list[i].DestroyFee,
			RealNums:       list[i].RealNums,
			TxType:         list[i].TxType,
			BillStatus:     list[i].BillStatus,
			OrderResult:    list[i].OrderResult,
			ResultName:     resultName,
			CoinName:       list[i].CoinName,
			ChainName:      list[i].ChainName,
			ServiceName:    list[i].ServiceName,
			TxTypeName:     dict.TxTypeNameList[list[i].TxType],
			BillStatusName: statusName,
			TxFromAddr:     list[i].TxFromAddr,
			TxToAddr:       list[i].TxToAddr,
			Remark:         list[i].Remark,
			Memo:           list[i].Memo,
			State:          list[i].State,
			CreateByUser:   list[i].CreateByUser,
			ColorType:      dict.BillStateColors[statusName],
			ColorResult:    dict.OrderResultColor[resultName],
			TxTime:         tx,
			AuditTime:      at,
			ConfirmTime:    cf,
		})
	}
	return billList, err
}

func FindBillInfoBySerialNo(serialNo string) (domain.BillInfo, error) {
	var (
		billList = domain.BillInfo{}
		billDao  = new(bill.BillDetail)
	)

	ifo, err := billDao.FindBillDetailBySerialNo(serialNo)
	if err != nil {
		return billList, err
	}

	cf := ""
	tx := ""
	at := ""
	if !ifo.TxTime.IsZero() {
		tx = ifo.TxTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !ifo.AuditTime.IsZero() {
		at = ifo.AuditTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !ifo.ConfirmTime.IsZero() {
		cf = ifo.ConfirmTime.Format(global.YyyyMmDdHhMmSs)
	}
	statusName := dict.BillStateList[ifo.BillStatus]
	resultName := dict.OrderResult[ifo.OrderResult]
	billList = domain.BillInfo{
		Id:             ifo.Id,
		TxId:           ifo.TxId,
		MerchantId:     ifo.MerchantId,
		Phone:          ifo.Phone,
		SerialNo:       ifo.SerialNo,
		CoinId:         ifo.CoinId,
		ChainId:        ifo.ChainId,
		ServiceId:      ifo.ServiceId,
		Nums:           ifo.Nums,
		Fee:            ifo.Fee,
		UpChainFee:     ifo.UpChainFee,
		BurnFee:        ifo.BurnFee,
		DestroyFee:     ifo.DestroyFee,
		RealNums:       ifo.RealNums,
		TxType:         ifo.TxType,
		BillStatus:     ifo.BillStatus,
		OrderResult:    ifo.OrderResult,
		ResultName:     resultName,
		CoinName:       ifo.CoinName,
		ChainName:      ifo.ChainName,
		ServiceName:    ifo.ServiceName,
		TxTypeName:     dict.TxTypeNameList[ifo.TxType],
		BillStatusName: statusName,
		TxFromAddr:     ifo.TxFromAddr,
		TxToAddr:       ifo.TxToAddr,
		Remark:         ifo.Remark,
		Memo:           ifo.Memo,
		State:          ifo.State,
		CreateByUser:   ifo.CreateByUser,
		ColorType:      dict.BillStateColors[statusName],
		ColorResult:    dict.OrderResultColor[resultName],
		TxTime:         tx,
		AuditTime:      at,
		ConfirmTime:    cf,
	}
	return billList, err
}

func FindOutBillInfoBySerialNo(serialNo string) (map[string]interface{}, error) {
	var (
		billList = make(map[string]interface{}, 0)
		billDao  = new(bill.BillDetail)
	)

	ifo, err := billDao.FindBillDetailBySerialNo(serialNo)
	if err != nil {
		return billList, err
	}

	cf := ""
	tx := ""
	at := ""
	if !ifo.TxTime.IsZero() {
		tx = ifo.TxTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !ifo.AuditTime.IsZero() {
		at = ifo.AuditTime.Format(global.YyyyMmDdHhMmSs)
	}
	if !ifo.ConfirmTime.IsZero() {
		cf = ifo.ConfirmTime.Format(global.YyyyMmDdHhMmSs)
	}

	statusName := dict.BillStateList[ifo.BillStatus]
	resultName := dict.OrderResult[ifo.OrderResult]
	billList = map[string]interface{}{
		"id":             ifo.Id,
		"txId":           ifo.TxId,
		"phone":          ifo.Phone,
		"serialNo":       ifo.SerialNo,
		"nums":           ifo.Nums,
		"fee":            ifo.Fee,
		"upChainFee":     ifo.UpChainFee,
		"destroyFee":     ifo.DestroyFee,
		"realNums":       ifo.RealNums,
		"resultName":     resultName,
		"coinName":       ifo.CoinName,
		"chainName":      ifo.ChainName,
		"serviceName":    ifo.ServiceName,
		"txTypeName":     dict.TxTypeNameList[ifo.TxType],
		"billStatusName": statusName,
		"txFromAddr":     ifo.TxFromAddr,
		"txToAddr":       ifo.TxToAddr,
		"remark":         ifo.Remark,
		"memo":           ifo.Memo,
		"txTime":         tx,
		"auditTime":      at,
		"confirmTime":    cf,
	}
	return billList, err
}

func FindBillBalance(blist *domain.BillBalance, id int64) (*domain.BillBalance, error) {
	var (
		err  error
		ats  = new(assets.Assets)
		unit = new(unitUsdt.UnitUsdt)
		dao  = new(bill.BillDetail)
	)
	blist.BalanceList = []domain.BalanceList{}
	assets, err := ats.GetAssets(id)
	if err != nil {
		return blist, err
	}
	if assets == nil {
		return blist, err
	}
	rname := "cny"
	usdtInfo, err := unit.GetUnitUsdtById(blist.UnitId)
	if err != nil {
		return blist, err
	}
	if usdtInfo != nil {
		rname = usdtInfo.Name
	}
	coinPrice := dict.GetHooPriceByName("usdt", rname)
	// 提币，发送
	send, err := dao.FindBillByStatus(id, 4)
	if err != nil {
		return blist, err
	}
	// 接收
	receive, err := dao.FindBillByStatus(id, 1)
	if err != nil {
		return blist, err
	}
	endName := strings.ToUpper(rname)
	receivePrice := DealPrice(receive)
	blist.BalanceList = append(blist.BalanceList, domain.BalanceList{
		Title:    "receive",
		Icon:     "#icon-jieshou",
		UsdtNums: receivePrice,
		UnitNums: receivePrice.Mul(coinPrice).String() + " " + endName,
	})

	sendPrice := DealPrice(send)
	blist.BalanceList = append(blist.BalanceList, domain.BalanceList{
		Title:    "send",
		Icon:     "#icon-fasong",
		UsdtNums: sendPrice,
		UnitNums: sendPrice.Mul(coinPrice).String() + " " + endName,
	})

	total := decimal.Zero
	for i, _ := range assets {
		coin, err := base.FindCoinsById(assets[i].CoinId)
		if err != nil {
			return nil, err
		}
		tPrice := dict.GetHooPriceByName(coin.Name, "usd")
		numsPrice := assets[i].Nums.Mul(tPrice)
		freezeNums := assets[i].Freeze.Mul(tPrice)
		total = total.Add(numsPrice).Add(freezeNums)
	}

	blist.BalanceList = append(blist.BalanceList, domain.BalanceList{
		Title:    "total",
		Icon:     "#icon-shengyu",
		UsdtNums: total,
		UnitNums: total.Mul(coinPrice).String() + " " + endName,
	})
	return blist, nil
}
func DealPrice(bill []bill.BillNums) decimal.Decimal {

	var billPrice = decimal.Decimal{}
	for i, _ := range bill {
		scoin, err := base.FindCoinsById(bill[i].CoinId)
		if err != nil {
			return billPrice
		}
		sprice := dict.GetHooPriceByName(scoin.Name, "usd")
		price := bill[i].Nums.Mul(sprice)
		billPrice = billPrice.Add(price)
	}
	return billPrice.Round(6)
}

// GetWithdrawalFee
// 提现手续费处理
func GetWithdrawalFee(chainName, coinName string, chainId, coinId int) (decimal.Decimal, error) {
	var (
		fee   = decimal.Decimal{}
		cbDao = chainBill.NewEntity()
	)
	// 手续费
	rate := decimal.NewFromFloat(conf.Conf.Fee.Rate)
	// 查询前十条记录
	blst, err := cbDao.FindBillLimit(conf.Conf.Fee.Limit, chainId, coinId)
	if err != nil {
		return fee, err
	}

	n := 0
	// 查询前十条累加取平均值
	for _, bl := range blst {
		if !bl.UpChainFee.IsZero() {
			print(bl.UpChainFee.String())
			n += 1
			fee = fee.Add(bl.UpChainFee)
		}
	}

	// fee不为空
	if !fee.IsZero() {
		// 取fee平均值
		fee = fee.Div(decimal.NewFromInt(int64(n)))
		// 乘费率
		fee = fee.Mul(rate)
	}
	// 查询前十条记录为空
	if len(blst) == 0 || fee.IsZero() {
		// 是否使用交易所的收费规则
		//fee = decimal.NewFromFloat(0.4)
		if conf.Conf.Fee.Open {
			fee = dict.GetHooFee(chainName, coinName)
		}
	}
	return fee, err
}

// FindAddrAndSId
// 检查地址和业务线
func FindAddrAndSId(address string) (int, string, error) {

	var (
		userAddrDao = userAddr.NewEntity()
		chainsDao   = serviceChains.NewEntity()
		sid         = 0
		uId         = ""
	)

	err := chainsDao.GetMerchantChainsByAddr(address)
	if err != nil {
		log.Error(err.Error())
		return 0, "", err
	}
	if chainsDao.Id == 0 {
		err = userAddrDao.FindAddressByAddr(address)
		if err != nil {
			log.Error(err.Error())
			return 0, "", err
		}
		if userAddrDao.Id == 0 {
			log.Error("地址找不到业务线")
			return 0, "", errors.New("地址找不到业务线")
		}
		sid = int(userAddrDao.ServiceId)
		uId = userAddrDao.MerchantUser
	} else {
		sid = chainsDao.ServiceId
		uId = fmt.Sprintf("%d", chainsDao.MerchantId)
	}
	return sid, uId, err
}

// CheckAssetNum
// 检查资产数量是否够扣除手续费
func CheckAssetNum(sid, coinId int, chainName string, nums, fee decimal.Decimal) error {

	bs := serviceChains.NewEntity()
	err := bs.FindServiceChainsInfo(sid, chainName)
	if err != nil {
		return err
	}
	asset := assets.NewEntity()
	ass, err := asset.GetAssetsNumsBySIdAndCId(sid, coinId)
	if err != nil {
		return err
	}
	if ass.Nums.LessThan(nums) {
		return errors.New("业务线无法转账：金额不足")
	}
	// 2.2 判断主链币够不够扣手续费
	coin, err := base.FindCoinsByName(chainName)
	if err != nil {
		return err
	}
	chainAsset, err := asset.GetDateAssetsBySIdAndCId(sid, int(coin.Id))
	if err != nil {
		return err
	}
	if chainAsset.Nums.LessThan(fee) {
		return errors.New("主链币余额不足")
	}
	return nil
}
