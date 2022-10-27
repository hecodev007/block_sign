package deals

import (
	"custody-merchant-admin/internal/domain"
	modelUser "custody-merchant-admin/model/adminPermission/user"
	send "custody-merchant-admin/model/sends"
)

type PhoneDeal struct{}

func FindPhoneCode(code string) (*domain.PhoneInfo, error) {
	var pinfo = new(domain.PhoneInfo)

	dao := send.NewEntity()
	clist, err := dao.FindPhoneByCode(code)
	if err != nil {
		return nil, err
	}
	if clist != nil {
		pinfo.Tag = clist.Tag
		pinfo.CodeValue = clist.CodeValue
		pinfo.CodeName = clist.CodeName
		return pinfo, nil
	}
	return nil, nil
}

func FindPhoneCodeAll() ([]domain.PhoneInfo, error) {
	var pinfo []domain.PhoneInfo
	dao := send.NewEntity()
	clist, err := dao.FindPhoneCodeAll()
	if err != nil {
		return nil, err
	}
	for i, _ := range clist {
		pinfo = append(pinfo, domain.PhoneInfo{
			Tag:       clist[i].Tag,
			CodeName:  clist[i].CodeValue,
			CodeValue: clist[i].CodeName,
		})
	}
	return pinfo, err
}

func ValiDataEmail(email string) (*modelUser.Entity, error) {
	dao := modelUser.NewEntity()
	return dao.GetUserByEmail(email)
}

func ValiDataPhone(phone string) (*modelUser.Entity, error) {
	dao := modelUser.NewEntity()
	return dao.GetUserByPhone(phone)
}
func UpdateAdminUserByUid(uId int64, mp map[string]interface{}) (int64, error) {
	// 通过用户ID获取用户的路由权限
	dao := modelUser.NewEntity()
	return dao.UpdateUserById(uId, mp)
}
