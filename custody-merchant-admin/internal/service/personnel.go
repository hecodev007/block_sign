package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin/adminDeal"
)

func AddAdminUserInfo(u *domain.SaveUserInfo) error {
	uId, err := adminDeal.SaveAdminUser(u)
	if err != nil {
		return err
	}
	u.Id = uId
	err = adminDeal.SaveAdminUserMenu(u.Id, u.Menus)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAdminUserInfo(u *domain.SaveUserInfo) error {

	err := adminDeal.UpdateUser(u)
	if err != nil {
		return err
	}
	err = adminDeal.SaveAdminUserMenu(u.Id, u.Menus)
	if err != nil {
		return err
	}

	return nil
}

func GetUserInfoListService(userSelect *domain.SelectUserInfo) ([]domain.SelectUserList, int64, error) {
	userList, total, err := adminDeal.GetUserInfoList(userSelect)
	if err != nil {
		return userList, 0, err
	}
	return userList, total, err
}

func GetMainUserInfoService(sel *domain.SelectUserInfo) (domain.SelectUserList, error) {
	userInfo, err := adminDeal.GetMainUserInfo(sel)
	if err != nil {
		return userInfo, err
	}
	return userInfo, err
}

func GetAdminUserInfoListService(userSelect *domain.SelectUserInfo, user *domain.JwtCustomClaims) ([]domain.SelectAdminUserList, int64, error) {
	userList, total, err := adminDeal.GetAdminUserInfoList(userSelect, user)
	if err != nil {
		return userList, 0, err
	}
	return userList, total, err
}

func GetAdminUserInfoById(uid int64) (domain.SelectAdminUserList, error) {
	userList, err := adminDeal.GetAdminUserInfoById(uid)
	if err != nil {
		return userList, err
	}
	return userList, err
}

func DelAdminUserInfo(u *domain.SelectUserList) (int64, error) {
	// 删除全部用户有关的权限
	return adminDeal.UpdateAdminUserInfoById(u.Id, u)
}

func UpdateUserState(uid int64, uIfon *domain.SelectUserList) (int64, error) {
	return adminDeal.UpdateAdminUserInfoById(uid, uIfon)
}

func HaveUserId(id int64) error {
	return adminDeal.HaveUserId(id)
}

func HaveAdminUserByPIdAndUId(id, pid int64) error {
	return adminDeal.HaveAdminUserByPIdAndUId(id, pid)
}

func SaveSuperAudit(s *domain.MerchantService) error {
	return adminDeal.SaveSuperAudit(s)
}

func FindSuperAudit(s *domain.MerchantService) (domain.MerchantServiceList, error) {
	return adminDeal.FindServiceAudit(s.Id)
}

func GetMainServiceList(merchantId int64) (domain.SelectUserList, error) {
	// 获取商户有的业务线

	userInfo, err := adminDeal.GetMainUserInfo(&domain.SelectUserInfo{MerchantId: merchantId})
	if err != nil {
		return userInfo, err
	}
	return userInfo, err
}
