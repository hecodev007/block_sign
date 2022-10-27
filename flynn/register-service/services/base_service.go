package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/flynn/register-service/conf"
	"github.com/group-coldwallet/flynn/register-service/model"
	"github.com/group-coldwallet/flynn/register-service/util"
)

func validServiceIsCanUse(name string) error {
	//第一步，判断是否支持该币种插入地址功能

	if !conf.IsSupportThisCoin(name) {
		return fmt.Errorf("不支持该主链币种：%s", name)
	}
	//第二步： 先判断数据服务是否能调通
	err := testServiceOk(name)
	if err != nil {
		return fmt.Errorf("测试数据服务访问失败: %v", err)
	}
	return nil
}

func testServiceOk(name string) error {
	testUrl := fmt.Sprintf("%s/info", conf.Config.ScanServices[name].Url)
	testReq := util.HttpGet(testUrl)
	respData, err := testReq.Bytes()
	if err != nil {
		return err
	}
	var resp model.ResponseData
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}
	return nil
}
