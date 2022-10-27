package dao

import (
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
)

//FcCoinList 获取所有币
func FcCoinList(limit int, offset int) ([]*entity.FcCoinSet, error) {
	results := make([]*entity.FcCoinSet, 0)
	err := db.Conn.Select("fc_coin_set.id,fc_coin_set.pid,fc_coin_set.price,fc_coin_set.name,fc_coin_set.status,"+
		"fc_coin_set.token,fc_coin_set.decimal,fc_coin_set.confirm,fc_coin_set.num,fc_coin_set.huge_num").Limit(limit, offset).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
	//
}

//status 有效
func FcCoinSetFindByStatus(status int) ([]*entity.FcCoinSet, error) {
	results := make([]*entity.FcCoinSet, 0)
	err := db.Conn.Where("status = ?", status).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcCoinSetGetByStatus(id int, status int) (*entity.FcCoinSet, error) {
	result := new(entity.FcCoinSet)
	has, err := db.Conn.Where("id = ? and status = ?", id, status).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcCoinSetFindByPidStatus(pid int, status int) ([]*entity.FcCoinSet, error) {
	results := make([]*entity.FcCoinSet, 0)
	err := db.Conn.Where("pid = ? and status = ?", pid, status).Find(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func FcCoinSetGetByName(coinName string, status int) (*entity.FcCoinSet, error) {
	result := new(entity.FcCoinSet)
	has, err := db.Conn.Where("name = ? and status = ?", coinName, status).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func GetFcCoinSetByName(coinName string) (*entity.FcCoinSet, error) {
	result := new(entity.FcCoinSet)
	has, err := db.Conn.Where("name = ?", coinName).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("coinName %s not found in coinSet", coinName)
	}
	return result, nil
}

func FcCoinSetGetByToken(token string, pid int) (*entity.FcCoinSet, error) {
	result := new(entity.FcCoinSet)
	has, err := db.Conn.Where("pid = ? and token = ?", pid, token).Get(result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return result, nil
}

func FcCoinSetGetCoinId(name, contractAddress string) (*entity.FcCoinSet, error) {
	result := entity.FcCoinSet{}
	var (
		has bool
		err error
	)
	if contractAddress == "" {
		has, err = db.Conn.Where("name = ?", name).Get(&result)
	} else {
		has, err = db.Conn.Where("name = ? and token = ? and pid > 0", name, contractAddress).Get(&result)
	}
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return &result, nil
}

func FcCoinSetGetCoinInfo(id int) (*entity.FcCoinSet, error) {
	result := entity.FcCoinSet{}
	var (
		has bool
		err error
	)
	has, err = db.Conn.Id(id).Get(&result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return &result, nil
}

func FcCoinSetGetCoinByContract(contractAddress string) (*entity.FcCoinSet, error) {
	result := entity.FcCoinSet{}
	var (
		has bool
		err error
	)
	has, err = db.Conn.Where("token = ? and pid > 0", contractAddress).Get(&result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return &result, nil
}
func FcCoinSetGetCoinByContractAndPid(contractAddress string, pid int) (*entity.FcCoinSet, error) {
	result := entity.FcCoinSet{}
	var (
		has bool
		err error
	)
	has, err = db.Conn.Where("token = ? and pid = ?", contractAddress, pid).Get(&result)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Not Fount!")
	}
	return &result, nil
}

func FcCoinSetCloseAllCollect(pid int) error {
	fc := entity.FcCoinSet{
		IsCollect: 0,
	}
	_, err := db.Conn.Cols("is_collect").Where("pid = ?", pid).Update(&fc)
	if err != nil {
		return err
	}
	return nil
}

func FcCoinSetCloseCollect(name string) error {
	fc := entity.FcCoinSet{
		IsCollect: 0,
	}
	_, err := db.Conn.Cols("is_collect").Where("name = ?", name).Update(&fc)
	if err != nil {
		return err
	}
	return nil
}

func FcCoinSetOpenCollect(name string, bof string) error {
	var err error
	fc := entity.FcCoinSet{
		IsCollect:        1,
		CollectThreshold: bof,
	}
	if bof != "" {
		_, err = db.Conn.Cols("is_collect", "collect_threshold").Where("name = ?", name).Update(&fc)
	} else {
		_, err = db.Conn.Cols("is_collect").Where("name = ?", name).Update(&fc)
	}
	if err != nil {
		return err
	}
	return nil
}
