package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/business"
	"custody-merchant-admin/model/businessPackage"
	_package "custody-merchant-admin/model/package"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
)

// CreatePackageItem 创建套餐
func CreatePackageItem(user *domain.JwtCustomClaims, req *domain.PackageInfo) (err error) {

	b, errB := json.Marshal(req)
	if errB != nil {
		return global.OperationErrorText(errB.Error())
	}

	rInfo := map[string]interface{}{}
	err = json.Unmarshal(b, &rInfo)
	if err != nil {
		return global.OperationErrorText(err.Error())
	}
	var isValid bool
	isValid, err = determainPackageParametersIsValid(rInfo)
	if !isValid {
		return err
	}

	pInfo := _package.NewEntity()
	err = json.Unmarshal(b, pInfo)
	if err != nil {
		return global.OperationErrorText(err.Error())
	}
	pInfo.Id = 0
	err = pInfo.InsertNewPackage()
	return
}

// DeletePackageItem 删除套餐
func DeletePackageItem(userId, packageId int64) (err error) {
	//查询id是否存在
	pInfo := _package.NewEntity()
	err = pInfo.FindPackageItemById(packageId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return global.DaoError(err)
	}
	err = pInfo.DeletePackageItem(packageId)
	if err != nil {
		return global.DaoError(err)
	}
	return
}

// UpdatePackageItem 更新套餐
func UpdatePackageItem(user *domain.JwtCustomClaims, req map[string]interface{}) (err error) {

	//查询id是否存在
	pInfo := _package.NewEntity()
	id := GetIntFromInterface(req["id"])
	err = pInfo.FindPackageItemById(id)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	var b []byte
	b, err = json.Marshal(req)
	if err != nil {
		log.Errorf("json Marshal err :%v", err.Error())
		return global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	err = json.Unmarshal(b, pInfo)
	if err != nil {
		log.Errorf("json Marshal err :%v", err.Error())
		return global.WarnMsgError(global.DataWarnUpdateDataErr)
	}
	var isValid bool
	isValid, err = determainPackageParametersIsValid(req)
	if !isValid {
		return err
	}

	err = pInfo.UpdatePackageItem(id, *pInfo)

	return
}

// SearchPackageItem 搜索套餐列表
func SearchPackageItem(req *domain.PackageReqInfo) (info domain.PackageInfo, err error) {
	pInfo := _package.NewEntity()
	if req.Id != 0 {
		err = pInfo.FindPackageItemById(req.Id)
	} else if req.TypeName != "" && req.ModelName != "" {
		err = pInfo.FindPackageByTypeModel(req.TypeName, req.ModelName)
	} else {
		return info, global.WarnMsgError(global.DataWarnParamErr)
	}
	if err == gorm.ErrRecordNotFound {
		return info, global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if err != nil {
		return info, err
	}
	//查询id是否存在

	b, errB := json.Marshal(pInfo)
	if errB != nil {
		return info, global.OperationErrorText(errB.Error())
	}

	errC := json.Unmarshal(b, &info)
	if errC != nil {
		return info, global.OperationErrorText(errC.Error())
	}

	return info, nil
}

func SearchMchPackageItem(req *domain.MchPackageReqInfo) (info domain.MchPackageInfo, err error) {
	bpInfo := businessPackage.NewEntity()
	var list []businessPackage.MchPackageDB
	err = bpInfo.FindBPItemByUserId(req.AccountId)
	if err != nil {
		log.Errorf("FindBPItemByUserId err %+v\n", err)
		log.Errorf("FindBPItemByUserId req %+v\n", req)
		return domain.MchPackageInfo{}, err
	}
	var searchPId int64
	if req.PackageId != 0 {
		searchPId = req.PackageId
	} else {
		searchPId = bpInfo.PackageId
	}
	if searchPId != 0 && req.AccountId != 0 {
		list, err = bpInfo.FindMchBusinessByPackageId(searchPId, req.AccountId)
		if err != nil {
			log.Errorf("FindMchBusinessByPackageId err %+v\n", err)
			log.Errorf("FindMchBusinessByPackageId req %+v\n", req)
		}
	}
	ChainName := make([]string, 0)
	SubCoinName := make([]string, 0)
	businessName := make([]string, 0)
	deductCoin := make([]string, 0)
	coinArr := make([]string, 0)
	subCoinArr := make([]string, 0)
	Fee := make([]domain.MchFeeInfo, 0)
	var (
		TypeName     string
		ModelName    string
		BusinessName string
		TotalCost    decimal.Decimal
	)
	for _, item := range list {
		TypeName = item.TypeName
		ModelName = item.ModelName
		businessName = append(businessName, item.Name)
		//计算优惠费
		//是否增加主链
		var moreCoinFee decimal.Decimal
		var moreSubCoinFee decimal.Decimal
		if item.Coin != "" {
			coinArr = strings.Split(item.Coin, ",")
			if len(coinArr) > item.ChainNums {
				m := len(coinArr) - item.ChainNums
				moreCoinFee = item.ChainDiscountNums.Mul(decimal.NewFromInt(int64(m)))
			}
		}

		if item.SubCoin != "" {
			subCoinArr = strings.Split(item.SubCoin, ",")
			if len(subCoinArr) > item.CoinNums {
				m := len(subCoinArr) - item.CoinNums
				moreSubCoinFee = item.CoinDiscountNums.Mul(decimal.NewFromInt(int64(m)))
			}
		}

		DiscountFee := moreCoinFee.Add(moreSubCoinFee)
		ChainName = append(ChainName, coinArr...)
		SubCoinName = append(SubCoinName, subCoinArr...)
		if item.DeductCoin != "" {
			deductCoinArr := strings.Split(item.DeductCoin, ",")
			deductCoin = append(deductCoin, deductCoinArr...)
		}
		fee := domain.MchFeeInfo{
			ChainDiscountUnit: item.ChainDiscountUnit,
			ChainDiscountNums: item.ChainDiscountNums,
			CoinDiscountUnit:  item.CoinDiscountUnit,
			CoinDiscountNums:  item.CoinDiscountNums,
			DeployFee:         item.DeployFee,
			CoverFee:          item.CoverFee,
			DiscountFee:       DiscountFee,
			MinerFee:          "以链上为准",
			DepositFee:        item.DepositFee,
			AddrNums:          item.AddrNums,
		}
		Fee = append(Fee, fee)
	}
	BusinessName = strings.Join(businessName, ";")
	//总费用未计算,差额为计算

	//获取当前业务线-套餐

	bInfo := business.NewEntity()
	err = bInfo.FindBusinessItemById(bpInfo.BusinessId)
	if err != nil {
		log.Errorf("FindBusinessItemById err %+v\n", err)
		log.Errorf("FindBusinessItemById bpInfo %+v\n", bpInfo)
	}
	var pn decimal.Decimal            //全价
	deployFee := bpInfo.DeployFee     //部署费
	coverFee := bpInfo.CoverFee       //服务费（充值手续费）
	depositFee := bpInfo.DepositFee   //押金费
	var custodySumFee decimal.Decimal //托管月费

	var moreBusinessFee decimal.Decimal //业务线优惠
	var comboFee decimal.Decimal        //套餐优惠
	var yearFee decimal.Decimal         //满年优惠
	var sumFee decimal.Decimal          //计算总优惠费
	var moreCoinFee decimal.Decimal     //计算增加主链币费
	var moreSubCoinFee decimal.Decimal  //计算增加代币费

	chainArr := strings.Split(bInfo.Coin, ",")
	if len(chainArr) > bpInfo.ChainNums {
		m := len(chainArr) - bpInfo.ChainNums
		moreCoinFee = bpInfo.ChainDiscountNums.Mul(decimal.NewFromInt(int64(m)))
	}
	subCoinArr1 := strings.Split(bInfo.SubCoin, ",")
	if len(subCoinArr1) > bpInfo.CoinNums {
		m := len(subCoinArr1) - bpInfo.CoinNums
		moreCoinFee = bpInfo.CoinDiscountNums.Mul(decimal.NewFromInt(int64(m)))
	}

	//计算托管时间：月
	custodyTime := entrustTime(req.AccountId)
	custodySumFee = bpInfo.CustodyFee.Mul(decimal.NewFromInt(int64(custodyTime)))
	//TODO:全价：指部署费+托管月费+押金费+增加主链币费+增加代币费
	isComboFee := true //是否计算套餐优惠
	isYearFee := true  //是否计算满年优惠
	var orderTypeName string
	if bpInfo.TypeName == "flow" || bpInfo.TypeName == "续费流水套餐" || bpInfo.TypeName == "流水收费套餐" { //续费流水套餐
		custodySumFee = decimal.Zero
		orderTypeName = "续费流水套餐"
	} else if bpInfo.TypeName == "address" || bpInfo.TypeName == "续费地址套餐" || bpInfo.TypeName == "地址收费套餐" { //续费地址套餐
		custodySumFee = decimal.Zero
		orderTypeName = "续费地址套餐"
	} else if bpInfo.TypeName == "month" || bpInfo.TypeName == "续费包月套餐" || bpInfo.TypeName == "包月收费套餐" { //续费包月套餐
		orderTypeName = "地址收费套餐"
	}
	log.Errorf("业务线订单orderTypeName：%v\n ", orderTypeName)

	pn = decimal.Sum(deployFee, custodySumFee, depositFee, moreCoinFee, moreSubCoinFee)
	//业务线优惠
	pInfo := _package.NewEntity()
	err = pInfo.FindPackageItemById(bpInfo.PackageId)
	isMoreBFee := isMoreBusinessFee(bpInfo.PackageId, bpInfo.AccountId, bpInfo.Id, pInfo.ServiceNums)
	if isMoreBFee {
		if pInfo.ServiceDiscountUnit == 1 { //金额
			moreBusinessFee = pInfo.ServiceDiscountNums
		} else { //折扣
			feeDiscount := decimal.NewFromInt(1).Mul(pInfo.ServiceDiscountNums)
			moreBusinessFee = pn.Mul(feeDiscount) //全价*（1-折扣）
		}
	}
	sumFee = decimal.Sum(sumFee, moreBusinessFee)
	//套餐优惠
	if isComboFee {
		if pInfo.ComboDiscountUnit == 1 {
			comboFee = pInfo.ComboDiscountNums
		} else {
			feeDiscount := decimal.NewFromInt(1).Mul(pInfo.ComboDiscountNums)
			p2 := pn.Sub(moreBusinessFee)
			comboFee = p2.Mul(feeDiscount)
		}
	}

	sumFee = decimal.Sum(sumFee, comboFee)

	//满年优惠
	if isYearFee {
		yearFee = decimal.NewFromFloat(0)
		sumFee = decimal.Sum(sumFee, yearFee)
	}

	amount := pn.Sub(sumFee)
	log.Errorf("业务线订单 详情 \n全价：%v，\n计算增加主链币费：%v，\n计算增加代币费：%v，\n业务线优惠：%v，"+
		"\n套餐优惠：%v，\n满年优惠：%v，\n计算总优惠费：%v, \n订单支付金额：%v ", pn, moreCoinFee, moreSubCoinFee,
		moreBusinessFee, comboFee, yearFee, sumFee, amount)
	log.Errorf("业务线订单 部署费 \n全价：%v，\n服务费（充值手续费）：%v，\n押金费：%v ", deployFee, coverFee, depositFee)

	//商户自己续费 审核状态自动同意
	//accountVerifyStatus := "agree"
	TotalCost = sumFee.Add(amount)
	info = domain.MchPackageInfo{
		TypeName:     TypeName,
		ModelName:    ModelName,
		BusinessName: BusinessName,
		ChainName:    ChainName,
		SubCoinName:  SubCoinName,
		Fee:          Fee,
		DeductCoin:   deductCoin,
		DiffFee:      sumFee,
		TotalCost:    TotalCost,
		EndCost:      amount,
	}

	return info, nil

}

// SearchPackages 搜索套餐列表
func SearchPackages(req *domain.PackageReqInfo) (list []domain.PackageListInfo, total int64, err error) {

	var l []_package.Entity
	pInfo := _package.NewEntity()
	l, total, err = pInfo.FindPackageListByReq(*req)
	if err != nil {
		return list, total, global.DaoError(err)
	}
	list = make([]domain.PackageListInfo, 0)
	for _, item := range l {
		b, errB := json.Marshal(item)
		if errB == nil {
			pInfo := domain.PackageListInfo{}
			errC := json.Unmarshal(b, &pInfo)
			if errC == nil {
				list = append(list, pInfo)
			}
		}
	}
	return
}

// SearchPackageScreen 搜索套餐筛选列表
func SearchPackageScreen(req *domain.PackageReqInfo) (typeList []domain.PackageScreenInfo, tradeList []domain.PackageScreenInfo, modelMap map[string][]string, err error) {

	typeList = make([]domain.PackageScreenInfo, 0)
	tradeList = make([]domain.PackageScreenInfo, 0)
	//modelList = make([]domain.PackageScreenInfo, 0)

	pInfo := _package.NewEntity()
	var typeArr []_package.PackagePay
	var tradeArr []_package.PackageTrade
	//modelMap map[string][]string

	typeArr, tradeArr, modelMap, err = pInfo.FindPackageScreen(*req)
	if err != nil {
		err = global.DaoError(err)
		return
	}
	for _, item := range typeArr {
		s := domain.PackageScreenInfo{
			Name:    item.PayName,
			PayType: item.PayType,
		}
		typeList = append(typeList, s)
	}
	for _, item := range tradeArr {
		s := domain.PackageScreenInfo{
			Name:    item.TradeName,
			PayType: item.TradeType,
		}
		tradeList = append(tradeList, s)
	}
	//for _, item := range modelArr {
	//	s := domain.PackageScreenInfo{
	//		Name: item,
	//	}
	//	modelList = append(modelList, s)
	//}
	return
}

//套餐条件判断
//判断是否满足各种限制
func determainPackageParametersIsValid(req map[string]interface{}) (bool, error) {
	return true, nil
}
