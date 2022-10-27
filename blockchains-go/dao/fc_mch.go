package dao

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/pkg/errors"
	"xorm.io/xorm"
)

type MchVerifyInfo struct {
	MchId     int64
	ApiKey    string
	ApiSecret string
}


func InsertNewMchItem(db *xorm.Session, item entity.FcMch) (err error) {
	_, err = db.Insert(item)
	return  err
}


func CustodyFcMchUpdateByApikey(item entity.FcMch)error {
	has, err := db.Conn.Where("api_key = ? and status = 2 ", item.ApiKey).Update(item)
	if err != nil {
		return err
	}
	if has<=0 {
		return  errors.New("Not Fount!")
	}
	return nil
}

func CustodyFcMchFindByApikey(db *xorm.Session, apiKey string) (*entity.FcMch, error) {
	mch := &entity.FcMch{}
	has, err := db.Where("api_key = ? and status = 2", apiKey).Get(mch)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return mch, nil
}


//获取商户计算验证的信息
func GetMchVerifyInfo(mchName string) (MchVerifyInfo, error) {
	mch := &entity.FcMch{}
	ok, err := db.Conn.Where("platform = ?", mchName).Get(mch)
	if !ok {
		return MchVerifyInfo{}, errors.New("get mch info not ok")
	}
	if err != nil {
		return MchVerifyInfo{}, err
	}
	return MchVerifyInfo{
		MchId:     int64(mch.Id),
		ApiKey:    mch.ApiKey,
		ApiSecret: mch.ApiSecret,
	}, nil
}

//func GetMchVerifyInfo(mchName string) (MchVerifyInfo, error) {
//	mch := &entity.FcMch{}
//	ok, err := db.Conn.Where("login_name = ?", mchName).Get(mch)
//	if !ok {
//		return MchVerifyInfo{}, errors.New("get mch info not ok")
//	}
//	if err != nil {
//		return MchVerifyInfo{}, err
//	}
//	return MchVerifyInfo{
//		MchId:     int64(mch.Id),
//		ApiKey:    mch.ApiKey,
//		ApiSecret: mch.ApiSecret,
//	}, nil
//}

//func GetMchInfo(mchName string) (*entity.FcMch, error) {
//	mch := &entity.FcMch{}
//	has, err := db.Conn.Where("login_name = ?", mchName).Get(mch)
//	if err != nil {
//		return nil, err
//	}
//	if !has {
//		return nil, errors.New("Not Fount!")
//	}
//	return mch, nil
//}

func FcMchFindByPlatform(platform string) (*entity.FcMch, error) {
	mch := &entity.FcMch{}
	has, err := db.Conn.Where("platform = ?", platform).Get(mch)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return mch, nil
}

func FcMchFindById(id int) (*entity.FcMch, error) {
	mch := &entity.FcMch{}
	has, err := db.Conn.Id(id).Get(mch)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return mch, nil
}

func FcMchFindByApikey(apiKey string) (*entity.FcMch, error) {
	mch := &entity.FcMch{}
	has, err := db.Conn.Where("api_key = ? and status = 2 ", apiKey).Get(mch)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return mch, nil
}

//status 有效
func FcMchFindByPlatformsAndStatus(status int, platforms []string) ([]*entity.FcMch, error) {
	results := make([]*entity.FcMch, 0)
	err := db.Conn.Where("status = ?", status).In("platform", platforms).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//查询有效的商户
func FcMchFindsVaild() ([]*entity.FcMch, error) {
	results := make([]*entity.FcMch, 0)
	err := db.Conn.Where("status = ?", 2).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
