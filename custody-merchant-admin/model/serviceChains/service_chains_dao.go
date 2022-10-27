package serviceChains

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"strings"
)

type ChainAndName struct {
	Id   int    `gorm:"column:id" json:"id,omitempty"`
	Name string `gorm:"column:name" json:"name,omitempty"`
}

type ChainAndCoin struct {
	ChainId   int    `json:"chain_id"`
	CoinId    int    `json:"coin_id"`
	ChainName string `json:"chain_name"`
	CoinName  string `json:"coin_name"`
}

type ServiceChainsInfo struct {
	Name       string `gorm:"column:name" json:"name,omitempty"`
	MerchantId int64  `json:"merchant_id" gorm:"column:merchant_id"`
	ServiceId  int    `json:"service_id" gorm:"column:service_id"`
}

type SCUInfo struct {
	Id               int64  `json:"id" gorm:"column:id"`
	MerchantId       int64  `json:"merchant_id" gorm:"column:merchant_id"`
	ServiceId        int    `json:"service_id" gorm:"column:service_id"`
	ServiceName      string `json:"service_name" gorm:"column:service_name"`
	CoinId           int    `json:"coin_id" gorm:"column:coin_id" `
	ChainAddr        string `json:"chain_addr" gorm:"column:chain_addr"`
	IsGetAddr        int    `json:"is_get_addr" gorm:"column:is_get_addr"`
	IsWithdrawal     int    `json:"is_withdrawal" gorm:"column:is_withdrawal"`
	IsIp             int    `json:"is_ip" gorm:"column:is_ip"`
	IsTest           int    `json:"is_test" gorm:"column:is_test"`
	UserName         string `gorm:"column:user_name" json:"user_name"`
	Phone            string `gorm:"column:phone" json:"phone"`
	Email            string `gorm:"column:email" json:"email"`
	PhoneCode        string `gorm:"column:phone_code" json:"phone_code"`
	UserState        int    `gorm:"column:user_state" json:"user_state"`
	CoinName         string `gorm:"column:coin_name" json:"coin_name"`
	ChainName        string `gorm:"column:chain_name" json:"chain_name"`
	IpAddr           string `gorm:"column:ip_addr" json:"ip_addr"`
	MUrl             string `gorm:"column:m_url" json:"m_url"`
	ChainState       int    `gorm:"column:chain_state" json:"chain_state"`
	ChainStateName   string `json:"chain_state_name"`
	IsIpName         string `json:"is_ip_name"`
	IsGetAddrName    string `json:"is_get_addr_name"`
	IsWithdrawalName string `json:"is_withdrawal_name"`
	UserStateName    string `json:"user_state_name"`
	IsTestName       string `json:"is_test_name"`
}

func (e *Entity) GetMerchantChainsByMidAndSid() error {
	db := e.Db.Table(e.TableName()).
		Where("merchant_id =? and service_id =? and coin_id =? and state !=2",
			e.MerchantId,
			e.ServiceId,
			e.CoinId).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetMerchantChainsByAddr(addr string) error {
	db := e.Db.Table(e.TableName()).
		Where("chain_addr =?  and state !=2", addr).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetMerchantChainsBySecureKey(secureKey string) error {
	db := e.Db.Table(e.TableName()).
		Where("secure_key =? and state !=2", secureKey).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) InsertNewItem() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

func (e *Entity) GetMerchantChainList(sc *domain.SearchChains) ([]SCUInfo, int64, error) {

	var (
		count int64
		lst   = []SCUInfo{}
	)
	db := model.DB().Table("service_chains").
		Select(
			"service_chains.id as id," +
				"service.id as service_id," +
				"service.name as service_name," +
				"service_chains.merchant_id as merchant_id," +
				"service_chains.chain_addr," +
				"service_chains.state as chain_state," +
				"service_chains.is_get_addr as is_get_addr," +
				"service_chains.is_withdrawal as is_withdrawal," +
				"user_info.name as user_name," +
				"service.phone," +
				"user_info.is_test as is_test," +
				"user_info.state as user_state," +
				"service.email," +
				"user_info.phone_code," +
				"coin_info.id as coin_id," +
				"chain_info.id as chain_id," +
				"coin_info.name as coin_name," +
				"chain_info.name as chain_name").
		Joins("left join user_info on user_info.id = service_chains.merchant_id").
		Joins("left join service on service.id = service_chains.service_id").
		Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Joins("left join chain_info on chain_info.id = coin_info.chain_id").
		Where("service_chains.state !=2")
	if sc.MerchantId != 0 {
		db.Where("user_info.id=?", sc.MerchantId)
	}
	if sc.ServiceId != 0 {
		db.Where("service.id=?", sc.ServiceId)
	}
	if sc.Account != "" {
		if strings.Contains(sc.Account, "@") {
			db.Where("user_info.email=?", sc.Account)
		} else {
			db.Where("user_info.phone=?", sc.Account)
		}
	}
	db.Order("service_chains.id desc").Offset(sc.Offset).Limit(sc.Limit).Find(&lst).Offset(-1).Limit(-1).Count(&count).Debug()
	return lst, count, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetHaveServiceChainsList(id int64) ([]ServiceChainsInfo, error) {
	var s = []ServiceChainsInfo{}
	db := model.DB().Raw("select distinct service_chains.service_id,s.name,service_chains.merchant_id "+
		"from service_chains  join service s on s.id = service_chains.service_id "+
		"where (select count(1) from service_audit_role where service_chains.service_id = service_audit_role.sid  and service_audit_role.uid = ? limit 1) > 0 group by service_chains.service_id,service_chains.merchant_id", id).
		Scan(&s)
	return s, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetNoServiceChainsList(ids []int) ([]ServiceChainsInfo, error) {
	var s = []ServiceChainsInfo{}
	db := model.DB().Raw("select service_chains.service_id,s.name,service_chains.merchant_id "+
		"from service_chains join service s on s.id = service_chains.service_id "+
		"where service_chains.service_id not in (?) group by service_chains.service_id,service_chains.merchant_id", ids).
		Scan(&s)
	return s, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetServiceChainsInfo(id int64) (*SCUInfo, error) {
	var lst = new(SCUInfo)
	db := model.DB().Table("service_chains").
		Select(
			"service_chains.id as id,"+
				"service.id as service_id,"+
				"service.name as service_name,"+
				"service_chains.merchant_id as merchant_id,"+
				"service_chains.chain_addr,"+
				"service_chains.state as chain_state,"+
				"service_chains.is_get_addr as is_get_addr,"+
				"service_chains.is_withdrawal as is_withdrawal,"+
				"user_info.name as user_name,"+
				"user_info.phone,"+
				"user_info.state as user_state,"+
				"user_info.email,"+
				"user_info.phone_code,"+
				"coin_info.id as coin_id,"+
				"chain_info.id as chain_id,"+
				"coin_info.name as coin_name,"+
				"chain_info.name as chain_name").
		Joins("left join user_info on user_info.id = service_chains.merchant_id").
		Joins("left join service on service.id = service_chains.service_id").
		Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Joins("left join chain_info on chain_info.id = coin_info.chain_id").
		Where("service_chains.id=? and service_chains.state !=2", id).
		First(lst)
	return lst, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetServiceChainslist(id int) ([]SCUInfo, error) {
	list := []SCUInfo{}
	db := model.DB().Table("service_chains").
		Select("service_chains.*,service.name as service_name,chain_info.name as chain_name,coin_info.name as coin_name").
		Joins("left join service on service.id = service_chains.service_id").
		Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Joins("left join chain_info on chain_info.id = coin_info.chain_id").
		Where("service_chains.service_id = ? and service_chains.state !=2", id).Find(&list)
	return list, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindServiceChainsByMainlist(sel *domain.SelectUserInfo) ([]SCUInfo, int64, error) {
	list := []SCUInfo{}
	var count int64
	db := model.DB().Table("service_chains").
		Select("service_chains.service_id as service_id,service.name as service_name, group_concat(distinct coin_info.name) as coin_name,group_concat(distinct chain_info.name) as chain_name").
		Joins("left join service on service.id = service_chains.service_id").
		Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Joins("left join chain_info on chain_info.id = coin_info.chain_id").
		Where("service_chains.merchant_id = ? and service_chains.state !=2", sel.MerchantId).
		Group("service_chains.service_id").
		Offset(sel.Offset).Limit(sel.Limit).Find(&list).Offset(-1).Limit(-1).Count(&count)

	return list, count, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) UpdateServiceChainsInfo(id int64, mp map[string]interface{}) error {
	db := model.DB().Table("service_chains").Where("id = ?", id).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func (e *Entity) FindServiceChainsInfo(sid int, coin string) error {
	db := e.Db.Table("service_chains").Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Where("service_chains.service_id=? and coin_info.name = ? and service_chains.state !=2", sid, coin).First(e)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func (e *Entity) FindServiceChainsInfoByAddr(addr string, coin string) error {
	db := e.Db.Table("service_chains").Joins("left join coin_info on coin_info.id = service_chains.coin_id").
		Where("service_chains.chain_addr=? and coin_info.name = ? and service_chains.state !=2", addr, coin).First(e)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}
