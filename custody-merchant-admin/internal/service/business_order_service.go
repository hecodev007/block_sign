package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	modelUser "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/model/apply"
	"custody-merchant-admin/model/assets"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/business"
	"custody-merchant-admin/model/businessOrder"
	"custody-merchant-admin/model/businessPackage"
	"custody-merchant-admin/model/fullYear"
	"custody-merchant-admin/model/orm"
	_package "custody-merchant-admin/model/package"
	"fmt"

	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
	"time"
)

//NewBusinessCreateNewOrder 新业务线创建业务线订单
func NewBusinessCreateNewOrder(user *domain.JwtCustomClaims, bInfo *business.Entity, bpInfo *businessPackage.Entity) (err error) {
	orderId := xkutils.NewUUId("YWXDD")
	var pn decimal.Decimal //全价

	var custodySumFee decimal.Decimal   //托管月费
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
	subCoinArr := strings.Split(bInfo.SubCoin, ",")
	if len(subCoinArr) > bpInfo.CoinNums {
		m := len(subCoinArr) - bpInfo.CoinNums
		moreCoinFee = bpInfo.CoinDiscountNums.Mul(decimal.NewFromInt(int64(m)))
	}

	//计算托管时间：月
	if bpInfo.TypeName == "month" { //只有包月套餐收包月费
		custodyTime := entrustTime(user.Id)
		custodySumFee = bpInfo.CustodyFee.Mul(decimal.NewFromInt(int64(custodyTime)))
	}
	//全价：指部署费+托管月费+押金费+增加主链币费+增加代币费
	pn = decimal.Sum(bpInfo.DeployFee, custodySumFee, bpInfo.DepositFee, moreCoinFee, moreSubCoinFee)
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
	if pInfo.ComboDiscountUnit == 1 {
		comboFee = pInfo.ComboDiscountNums
	} else {
		feeDiscount := decimal.NewFromInt(1).Mul(pInfo.ComboDiscountNums)
		p2 := pn.Sub(moreBusinessFee)
		comboFee = p2.Mul(feeDiscount)
	}
	sumFee = decimal.Sum(sumFee, comboFee)

	//满年优惠
	if isFullYear(bpInfo) {
		if pInfo.YearDiscountUnit == 1 { //金额
			yearFee = pInfo.YearDiscountNums
		} else { //折扣
			feeDiscount := decimal.NewFromInt(1).Mul(pInfo.YearDiscountNums)
			yearFee = pn.Mul(feeDiscount) //全价*（1-折扣）
		}
	}
	sumFee = decimal.Sum(sumFee, yearFee)
	amount := pn.Sub(sumFee)
	log.Infof("业务线订单 %v 详情 \n全价：%v，\n计算增加主链币费：%v，\n计算增加代币费：%v，\n业务线优惠：%v，"+
		"\n套餐优惠：%v，\n满年优惠：%v，\n计算总优惠费：%v, \n订单支付金额：%v ", orderId, pn, moreCoinFee, moreSubCoinFee,
		moreBusinessFee, comboFee, yearFee, sumFee, amount)
	//暂定只根据部署费收费
	amount = bpInfo.DeployFee
	oInfo := businessOrder.Entity{
		Db:                bInfo.Db,
		AccountId:         bInfo.AccountId,
		OrderType:         "首次开通",
		OrderId:           orderId,
		BusinessId:        bInfo.Id,
		PackageId:         bpInfo.PackageId,
		BusinessPackageId: int64(bpInfo.Id),
		//AddBusinessFee:    moreBusinessFee,
		//AddChainFee:       moreCoinFee,
		//AddSubChainFee:    moreSubCoinFee,
		//FullYearFee:       yearFee,
		//DiscountFee:       sumFee,
		OrderCoinName: bpInfo.DeductCoin,
		ProfitNumber:  amount,
	}
	err = oInfo.InsertNewItem()
	log.Errorf("InsertNew order err = %v", err)
	return
}

//UpdateBusinessCreateNewOrder 更新业务线创建业务线订单
func UpdateBusinessCreateNewOrder(user *domain.JwtCustomClaims, bInfo *business.Entity, bpInfo *businessPackage.Entity, orderType string) (err error) {
	orderId := xkutils.NewUUId("YWXDD")
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
	subCoinArr := strings.Split(bInfo.SubCoin, ",")
	if len(subCoinArr) > bpInfo.CoinNums {
		m := len(subCoinArr) - bpInfo.CoinNums
		moreCoinFee = bpInfo.CoinDiscountNums.Mul(decimal.NewFromInt(int64(m)))
	}

	//计算托管时间：月
	custodyTime := entrustTime(user.Id)
	custodySumFee = bpInfo.CustodyFee.Mul(decimal.NewFromInt(int64(custodyTime)))
	//TODO:全价：指部署费+托管月费+押金费+增加主链币费+增加代币费
	isComboFee := true //是否计算套餐优惠
	isYearFee := true  //是否计算满年优惠

	if orderType == "change_type" || orderType == "变更套餐类型" { //变更套餐类型
		if !strings.Contains(bpInfo.ModelName, "包月") && !strings.Contains(bpInfo.ModelName, "month") {
			custodySumFee = decimal.Zero
		}
	} else if orderType == "change_model" || orderType == "变更收费模式" { //变更收费模式
		deployFee = decimal.Zero
		custodySumFee = decimal.Zero
		depositFee = decimal.Zero
		isComboFee = false
	} else if orderType == "add_chain" || orderType == "增加主链币" { //增加主链币
		deployFee = decimal.Zero
		custodySumFee = decimal.Zero
		depositFee = decimal.Zero
		isComboFee = false
		isYearFee = false
	} else if orderType == "add_subcoin" || orderType == "增加代币" { //增加代币
		deployFee = decimal.Zero
		custodySumFee = decimal.Zero
		depositFee = decimal.Zero
		moreCoinFee = decimal.Zero
		isComboFee = false
		isYearFee = false
	} else if orderType == "renew_flow" || orderType == "续费流水套餐" { //续费流水套餐
		custodySumFee = decimal.Zero
	} else if orderType == "renew_address" || orderType == "续费地址套餐" { //续费地址套餐
		custodySumFee = decimal.Zero
	} else if orderType == "renew_month" || orderType == "续费包月套餐" { //续费包月套餐
	}
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
	log.Errorf("业务线订单 %v详情 \n全价：%v，\n计算增加主链币费：%v，\n计算增加代币费：%v，\n业务线优惠：%v，"+
		"\n套餐优惠：%v，\n满年优惠：%v，\n计算总优惠费：%v, \n订单支付金额：%v ", orderId, pn, moreCoinFee, moreSubCoinFee,
		moreBusinessFee, comboFee, yearFee, sumFee, amount)
	log.Errorf("业务线订单 部署费 \n全价：%v，\n服务费（充值手续费）：%v，\n押金费：%v ", deployFee, coverFee, depositFee)
	//暂定只根据部署费收费
	amount = bpInfo.DeployFee
	oInfo := businessOrder.Entity{
		Db:                bInfo.Db,
		AccountId:         bInfo.AccountId,
		OrderType:         orderType,
		OrderId:           orderId,
		BusinessId:        bInfo.Id,
		PackageId:         bpInfo.PackageId,
		BusinessPackageId: int64(bpInfo.Id),
		//AddBusinessFee:    moreBusinessFee,
		//AddChainFee:       moreCoinFee,
		//AddSubChainFee:    moreSubCoinFee,
		//DiscountFee:       sumFee,
		OrderCoinName: bpInfo.DeductCoin,
		ProfitNumber:  amount,
	}
	err = oInfo.InsertNewItem()
	return
}

//RenewBusinessOrder //商家 续费业务线订单
func RenewBusinessOrder(accountId int64) (err error) {
	//获取当前业务线-套餐
	bpInfo := businessPackage.NewEntity()
	err = bpInfo.FindBPItemByAccountId(accountId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if accountId != bpInfo.AccountId {
		return global.WarnMsgError(global.DataWarnNoBelongErr)
	}

	bInfo := business.NewEntity()
	err = bInfo.FindBusinessItemById(bpInfo.BusinessId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	orderId := xkutils.NewUUId("YWXDD")
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
	subCoinArr := strings.Split(bInfo.SubCoin, ",")
	if len(subCoinArr) > bpInfo.CoinNums {
		m := len(subCoinArr) - bpInfo.CoinNums
		moreCoinFee = bpInfo.CoinDiscountNums.Mul(decimal.NewFromInt(int64(m)))
	}

	//计算托管时间：月
	custodyTime := entrustTime(accountId)
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

	//暂定只根据部署费收费
	amount = bpInfo.DeployFee
	oInfo := businessOrder.Entity{
		Db:                bInfo.Db,
		AccountId:         bInfo.AccountId,
		OrderType:         orderTypeName,
		OrderId:           orderId,
		BusinessId:        bInfo.Id,
		PackageId:         bpInfo.PackageId,
		BusinessPackageId: int64(bpInfo.Id),
		//AddBusinessFee:    moreBusinessFee,
		//AddChainFee:       moreCoinFee,
		//AddSubChainFee:    moreSubCoinFee,
		//DiscountFee:       sumFee,
		OrderCoinName: bpInfo.DeductCoin,
		ProfitNumber:  amount,
		//AccountVerifyState: accountVerifyStatus,
	}
	err = oInfo.InsertNewItem()
	return
}

//isMoreBusinessFee 查询当前业务线的套餐 属于第几个，是否是超出业务线
func isMoreBusinessFee(pId, userId int64, bpId, num int) bool {
	bpInfo := businessPackage.NewEntity()
	list, _ := bpInfo.FindBPItemByPIdAndUserId(pId, userId)
	for i, item := range list {
		if item.Id == bpId {
			n := i + 1
			if n > num {
				return true
			} else {
				return false
			}
		}
	}
	return false

}

//entrustTime 用户托管/合同时间
func entrustTime(userId int64) (month int) {
	uInfo := apply.NewEntity()
	uInfo.FindApplyItemByAccountId(userId)
	if uInfo.ContractEndAt == nil || uInfo.ContractStartAt == nil {
		return 0
	}
	if TimeIsNull(*uInfo.ContractEndAt) || TimeIsNull(*uInfo.ContractStartAt) {
		return 0
	}
	month = SubMonth(*uInfo.ContractEndAt, *uInfo.ContractStartAt)
	return
}

//isFullYear 计算是否满足满年优惠
func isFullYear(bpInfo *businessPackage.Entity) bool {
	fInfo := fullYear.NewEntity()
	fInfo.FindItemById(bpInfo.AccountId, bpInfo.PackageId)
	if fInfo.Id == 0 {
		log.Errorf("没有满年优惠数据 bpInfo.AccountId= %v, bpInfo.PackageId= %v", bpInfo.AccountId, bpInfo.PackageId)
		return false
	}
	t := time.Now().Local()
	allH := int(t.Sub(fInfo.LatestTime).Hours())
	allD := allH / 24
	if allD < 360 {
		return false
	}
	return true
}

//AccountVerifyBusinessOrder 商家审核业务线订单,同意/拒绝（商户操作）
func AccountVerifyBusinessOrder(req *domain.AccountOperateInfo) (err error) {
	//获取当前业务线-套餐
	bInfo := businessOrder.NewEntity()
	err = bInfo.FindBusinessOrderItemByOrderId(req.OrderId)
	if err == gorm.ErrRecordNotFound {
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if req.AccountId != bInfo.AccountId {
		return global.WarnMsgError(global.DataWarnNoBelongErr)
	}
	if bInfo.AccountVerifyState != "" && bInfo.AccountVerifyState != "wait" {
		return global.WarnMsgError(global.DataWarnHadVerifyErr)
	}
	if req.Operate == "agree" || req.Operate == "refuse" {
		bInfo.AccountVerifyState = req.Operate
		t := time.Now().Local()
		bInfo.AccountVerifyTime = &t
		bInfo.AccountRemark = req.Remark
		err = bInfo.UpdateBusinessOrderItem()
	} else {
		return global.WarnMsgError(global.DataWarnNoOperateErr)
	}

	return
}

//AdminVerifyBusinessOrder 管理后台审核业务线订单,执行扣款/拒绝（管理员操作）
/*
扣款 只在托管后台数据扣除，钱包地址不进行扣款
在托管后台记录钱包地址这笔钱属于财务
*/
func AdminVerifyBusinessOrder(user *domain.JwtCustomClaims, req *domain.AccountOperateInfo) (err error) {
	//获取当前业务线-套餐
	if req.OrderId == "" {
		req.OrderId = req.Id
	}
	db := orm.Cache(model.DB().Begin())
	boInfo := businessOrder.NewEntity()
	boInfo.Db = db
	err = boInfo.FindBusinessOrderItemByOrderId(req.OrderId)
	log.Infof("业务线订单 1 req:%+v\n", req)

	if err == gorm.ErrRecordNotFound {
		log.Errorf("业务线订单 FindBusinessOrderItemByOrderId err:%v", err)
		return global.WarnMsgError(global.DataWarnNoDataErr)
	}
	if boInfo.AccountVerifyState == "" {
		return global.WarnMsgError(global.DataWarnNoAccountVerifyErr)
	}
	if boInfo.AccountVerifyState == "refuse" {
		return global.WarnMsgError(global.DataWarnAccountRefuseErr)
	}
	if boInfo.AccountVerifyState != "agree" {
		return global.WarnMsgError(global.DataWarnNoAccountAgreeErr)
	}
	if req.Operate == "agree" {
		if boInfo.FullYearFee.Cmp(decimal.NewFromInt(0)) == 1 {
			//判断是否满足满年
			bpInfo := businessPackage.NewEntity()
			bpInfo, err = bpInfo.FindBPItemById(boInfo.BusinessPackageId)
			if bpInfo.Id == 0 {
				log.Errorf("业务线订单err:%v", err)
				return global.WarnMsgError(global.DataWarnNoDataErr)
			}
			if !isFullYear(bpInfo) {
				return global.WarnMsgError(global.DataWarnNoFeeErr)
			}
		}
		//换算成扣币币种
		orderCoinInfo, _ := GetCoinInfo(boInfo.OrderCoinName)
		if orderCoinInfo.PriceUsd.Cmp(decimal.Zero) != 1 {
			err = fmt.Errorf("服务器币种金额错误%v:%v\n", boInfo.OrderCoinName, orderCoinInfo.PriceUsd)
			return
		}
		profitNumber := boInfo.ProfitNumber.Div(orderCoinInfo.PriceUsd)

		//冻结业务线金额
		err = LockAsset(db, boInfo.BusinessId, profitNumber, boInfo.OrderCoinName, "order_lock", "业务线订单冻结")
		if err != nil {
			db.Rollback()
			log.Errorf("执行扣款err :%v", err)
			return global.WarnMsgError(global.DataWarnOrderDeductErr)
		}

		//托管后台assets扣款
		aInfo := assets.NewEntity()
		aInfo.Db = db
		var assetItem *assets.Assets
		var orderCoinId int64
		if boInfo.OrderCoinId == 0 {
			cInfo, _ := base.FindCoinsByName(boInfo.OrderCoinName)
			orderCoinId = cInfo.Id
		} else {
			orderCoinId = boInfo.OrderCoinId
		}
		boInfo.OrderCoinId = orderCoinId
		assetItem, err = GetDateAssetsBySIdAndCId(int(boInfo.BusinessId), int(orderCoinId))
		log.Errorf("账户 余额:%v，订单金额:%v(USDT),订单金额:%v(%v)", assetItem.Nums, boInfo.ProfitNumber, profitNumber, boInfo.OrderCoinName)

		if assetItem.Nums.Cmp(profitNumber) == -1 {
			db.Rollback()
			return global.WarnMsgError(global.DataWarnBalanceErr)
		}
		err = aInfo.UpDateAssetsSubNumsBySCId(int(boInfo.BusinessId), int(orderCoinId), profitNumber)
		if err != nil {
			db.Rollback()
			log.Errorf("执行扣款err :%v", err)
			return global.WarnMsgError(global.DataWarnOrderDeductErr)
		}
		boInfo.DeductState = "success"
		//记录至财务账号
		err = SyncToFinanceAssetsByBusinessOrder(boInfo, aInfo)
		if err != nil {
			db.Rollback()
			log.Errorf("记录流水err :%v", err)
			return global.WarnMsgError(global.DataWarnOrderDeductErr)
		}

		//更新满年优惠时间
		fyInfo := fullYear.NewEntity()
		fyInfo.Db = db
		fyInfo.AccountId = boInfo.AccountId
		fyInfo.PackageId = boInfo.PackageId
		fyInfo.BusinessId = boInfo.BusinessId
		fyInfo.LatestTime = time.Now().Local()
		err = fyInfo.InsertNewItem()
		if err != nil {
			db.Rollback()
			log.Errorf("err:%v", err)
			return global.WarnMsgError(global.DataWarnNoDataErr)
		}
		//记录商户业务线
		err = DealOrderAgreeBusinessChange(db, boInfo)
		if err != nil {
			db.Rollback()
			log.Errorf("DealOrderAgreeBusinessChange err:%v", err)
			return global.WarnMsgError(global.DataWarnUpdateDataErr)
		}

		// 续费、开通的时候添加套餐收益户
		err = SaveComboIncome(profitNumber, boInfo.AccountId, boInfo.BusinessId, boInfo.OrderCoinName)
		if err != nil {
			db.Rollback()
			log.Errorf("续费、开通的时候添加套餐收益户err:%v", err)
			return global.WarnMsgError(global.DataWarnNoDataErr)
		}
	}

	log.Errorf("业务线订单  Operate:%+v\n", req.Operate)
	if req.Operate == "agree" || req.Operate == "refuse" {
		boInfo.AdminVerifyState = req.Operate
		t := time.Now().Local()
		boInfo.AdminVerifyTime = &t
		boInfo.AdminVerifyId = user.Id
		boInfo.AdminRemark = req.Remark
		err = boInfo.UpdateBusinessOrderItem()
		if err != nil {
			db.Rollback()
			log.Errorf("执行扣款更新 err :%v", err)
			return global.WarnMsgError(global.DataWarnOrderDeductErr)
		}
	} else {
		db.Rollback()
		err = global.WarnMsgError(global.DataWarnNoOperateErr)
	}
	db.Commit()
	return
}

//SearchBusinessOrderList 业务线订单列表
func SearchBusinessOrderList(req *domain.OrderReqInfo) (list []domain.BusinessOrderInfo, total int64, err error) {
	//获取当前业务线-套餐
	bpInfo := businessOrder.NewEntity()
	var arr []businessOrder.Item
	arr, total, err = bpInfo.FindBusinessOrderInfoListByReq(req)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = global.OperationErrorText(err.Error())
			return
		} else {
			log.Error(err.Error())
			err = nil
			return nil, 0, nil
		}
	}
	list = make([]domain.BusinessOrderInfo, 0)
	for _, item := range arr {
		at := GetTimeString(item.AdminVerifyTime)
		avt := GetTimeString(item.AccountVerifyTime)
		ct := GetTimeString(item.CreatedAt)
		totalFee := item.DeployFee.
			Add(item.CustodyFee).
			Add(item.DepositFee).
			Add(item.AddBusinessFee).
			Add(item.AddChainFee).
			Add(item.AddSubChainFee).
			Add(item.DiscountFee)
		adminVerifyState := item.AdminVerifyState
		accountVerifyState := item.AccountVerifyState
		if adminVerifyState == "" {
			if accountVerifyState == "refuse" {
				adminVerifyState = "admin-refuse"
			} else {
				adminVerifyState = "wait"
			}
		}
		if accountVerifyState == "" {
			if adminVerifyState == "refuse" {
				accountVerifyState = "account-refuse"
			} else {
				accountVerifyState = "wait"
			}
		}
		dao := modelUser.NewEntity()
		info, _ := dao.GetUserById(item.AdminVerifyId)
		var adminName string
		if info != nil {
			adminName = info.Name
		}

		i := domain.BusinessOrderInfo{
			Name:               item.Name,
			AccountStatus:      item.AccountStatus,
			AccountId:          item.AccountId,
			Email:              item.Email,
			Phone:              item.Phone,
			OrderType:          item.OrderType,
			OrderId:            item.OrderId,
			TypeName:           item.TypeName,
			ModelName:          item.ModelName,
			BusinessId:         item.BusinessId,
			BusinessName:       item.BusinessName,
			Coin:               item.Coin,
			SubCoin:            item.SubCoin,
			DeployFee:          item.DeployFee,
			CustodyFee:         item.CustodyFee,
			DepositFee:         item.DepositFee,
			CoverFee:           item.CoverFee,
			AddBusinessFee:     item.AddBusinessFee,
			AddChainFee:        item.AddChainFee,
			AddSubChainFee:     item.AddSubChainFee,
			DiscountFee:        item.DiscountFee,
			TotalFee:           totalFee,
			ProfitNumber:       item.ProfitNumber,
			DeductCoin:         item.OrderCoinName, //扣费币种
			DeductCoinName:     item.OrderCoinName, //扣费币种
			AdminVerifyId:      item.AdminVerifyId,
			AdminVerifyTime:    at,
			AdminVerifyState:   adminVerifyState,
			AdminVerifyName:    adminName,
			AccountVerifyTime:  avt,
			AccountVerifyState: accountVerifyState,
			Remark:             item.Remark,
			CreateTime:         ct,
		}
		list = append(list, i)
	}
	return
}

//DealOrderAgreeBusinessChange 订单同意是 业务线变化
func DealOrderAgreeBusinessChange(db *orm.CacheDB, boInfo *businessOrder.Entity) (err error) {
	if boInfo.OrderType == "open" || boInfo.OrderType == "首次开通" {
		bInfo := business.NewEntity()
		bInfo.Id = boInfo.BusinessId
		bInfo.Db = db
		err = bInfo.UpdateBusinessItemByMap(map[string]interface{}{"state": 0})
	} else if boInfo.OrderType == "renew_flow" || boInfo.OrderType == "续费流水套餐" || boInfo.OrderType == "renew_address" || boInfo.OrderType == "续费地址套餐" { //续费流水套餐
		pInfo := _package.NewEntity()
		err = pInfo.FindPackageItemById(boInfo.PackageId)
		if err != nil || pInfo.Id == 0 {
			err = fmt.Errorf("订单无套餐")
			log.Errorf("FindPackageItemById pid = %v", boInfo.PackageId)
			log.Errorf("FindPackageItemById err = %v", err)
			return
		}
		bpInfo := businessPackage.NewEntity()
		bpInfo.Db = db
		err = bpInfo.AddBPItemTypeNumsBySIdByMap(boInfo.BusinessPackageId, pInfo.TypeNums)
	} else if boInfo.OrderType == "renew_month" || boInfo.OrderType == "续费包月套餐" { //续费包月套餐
	}
	return
}
