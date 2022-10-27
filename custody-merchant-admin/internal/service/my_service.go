package service

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin/adminDeal"
)

func UpdateOurInfo(uInfo *domain.SaveUserInfo) error {
	_, err := adminDeal.UpdateUserByUid(uInfo.Id, map[string]interface{}{
		"name":       uInfo.Name,
		"phone":      uInfo.Phone,
		"phone_code": uInfo.PhoneCode,
		"email":      uInfo.Email,
		"sex":        uInfo.Sex,
		"identity":   uInfo.Identity,
		"passport":   uInfo.Passport,
	})
	return err
}
