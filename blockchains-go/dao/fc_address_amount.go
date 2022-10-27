package dao

import (
	"errors"
	"fmt"

	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

type FcExcludeFreeze struct {
	Address      string `json:"address" xorm:"not null comment('地址') VARCHAR(128)"`
	Amount       string `json:"amount" xorm:"not null comment('可用金额') VARCHAR(128)"`
	FreezeAmount string `json:"forzen_amount" xorm:"not null comment('冻结金额') VARCHAR(128)"`
	Type         int    `json:"type" xorm:"not null default 0 comment('地址类型 1 冷地址 2 用户地址  3 手续费地址') TINYINT(3)"`
}

// FcAddressAmountExcludeFreeze 获取地址的可用金额（排除冻结的金额）数据
// mchId 商户id
// limitAmount 限制的最低金额，排除低于此值的数据
// coin 币种
// limit 获取条数
func FcAddressAmountExcludeFreeze(mchId int, limitAmount string, coin string, limit int) ([]FcExcludeFreeze, error) {
	results := make([]FcExcludeFreeze, 0)
	err := db.Conn.SQL("SELECT * FROM  (select (amount-forzen_amount) AS amount,address,forzen_amount,type from fc_address_amount where coin_type = ? and app_id = ? and type in (1,2)) AS t where t.amount > ? order by amount desc limit ?", coin, mchId, limitAmount, limit).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

type FcBalanceSta struct {
	Coin   string `json:"coin" xorm:"not null comment('币种名称') VARCHAR(48)"`
	Amount string `json:"amount" xorm:"not null comment('总金额') VARCHAR(128)"`
}

// 获取总余额 = 出账地址 + 用户地址
func FcAddressAmountBalanceTotal(mchId int) ([]*FcBalanceSta, error) {
	results := make([]*FcBalanceSta, 0)

	err := db.Conn.SQL("select coin_type AS coin, sum(amount) AS amount from fc_address_amount where type IN(1,2) AND app_id=? AND amount > 0 group by coin_type", mchId).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

// 获取垫资地址 = 手续费地址
func FcAddressAmountBalanceFee(mchId int) ([]*FcBalanceSta, error) {
	results := make([]*FcBalanceSta, 0)

	err := db.Conn.SQL("select coin_type AS coin, sum(amount) AS amount from fc_address_amount where type IN(3) AND app_id=? AND amount > 0 group by coin_type", mchId).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

// 获取出账地址余额 = 出账地址
func FcAddressAmountBalanceOut(mchId int) ([]*FcBalanceSta, error) {
	results := make([]*FcBalanceSta, 0)

	err := db.Conn.SQL("select coin_type AS coin, sum(amount) AS amount from fc_address_amount where type=1 and app_id=? AND amount > 0 group by coin_type", mchId).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

// 获取可用余额 = 出账地址 + 可归集用户地址
func FcAddressAmountBalanceLiquid(mchId int) ([]*FcBalanceSta, error) {
	results := make([]*FcBalanceSta, 0)

	err := db.Conn.SQL("select fa.coin_type AS coin, sum(fa.amount) AS amount from fc_coin_set fc join fc_address_amount fa on fa.coin_type = fc.name  where fa.amount > fc.sta_threshold and fa.type in (1,2) and fa.app_id=? group by fa.coin_type", mchId).Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

// 获取可用余额 = 出账地址 + 可归集用户地址
func FcAddressAmountBalanceLiquidByCoin(mchId int64, coin string) (*FcBalanceSta, error) {
	result := new(FcBalanceSta)
	get, err := db.Conn.SQL("select fa.coin_type AS coin, sum(fa.amount) AS amount from fc_coin_set fc join fc_address_amount fa on fa.coin_type = fc.name where fa.amount > fc.sta_threshold and fa.type in (1, 2) and fa.app_id = ? and fc.name = ?", mchId, coin).Get(result)
	if !get {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

//根据币种获取大于出账金额的出账冷地址
func FcAddressAmountGetCloudAddress(typed, status, mchId int, coinName string, amount string) (*entity.FcAddressAmount, error) {
	//In("column", builder.Select("column").From("table2").Where(builder.Eq{"a":1})).Find()
	result := new(entity.FcAddressAmount)
	if has, err := db.Conn.In("address", builder.Select("address").From("fc_generate_address_list").
		Where(builder.Eq{
			"type":        typed,
			"status":      status,
			"platform_id": mchId,
			"coin_name":   coinName,
		})).Where("amount >= ?", amount).Get(&result); err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("don't find for any %T", result)
	}
	return result, nil
}

func FcAddressAmountGetTotalAmount(mchId int, coinName string) (decimal.Decimal, error) {
	result := new(entity.FcAddressAmount)
	total, err := db.Conn.Where("app_id = ? and coin_type = ?", mchId, coinName).Sum(result, "amount")
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(total), nil
}
func FcAddressAmountGetTotalAmountWithType(mchId int, coinName string, coinType int) (decimal.Decimal, error) {
	result := new(entity.FcAddressAmount)
	total, err := db.Conn.Where("app_id = ? and coin_type = ? and type = ?", mchId, coinName, coinType).Sum(result, "amount")
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(total), nil
}

func FcAddressAmountGetByCoinAndAddrs(coinName string, addrs []string) (*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Where("coin_type = ? and forzen_amount = 0", coinName).
		And(builder.In("address", addrs)).
		Desc("amount").
		Limit(1).
		Find(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any data")
	}
	return results[0], nil
}

func FcFindAddressAmountUserAddrList(mchId int64, coinName string, addrs []string) ([]entity.FcAddressAmount, error) {
	results := make([]entity.FcAddressAmount, 0)
	err := db.Conn.Where("coin_type = ? AND app_id = ?", coinName, mchId).
		And(builder.In("address", addrs)).
		And(builder.In("type", []address.AddressType{address.AddressTypeUser, address.AddressTypeCold})).
		Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountGetTotalAmountWithLimit(mchId int, coinName string, limit int) (decimal.Decimal, error) {
	//负载较高，暂时隐藏

	//result := new(entity.FcAddressAmount)
	//total, err := db.Conn.Where("app_id = ? and coin_type = ? ", mchId, coinName).
	//	In("type", []int{1, 2}).
	//	Desc("amount").
	//	Limit(limit).
	//	Sum(result, "amount")
	//if err != nil {
	//	return decimal.Zero, err
	//}
	total := 0.0
	return decimal.NewFromFloat(total), nil
}

func FcAddressAmountFindAddresses(typed, status, mchId int, coinName string, amount string) ([]string, error) {
	results := make([]string, 0)
	err := db.Conn.Table("fc_address_amount").Cols("address").
		Where("type = ? and status = ? and platform_id = ? and coin_type = ? and amount >= ?", typed, status, mchId, coinName, amount).
		Find(&results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("don't find any address")
	}
	return results, nil
}

//typed  地址类型 1 冷地址 2 用户地址  3 手续费地址
//coinName  币种
//amount  限制金额查询
//limit  需要的数量
//查询无冻结的归集地址
func FcAddressAmountFindCollectAddr(typed int, coinName string, amount string, limit int) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Where("type = ? and coin_type = ? and amount >= ? and forzen_amount = 0", typed, coinName, amount).
		Desc("amount").
		Limit(limit).
		Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//更新冻结金额
func FcAddressAmountUpdateAddForzenAmount(address, coinType string, appId int64, addAmount decimal.Decimal) error {
	fa := new(entity.FcAddressAmount)
	has, err := db.Conn.Where("address = ? and coin_type = ? and app_id = ?", address, coinType, appId).Get(fa)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("Not Fount!")
	}
	forzenAmount, _ := decimal.NewFromString(fa.ForzenAmount)
	forzenAmount = forzenAmount.Add(addAmount)
	faNew := &entity.FcAddressAmount{
		ForzenAmount: forzenAmount.String(),
	}
	_, err = db.Conn.Id(fa.Id).Cols("forzen_amount").Update(faNew)
	return err
}

func FcAddressAmountUpdateSubForzenAmount(address, coinType string, appId int64, subAmount decimal.Decimal) error {
	fa := new(entity.FcAddressAmount)
	has, err := db.Conn.Where("address = ? and coin_type = ? and app_id = ?", address, coinType, appId).Get(fa)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("Not Fount!")
	}
	forzenAmount, _ := decimal.NewFromString(fa.ForzenAmount)
	forzenAmount = forzenAmount.Sub(subAmount)
	faNew := &entity.FcAddressAmount{
		ForzenAmount: forzenAmount.String(),
	}
	_, err = db.Conn.Id(fa.Id).Cols("forzen_amount").Update(faNew)
	return err
}

func FindLessThanOutAmountList(mchId int, coinType, amount, collectThreshold string, limit int) ([]entity.FcAddressAmount, error) {
	lockIds, err := FcCollectLockIds()
	if err != nil {
		return nil, err
	}
	log.Infof("获取到归集锁定Ids %v", lockIds)
	results := make([]entity.FcAddressAmount, 0)
	err = db.Conn.Table("fc_address_amount").
		Where("coin_type = ? AND amount >= ? AND amount < ? AND forzen_amount = 0 AND app_id = ? AND type = ?", coinType, collectThreshold, amount, mchId, address.AddressTypeUser).
		And(builder.NotIn("id", lockIds)).
		Desc("amount").
		Limit(limit).
		Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

func FindUserAddrAmountList(mchId int, coinCode string, count int) ([]entity.FcAddressAmount, error) {
	results := make([]entity.FcAddressAmount, 0)
	err := db.Conn.Table("fc_address_amount").
		Where("coin_type = ? AND forzen_amount = 0 AND app_id = ? AND type = ?", coinCode, mchId, address.AddressTypeUser).
		Desc("amount").
		Limit(count).
		Find(&results)

	if err != nil {
		return nil, err
	}
	return results, nil
}

func GetNearOutAmount(mchId int, coinType, amount string, collectThreshold string) (*entity.FcAddressAmount, error) {
	lockIds, err := FcCollectLockIds()
	if err != nil {
		return nil, err
	}
	log.Infof("[GetNearOutAmount]获取到归集锁定Ids %v", lockIds)
	// 出账金额和归集阈值，两者取大
	amt, _ := decimal.NewFromString(amount)
	ct, _ := decimal.NewFromString(collectThreshold)
	if ct.Cmp(amt) == 1 {
		amt = ct
	}
	result := new(entity.FcAddressAmount)
	exist, err := db.Conn.Table("fc_address_amount").
		Where("coin_type = ? AND amount >= ? AND forzen_amount = 0 AND app_id = ? AND type = ?", coinType, amt.String(), mchId, address.AddressTypeUser).
		And(builder.NotIn("id", lockIds)).
		Desc("amount").
		Get(result)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return result, nil
}

//params appId: 	商户ID
//params coinName: 	币种名
//params limit: 	需要的记录数
//params ascOrDescByAmount: 升序或者降序查询，升序从小到大查找出账地址，降序 从大到小查询地址
func FcAddressAmountFindTransfer(appId int64, coinName string, limit int, ascOrDescByAmount string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	var err error
	if ascOrDescByAmount == "asc" {
		err = db.Conn.Table("fc_address_amount").
			Where("type in (1,2) and coin_type = ? and amount > 0 and app_id = ?", coinName, appId).
			Asc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "desc" {
		err = db.Conn.Table("fc_address_amount").
			Where("type in (1,2) and coin_type = ? and amount > 0 and app_id = ?", coinName, appId).
			Desc("amount").
			Limit(limit).
			Find(&results)
	} else {
		err = errors.New("error ascOrDesc type")
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountFindTransferToBtcForCollect(appId int64, limit int, start int) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Table("fc_address_amount").
		Where("address like '3%' and type=2 and coin_type = 'btc' and amount>0.000001 and app_id = ?", appId).
		Desc("amount").
		Limit(limit, start).
		Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountFindTransferToBtcForMerge(appId int64, limit int, start int) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Table("fc_address_amount").
		Where("address like '3%' and type=1 and coin_type = 'btc' and amount>0.000001 and app_id = ?", appId).
		Desc("amount").
		Limit(limit, start).
		Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//btc特定查询使用，查询末尾金额。过滤指定金额，因为现有逻辑btc和usdt地址混在一起了，唯一区分只能是3开头地址
//params appId: 	商户ID
//params limit: 	需要的记录数
//params ascOrDescByAmount: 升序或者降序查询，升序从小到大查找出账地址，降序 从大到小查询地址
func FcAddressAmountFindTransferToBtc(appId int64, limit int, ascOrDescByAmount string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	var err error
	if ascOrDescByAmount == "asc" {
		err = db.Conn.Table("fc_address_amount").
			Where("address like '3%' and type in (1,2) and coin_type = 'btc' and amount > 0 and app_id = ?", appId).
			Asc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "desc" {
		err = db.Conn.Table("fc_address_amount").
			Where("address like '3%' and type in (1,2) and coin_type = 'btc' and amount > 0 and app_id = ?", appId).
			Desc("amount").
			Limit(limit).
			Find(&results)
	} else {
		err = errors.New("error ascOrDesc type")
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountFindTransferToUca(appId int64, limit int, ascOrDescByAmount string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	var err error
	if ascOrDescByAmount == "asc" {
		err = db.Conn.Table("fc_address_amount").
			Where("type in (1,2) and coin_type = 'uca' and amount > 0 and app_id = ?", appId).
			Asc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "desc" {
		err = db.Conn.Table("fc_address_amount").
			Where("type in (1,2) and coin_type = 'uca' and amount > 0 and app_id = ?", appId).
			Desc("amount").
			Limit(limit).
			Find(&results)
	} else {
		err = errors.New("error ascOrDesc type")
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

//usdt特定查询使用，查询末尾金额。过滤指定金额，因为现有逻辑btc和usdt地址混在一起了，唯一区分只能是1开头地址
//params appId: 	商户ID
//params limit: 	需要的记录数
//params ascOrDescByAmount: 升序或者降序查询，升序从小到大查找出账地址，降序 从大到小查询地址
func FcAddressAmountFindTransferToUsdt(appId int64, limit int, ascOrDescByAmount string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	var err error
	if ascOrDescByAmount == "asc" {
		err = db.Conn.Table("fc_address_amount").
			Where("address like '1%' and type = 2 and coin_type = 'usdt' and amount > 0 and app_id = ?", appId).
			Asc("amount").
			Limit(limit).
			Find(&results)
	} else if ascOrDescByAmount == "desc" {
		err = db.Conn.Table("fc_address_amount").
			Where("address like '1%' and type = 2 and coin_type = 'usdt' and amount > 0 and app_id = ?", appId).
			Desc("amount").
			Limit(limit).
			Find(&results)
	} else {
		err = errors.New("error ascOrDesc type")
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

//有些特殊的币种账户模型只有单个出账地址
func FcAddressAmountFindTransferToAccount(coinName string, mchId int64) (*entity.FcAddressAmount, error) {
	result := new(entity.FcAddressAmount)
	has, err := db.Conn.Where("coin_type = ? and app_id = ? and type = 1 ", coinName, mchId).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcAddressAmountInternal(mchId int, addressList []string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Where(builder.In("type", []address.AddressType{address.AddressTypeCold, address.AddressTypeFee}),
		builder.Eq{"app_id": mchId},
		builder.In("address", addressList)).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountFindAddress(typed, mchId int, address string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Where("type = ? and app_id = ?  and address = ?", typed, mchId, address).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountByAddrs(address []string, mchId int, coin string) ([]*entity.FcAddressAmount, error) {
	results := make([]*entity.FcAddressAmount, 0)
	err := db.Conn.Where("type = ? and app_id = ? and coin_type = ?", 1, mchId, coin).And(builder.In("address", address)).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcAddressAmountUpdatePendingAmount(coinName, address string, amount decimal.Decimal) error {
	result := new(entity.FcAddressAmount)
	has, err := db.Conn.Where("coin_type = ? and address = ?  ", coinName, address).Get(result)
	if err != nil {
		return err
	}
	if !has {
		//未找到
		return errors.New("not found")
	}
	pendingAmount, _ := decimal.NewFromString(result.PendingAmount)
	log.Infof("冻结金额 %s ，解冻金额 %s ", pendingAmount.String(), amount.String())
	if amount.GreaterThan(pendingAmount) {
		return fmt.Errorf("冻结金额 %s ，解冻金额 %s ", pendingAmount.String(), amount.String())
	}
	pendingAmount = pendingAmount.Sub(amount)
	//更新
	_, err = db.Conn.Exec("update fc_address_amount set pending_amount = ? where coin_type = ? and address = ? ",
		pendingAmount.String(), coinName, address)
	return err
}

func FcAddressAmountFindByCoinNameAndAddress(coinName, address string) (*entity.FcAddressAmount, error) {
	result := new(entity.FcAddressAmount)
	has, err := db.Conn.Where("coin_type = ? and address = ?", coinName, address).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcAddressAmountUpdateAmount(coinName, address string, amount decimal.Decimal) error {
	//更新
	_, err := db.Conn.Exec("update fc_address_amount set amount = ? where coin_type = ? and address = ? ",
		amount.String(), coinName, address)
	return err
}

func FcAddressAmountUpdateCoinId(id int64, coinId int) error {
	//更新
	_, err := db.Conn.Exec("update fc_address_amount set coin_id = ? where id = ? and coin_id=0", coinId, id)
	return err
}

func FcFindAddressAmountFilter(mchId int64, coinType string, addrType address.AddressType, limit int) ([]entity.FcAddressAmount, error) {
	results := make([]entity.FcAddressAmount, 0)
	err := db.Conn.Where("coin_type = ? AND app_id = ? AND type = ? AND amount > 0", coinType, mchId, addrType).
		Desc("amount").
		Limit(limit).
		Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

type FcUpdateFreeze struct {
	Address      string
	FreezeAmount string // 本次需要冻结的金额
}

func FcAddressAmountUpdateFreeze(coinType string, models []FcUpdateFreeze) error {
	session := db.Conn.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	for _, m := range models {
		session.Exec("update fc_address_amount set forzen_amount = forzen_amount + ? where coin_type = ? and address = ? limit 1",
			m.FreezeAmount, coinType, m.Address)
	}

	return session.Commit()
}

func FcAddressAmountUpdateUnFreeze(coinType string, models []FcUpdateFreeze) error {
	session := db.Conn.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}

	for _, m := range models {
		session.Exec("update fc_address_amount set forzen_amount = forzen_amount - ? where coin_type = ? and address = ? limit 1",
			m.FreezeAmount, coinType, m.Address)
	}

	return session.Commit()
}
