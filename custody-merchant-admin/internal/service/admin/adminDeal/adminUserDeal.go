package adminDeal

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/adminPermission/permission"
	modelUser "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/util/library"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"time"
)

func GetAdminInfoByUserId(id int64) (*domain.UserPersonal, error) {
	dao := modelUser.NewEntity()
	personal, err := dao.GetAdminPersonal(id)
	if err != nil {
		return nil, err
	}
	if personal != nil {
		dp := &domain.UserPersonal{
			Id:        personal.Id,
			Sex:       personal.Sex,
			Pid:       personal.Pid,
			Name:      personal.Name,
			Email:     personal.Email,
			Phone:     personal.Phone,
			PhoneCode: personal.PhoneCode,
			Role:      personal.Role,
			RoleName:  dict.SysRoleNameList[personal.Role-1],
			Passport:  personal.Passport,
			Identity:  personal.Identity,
			LoginTime: personal.LoginTime.Format(global.YyyyMmDdHhMmSs),
		}
		return dp, err
	} else {
		return nil, nil
	}
}

func SaveAdminUser(duser *domain.SaveUserInfo) (int64, error) {

	dao := modelUser.NewEntity()
	err := HaveUserPhoneEmail(duser.Phone, duser.Email, &modelUser.Entity{Email: "", Phone: ""})
	if err != nil {
		return 0, err
	}
	pInfo, err := dao.GetAdminPersonal(duser.Pid)
	if err != nil {
		return 0, err
	}
	if pInfo.Id == 0 {
		return 0, errors.New(global.MsgWarnPidIsNil)
	}
	u := &modelUser.Entity{
		Pid:       duser.Pid,
		Name:      duser.Name,
		Sex:       duser.Sex,
		Email:     duser.Email,
		Phone:     duser.Phone,
		PhoneCode: duser.PhoneCode,
		Password:  library.EncryptSha256Password(duser.Password, "noSalt"),
		Salt:      "noSalt",
		Passport:  duser.Passport,
		Identity:  duser.Identity,
		Role:      duser.Role,
		Remark:    duser.Remark,
		State:     0,
		CreatedAt: time.Now().Local(),
	}
	uId, err := dao.SaveUser(u)
	if err != nil {
		return uId, err
	}
	// 新增人员
	if uId > 0 {
		return uId, nil
	} else {
		return uId, errors.New(global.MsgWarnModelAdd)
	}
}

func UpdateUser(user *domain.SaveUserInfo) error {
	dao := modelUser.NewEntity()
	idu, err := dao.HaveUserId(user.Id)
	if err != nil {
		return err
	}
	if idu == nil {
		return errors.New(global.MsgWarnNotUser)
	}
	err = HaveUserPhoneEmail(user.Phone, user.Email, idu)
	if err != nil {
		return err
	}
	u := map[string]interface{}{
		"pid":        user.Pid,
		"name":       user.Name,
		"sex":        user.Sex,
		"email":      user.Email,
		"phone":      user.Phone,
		"phone_code": user.PhoneCode,
		"passport":   user.Passport,
		"identity":   user.Identity,
		"roles":      user.Role,
		"remark":     user.Remark,
		"state":      0,
		"updated_at": time.Now().Local(),
	}
	// 先更新用户信息
	err = dao.UpdatePersonalUser(user.Id, u)
	if err != nil {
		return err
	}
	return nil
}

func LoginPhone(phone, password string) (*modelUser.Entity, error) {
	dao := modelUser.NewEntity()
	return dao.GetUserByPhoneAndPwd(phone, password)
}

func LoginEmail(email, password string) (*modelUser.Entity, error) {
	dao := modelUser.NewEntity()
	return dao.GetUserByEmailAndPwd(email, password)
}

func UpdatePwdById(id int64, pwd, salt string) error {
	dao := modelUser.NewEntity()
	return dao.UpdatePwdById(id, pwd, salt)
}

func UpdatePwdByPhone(phone string, mp map[string]interface{}) error {
	dao := modelUser.NewEntity()
	return dao.UpdatePwdByPhone(phone, mp)
}

func UpdatePwdByEmail(email string, mp map[string]interface{}) error {
	dao := modelUser.NewEntity()
	return dao.UpdatePwdByEmail(email, mp)
}

func GetSysRoleIsSuperAdmin(ids int, tag string) (bool, error) {
	if ids <= 0 && ids > len(dict.SysRoleTagList) {
		return false, global.WarnMsgError(global.MsgWarnModelNil)
	}
	return xkutils.ThreeDo(dict.SysRoleTagList[ids-1] == tag, true, false).(bool), nil
}

func GetSaltByPhoneAndEmail(phone, email string) (*modelUser.Entity, error) {
	dao := modelUser.NewEntity()
	return dao.GetSaltByPhoneAndEmail(phone, email)
}

func GetAdminUserInfoList(userSelect *domain.SelectUserInfo, user *domain.JwtCustomClaims) ([]domain.SelectAdminUserList, int64, error) {

	var (
		dao        = modelUser.NewEntity()
		selectList = []domain.SelectAdminUserList{}
		uid        int64
	)

	// 不是超级管理员
	if !user.Admin {
		uid = user.Id
	}
	list, count, err := dao.FindAdminUserInfoList(uid, userSelect)
	if err != nil {
		return selectList, count, err
	}
	// 遍历用户信息
	for i, _ := range list {
		selectList = append(selectList, domain.SelectAdminUserList{
			Serial:       i + 1,
			Id:           list[i].Id,
			Name:         list[i].Name,
			Sex:          list[i].Sex,
			Email:        list[i].Email,
			Phone:        list[i].Phone,
			PhoneCode:    list[i].PhoneCode,
			Role:         list[i].Role,
			PwdErr:       list[i].PwdErr,
			PhoneCodeErr: list[i].PhoneCodeErr,
			EmailCodeErr: list[i].EmailCodeErr,
			RoleName:     dict.SysAdminRoleNameList[list[i].Role-1],
			Remark:       list[i].Remark,
			Reason:       list[i].Reason,
			State:        list[i].State,
			Show:         xkutils.ThreeDo(list[i].Id == user.Id, 0, 1).(int),
			IsMerchant:   xkutils.ThreeDo(list[i].Uid == 0, 0, 1).(int),
			Passport:     list[i].Passport,
			Identity:     list[i].Identity,
			SexName:      dict.SexText[list[i].Sex],
			StateName:    dict.StateText[list[i].State],
			LoginTime:    list[i].LoginTime.Local().Format(global.YyyyMmDdHhMmSs),
			CreateTime:   list[i].CreatedAt.Local().Format(global.YyyyMmDdHhMmSs),
		})
	}
	return selectList, count, nil
}

func GetAdminUserInfoById(uid int64) (domain.SelectAdminUserList, error) {
	var (
		dao        = modelUser.NewEntity()
		userPerm   = permission.NewEntity()
		selectList domain.SelectAdminUserList
		menus      []int
	)
	info, err := dao.GetAdminPersonal(uid)
	if err != nil {
		return selectList, err
	}

	uMenu, err := userPerm.GetUserPermission(uid)
	if err != nil {
		return domain.SelectAdminUserList{}, err
	}
	if uMenu != nil {
		menus, err = xkutils.IntSplitByString(uMenu.Mid, ",")
		if err != nil {
			return domain.SelectAdminUserList{}, err
		}
	}

	selectList = domain.SelectAdminUserList{
		Id:         info.Id,
		Name:       info.Name,
		Sex:        info.Sex,
		Email:      info.Email,
		Phone:      info.Phone,
		PhoneCode:  info.PhoneCode,
		Role:       info.Role,
		RoleName:   dict.SysRoleNameList[info.Role-1],
		Remark:     info.Remark,
		Reason:     info.Reason,
		State:      info.State,
		SexName:    dict.SexText[info.Sex],
		StateName:  dict.StateText[info.State],
		Passport:   info.Passport,
		Identity:   info.Identity,
		LoginTime:  info.LoginTime.Local().Format(global.YyyyMmDdHhMmSs),
		CreateTime: info.CreatedAt.Local().Format(global.YyyyMmDdHhMmSs),
		Menus:      menus,
	}
	return selectList, nil
}

func UpdateAdminUserInfoById(id int64, uIfon *domain.SelectUserList) (int64, error) {
	var (
		upMap map[string]interface{}
		dao   = modelUser.NewEntity()
	)
	if uIfon.State == 0 {
		upMap = map[string]interface{}{
			"state":  uIfon.State,
			"reason": uIfon.Reason,
		}
	}
	if uIfon.State == 1 || uIfon.State == 2 {
		upMap = map[string]interface{}{
			"state":  uIfon.State,
			"reason": uIfon.Reason,
		}
	}
	// OperationDelUserHaveServiceErr
	info, err := dao.GetUserById(id)
	if err != nil {
		return 0, err
	}
	if info == nil {
		return 0, global.NewError(global.MsgWarnAccountErr)
	}
	del, err := dao.UpdateUserById(id, upMap)
	if err != nil {
		return 0, err
	}
	return del, nil
}

func UpdateUserByUid(uId int64, mp map[string]interface{}) (int64, error) {
	dao := modelUser.NewEntity()
	// 通过用户ID获取用户的路由权限
	return dao.UpdateUserById(uId, mp)
}

func HaveUserPhoneEmail(phone, email string, user *modelUser.Entity) error {
	dao := modelUser.NewEntity()
	if phone == "" && email == "" {
		return global.WarnMsgError(global.MsgWarnNoEmailPhone)
	} else {
		if email != "" && user.Email != email {
			userEmail, err := dao.GetUserByEmail(email)
			if err != nil {
				return err
			}
			if userEmail != nil && userEmail.Id > 0 {
				return global.WarnMsgError(global.MsgWarnHaveEmail)
			}
		}
		if phone != "" && user.Phone != phone {
			userPhone, err := dao.GetUserByPhone(phone)
			if err != nil {
				return err
			}
			if userPhone != nil && userPhone.Id > 0 {
				return global.WarnMsgError(global.MsgWarnHavePhone)
			}
		}
	}
	return nil
}

func HaveUserId(id int64) error {
	dao := modelUser.NewEntity()
	user, err := dao.HaveUserId(id)
	if err != nil {
		return err
	}
	if user != nil {
		return errors.New("这个Id已经被占用")
	}
	return nil
}

func GetUserInfoByUserId(id int64) (*domain.UserPersonal, error) {
	dao := modelUser.NewEntity()
	personal, err := dao.GetAdminPersonal(id)
	if err != nil {
		return nil, err
	}
	if personal != nil {
		dp := &domain.UserPersonal{
			Sex:       personal.Sex,
			Pid:       personal.Pid,
			Name:      personal.Name,
			Email:     personal.Email,
			Phone:     personal.Phone,
			PhoneCode: personal.PhoneCode,
			Role:      personal.Role,
			RoleName:  dict.SysRoleNameList[personal.Role-1],
			Passport:  personal.Passport,
			Identity:  personal.Identity,
			LoginTime: personal.LoginTime.Format(global.YyyyMmDdHhMmSs),
		}
		return dp, err
	} else {
		return nil, nil
	}
}

func CheckPwdErr(id int64) (bool, error) {
	dao := modelUser.NewEntity()
	byId, err := dao.GetUserById(id)
	if err != nil {
		return false, err
	}
	if byId.PwdErr >= 5 {
		return false, nil
	}
	return true, nil
}

func CheckPhoneCodeErr(id int64) (bool, error) {

	dao := modelUser.NewEntity()
	byId, err := dao.GetUserById(id)
	if err != nil {
		return false, err
	}
	if byId.PhoneCodeErr >= 5 {
		return false, nil
	}
	return true, nil
}

func CheckEmailCodeErr(id int64) (bool, error) {

	dao := modelUser.NewEntity()
	byId, err := dao.GetUserById(id)
	if err != nil {
		return false, err
	}
	if byId.EmailCodeErr >= 5 {
		return false, nil
	}
	return true, nil
}

func HaveAdminUserByPIdAndUId(id, pid int64) error {
	dao := modelUser.NewEntity()
	user, err := dao.HaveAdminUserByPIdAndUId(id, pid)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(global.MsgWarnAccountIsSubErr)
	}
	return nil
}
