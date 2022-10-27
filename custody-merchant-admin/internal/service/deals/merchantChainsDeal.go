package deals

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/base"
	"custody-merchant-admin/model/merchant"
	"custody-merchant-admin/model/serviceChains"
	"custody-merchant-admin/model/serviceSecurity"
	"custody-merchant-admin/model/services"
	"custody-merchant-admin/module/dict"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

func SaveMerchantChains(chains *domain.UpdateChains) error {
	dao := serviceChains.NewEntity()
	coins, err := base.FindCoinsById(chains.CoinId)
	if err != nil {
		return err
	}
	dao.ServiceId = chains.ServiceId
	dao.State = 0
	dao.CoinId = chains.CoinId
	dao.ChainAddr = chains.ChainAddr
	dao.MerchantId = chains.MerchantId
	dao.IsGetAddr = chains.IsGetAddr
	dao.IsWithdrawal = chains.IsWithdrawal
	dao.Account = chains.Account
	dao.CoinName = coins.Name
	dao.CreatedAt = time.Now().Local()
	err = dao.InsertNewItem()
	if err != nil {
		return err
	}
	//新增链路 不需要在创建assets了

	//assetsDao := assets.NewEntity()
	//_, err = assetsDao.CreateAssets(assets.Assets{
	//	CoinId:    chains.CoinId,
	//	ServiceId: chains.ServiceId,
	//	CoinName:  coins.Name,
	//	Nums:      decimal.Zero,
	//	Freeze:    decimal.Zero,
	//})
	//if err != nil {
	//	return err
	//}
	return nil
}

func GetMerchantChainsByAddr(chains *domain.UpdateChains) error {
	dao := serviceChains.NewEntity()
	dao.ServiceId = chains.ServiceId
	dao.CoinId = chains.CoinId
	dao.MerchantId = chains.MerchantId
	dao.State = 0
	err := dao.GetMerchantChainsByMidAndSid()
	if err != nil {
		return err
	}
	if dao.Id > 0 {
		return errors.New("该业务线下的币种已经存在链路")
	}
	return nil
}

func GetMerchantChainList(userSelect *domain.SearchChains) ([]serviceChains.SCUInfo, int64, error) {
	dao := serviceChains.NewEntity()
	list, total, err := dao.GetMerchantChainList(userSelect)
	if err != nil {
		return nil, 0, err
	}
	for i, _ := range list {
		list[i].IsGetAddrName = dict.BaseText[list[i].IsGetAddr]
		list[i].IsWithdrawalName = dict.BaseText[list[i].IsWithdrawal]
		list[i].UserStateName = dict.StateText[list[i].UserState]
		list[i].ChainStateName = dict.StateText[list[i].ChainState]
		list[i].IsTestName = dict.IsTestText[list[i].IsTest]
		security := serviceSecurity.NewEntity()
		err = security.FindItemByBusinessId(int64(list[i].ServiceId))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, err
		}
		if security.Id <= 0 {
			continue
		}
		list[i].IpAddr = security.IpAddr
		list[i].IsIp = security.IsIp
		list[i].MUrl = security.CallbackUrl
		list[i].IsIpName = dict.BaseText[security.IsIp]
	}
	return list, total, nil
}

func GetServiceChainsInfo(id int64) (*serviceChains.SCUInfo, error) {
	dao := serviceChains.NewEntity()
	us, err := dao.GetServiceChainsInfo(id)
	if err != nil {
		return us, err
	}
	if us.MerchantId != 0 {
		us.IsGetAddrName = dict.BaseText[us.IsGetAddr]
		us.IsWithdrawalName = dict.BaseText[us.IsWithdrawal]
		us.UserStateName = dict.StateText[us.UserState]
		us.ChainStateName = dict.StateText[us.ChainState]
		security := serviceSecurity.NewEntity()
		err = security.FindItemByBusinessId(int64(us.ServiceId))
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			us.IpAddr = security.IpAddr
			us.IsIp = security.IsIp
			us.MUrl = security.CallbackUrl
			us.IpAddr = dict.StateText[us.UserState]
			us.IsIpName = dict.BaseText[security.IsIp]
		}
	}
	return us, nil
}

func UpdateMerchantChainsInfo(id int64, mp map[string]interface{}) error {
	dao := serviceChains.NewEntity()
	mp["updated_at"] = time.Now().Local()
	//info, err := dao.GetServiceChainsInfo(id)
	//if err != nil {
	//	return err
	//}
	err := dao.UpdateServiceChainsInfo(id, mp)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMerchantChainsInfo(id int64) error {
	dao := serviceChains.NewEntity()
	err := dao.UpdateServiceChainsInfo(id, map[string]interface{}{
		"state":      2,
		"deleted_at": time.Now().Local(),
	})
	if err != nil {
		return err
	}
	return nil
}

func FindServiceChainsByMainlist(sel *domain.SelectUserInfo) ([]domain.MerchantServiceChains, int64, error) {
	dao := serviceChains.NewEntity()
	audit := merchant.NewEntity()
	res := []domain.MerchantServiceChains{}
	chainslist, count, err := dao.FindServiceChainsByMainlist(sel)
	if err != nil {
		return res, count, err
	}
	i := 1
	for _, info := range chainslist {

		var (
			adminNums   int64
			auditNums   int64
			financeNums int64
			visitorNums int64
		)

		levels, err := audit.CountLevelBySid(info.ServiceId)
		if err != nil {
			return nil, 0, err
		}

		if levels != nil {

			for _, level := range levels {
				if level.Roles == 2 {
					// 管理员
					adminNums = level.Nums
				}
				if level.Roles == 3 {
					// 审核员
					auditNums = level.Nums
				}
				if level.Roles == 4 {
					// 财务
					financeNums = level.Nums
				}
				if level.Roles == 5 {
					// 财务
					visitorNums = level.Nums
				}
			}
		}
		res = append(res, domain.MerchantServiceChains{
			Serial:      i,
			ServiceId:   info.ServiceId,
			ServiceName: info.ServiceName,
			ChainName:   info.ChainName,
			CoinName:    info.CoinName,
			AdminNums:   adminNums,
			AuditNums:   auditNums,
			FinanceNums: financeNums,
			VisitorNums: visitorNums,
		})
	}
	return res, count, nil
}

func GetServiceChains(id int) (domain.ServiceAndCoin, error) {
	dao := serviceChains.NewEntity()
	res := domain.ServiceAndCoin{}

	chainslist, err := dao.GetServiceChainslist(id)
	if err != nil {
		return res, err
	}

	if len(chainslist) != 0 {
		res.ServiceId = chainslist[0].ServiceId
		res.ServiceName = chainslist[0].ServiceName
	}
	for _, info := range chainslist {
		res.ChainsCoinList = append(res.ChainsCoinList, domain.ChainsCoins{
			ChainName: info.ChainName,
			CoinName:  info.CoinName,
		})
	}
	return res, nil
}

func GetServiceChainsRolesInfo(id int) (domain.ServiceRolesInfo, error) {
	dao := merchant.NewEntity()
	s := services.ServiceEntity{}
	res := domain.ServiceRolesInfo{}
	sInfo, err := s.GetServiceById(id)
	if err != nil {
		return domain.ServiceRolesInfo{}, err
	}
	res.ServiceId = id
	res.ServiceName = sInfo.Name

	adminInfo, err := dao.FindUserInfosBySid(id, 2)

	if err != nil {
		return res, err
	}
	auditInfo, err := dao.FindUserInfosBySid(id, 3)
	if err != nil {
		return res, err
	}
	financeInfo, err := dao.FindUserInfosBySid(id, 4)
	if err != nil {
		return res, err
	}
	visitorInfo, err := dao.FindUserInfosBySid(id, 5)
	if err != nil {
		return res, err
	}
	admins := domain.ServiceRoles{}
	admins.Nums = len(adminInfo)
	admins.Name = "管理员"
	for _, entity := range adminInfo {
		admins.UserAndId = append(admins.UserAndId, fmt.Sprintf("%s-%d", entity.Name, entity.Id))
	}
	audits := domain.ServiceRoles{}
	audits.Nums = len(auditInfo)
	audits.Name = "审核员"
	for _, entity := range adminInfo {
		audits.UserAndId = append(audits.UserAndId, fmt.Sprintf("%s-%d", entity.Name, entity.Id))
	}
	finances := domain.ServiceRoles{}
	finances.Nums = len(financeInfo)
	finances.Name = "财务"
	for _, entity := range financeInfo {
		finances.UserAndId = append(finances.UserAndId, fmt.Sprintf("%s-%d", entity.Name, entity.Id))
	}
	visitors := domain.ServiceRoles{}
	visitors.Nums = len(visitorInfo)
	visitors.Name = "游客"
	for _, entity := range visitorInfo {
		visitors.UserAndId = append(visitors.UserAndId, fmt.Sprintf("%s-%d", entity.Name, entity.Id))
	}
	res.AdminInfo = admins
	res.AuditInfo = audits
	res.FinanceInfo = finances
	res.VisitorInfo = visitors
	return res, nil
}
