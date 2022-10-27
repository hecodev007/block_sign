package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin/adminDeal"
	"custody-merchant-admin/model/adminPermission/operate"
	"custody-merchant-admin/model/merchant"
	"errors"
	"time"
)

func UpdateSubUserInfo(u *domain.SaveUserInfo) error {
	err := adminDeal.UpdateMerchantSubUser(u)
	if err != nil {
		return err
	}
	return nil
}

func GetUserInfoById(uid int64) (domain.SelectUserList, error) {
	userList, err := adminDeal.GetUserInfoById(uid)
	if err != nil {
		return userList, err
	}
	return userList, err
}

func GetDaoUserInfoById(uid int64) (*merchant.Entity, error) {
	mDao := merchant.NewEntity()
	userList, err := mDao.GetUserMerchantPersonal(uid)
	if err != nil {
		return userList, err
	}
	return userList, err
}

func UpdateSubUserById(uInfo *domain.SelectUserList) error {
	err := adminDeal.UpdateSubUserInfoById(uInfo)
	if err != nil {
		return err
	}
	return nil
}

func UpdateClearUserByPId(pid int64, mp map[string]interface{}) error {
	err := adminDeal.ClearSubInfoByPId(pid, mp)
	if err != nil {
		return err
	}
	return nil
}

func UpdateClearUserById(id int64, mp map[string]interface{}) error {
	err := adminDeal.ClearSubInfoById(id, mp)
	if err != nil {
		return err
	}
	return nil
}

func GetUserServiceAudit(id int64) (*domain.UserHaveServiceAuditLevel, error) {
	return adminDeal.GetUserServiceAudit(id)
}
func GetAllMerchantService(id int64) ([]domain.UService, error) {
	return adminDeal.GetAllMerchantService(id)
}

func AddUserOperateUId(uid int64, userby, content string) error {
	dao := operate.NewEntity()
	mDao := merchant.NewEntity()
	personal, err := mDao.GetUserMerchantPersonal(uid)
	if err != nil {
		return err
	}
	if personal == nil {
		return errors.New(global.MsgWarnAccountErr)
	}
	dao.UserId = uid
	dao.UserName = personal.Name
	dao.CreatedBy = userby
	dao.CreatedAt = time.Now().Local()
	dao.Content = content
	dao.Platform = "管理后台"
	return dao.NewOperate()
}

func FindUserOperateUId(uid int64) ([]domain.OperateList, error) {
	dao := operate.NewEntity()
	operate := []domain.OperateList{}
	lst, err := dao.FindOperateByUserId(uid)
	if err != nil {
		return nil, err
	}
	for i, _ := range lst {
		operate = append(operate, domain.OperateList{
			Id:        lst[i].Id,
			UserId:    lst[i].UserId,
			UserName:  lst[i].UserName,
			CreatedBy: lst[i].CreatedBy,
			Content:   lst[i].Content,
			Platform:  lst[i].Platform,
			CreatedAt: lst[i].CreatedAt.Format(global.YyyyMmDdHhMmSs),
		})
	}
	return operate, nil
}

func HaveUserByPIdAndUId(id, pid int64) error {
	return adminDeal.HaveUserByPIdAndUId(id, pid)
}
