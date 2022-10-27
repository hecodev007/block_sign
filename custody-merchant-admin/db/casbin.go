package db

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/adminPermission/api"
	"custody-merchant-admin/module/log"
	"errors"
	"fmt"
	"github.com/casbin/casbin/util"
	"strings"
)

func (c *AdminCasbinRule) Create() error {
	e := CasbinDB()
	if success, _ := e.AddPolicy(c.V0, c.V1, c.V2); success == false {
		return errors.New("存在相同的API，添加失败")
	}
	return nil
}

func (c *AdminCasbinRule) CreateBatch(sr []api.Entity) error {
	e := CasbinDB()
	for i := 0; i < len(sr); i++ {
		if success, _ := e.AddPolicy(c.V0, sr[i].Path, sr[i].Method); success == false {
			return errors.New("存在相同的API，添加失败")
		}
	}
	return nil
}

func (c *AdminCasbinRule) Update(values interface{}) error {
	CasbinDB()
	if err := model.DB().Model(c).Where("impl = ? AND v2 = ?", c.V1, c.V2).Updates(values).Error; err != nil {
		return err
	}
	return nil
}

func (c *AdminCasbinRule) List() [][]string {
	e := CasbinDB()
	policy := e.GetFilteredPolicy(0, c.V0)
	return policy
}

func (c *AdminCasbinRule) Remove() error {
	e := CasbinDB()
	if success, _ := e.RemovePolicy(c.V0, c.V1, c.V2); success == false {
		return errors.New("没有API，添加失败")
	}
	return nil
}
func (c *AdminCasbinRule) RemoveBatch(sr []api.Entity) error {
	e := CasbinDB()
	for i := 0; i < len(sr); i++ {
		if success, _ := e.RemovePolicy(c.V0, sr[i].Path, sr[i].Method); success == false {
			return errors.New("没有API，添加失败")
		}
	}
	return nil
}

// ClearCasbin
// @function: ClearCasbin
// @description: 清除匹配的权限
// @param: v int, p ...string
// @return: bool
func ClearCasbin(v int, p ...string) bool {
	e := CasbinDB()
	policy, _ := e.RemoveFilteredPolicy(v, p...)
	return policy

}

// ParamsMatch
// @function: ParamsMatch
// @description: 自定义规则函数
// @param: fullNameKey1 string, key2 string
// @return: bool
func ParamsMatch(fullNameKey1 string, key2 string) bool {
	key1 := strings.Split(fullNameKey1, "?")[0]
	// 剥离路径后再使用casbin的keyMatch2
	return util.KeyMatch2(key1, key2)
}

// ParamsMatchFunc
// @function: ParamsMatchFunc
// @description: 自定义规则函数
// @param: args ...interface{}
// @return: interface{}, error
func ParamsMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return ParamsMatch(name1, name2), nil
}

func DeleteRuleByV0(uid int64) error {

	if err := model.DB().Where("v0 =?", fmt.Sprintf("%d", uid)).Delete(AdminCasbinRule{}).Error; err != nil {
		log.Errorf("DeleteRuleByV0 error: %v", err)
		return err
	}
	return nil
}
