package services

import (
	"fmt"
	"github.com/group-coldwallet/flynn/register-service/model"
	"github.com/group-coldwallet/flynn/register-service/model/po"
	"strings"
)

func DeleteWatchAddress(req *model.RemoveRequest) error {
	name := strings.ToLower(req.Name)
	err := validServiceIsCanUse(name)
	if err != nil {
		return err
	}
	//删除数据库地址
	dbAddresses, err := po.FindAddresses(name, req.UserId, req.Addresses)
	if err != nil {
		return fmt.Errorf("获取数据库地址错误：%v,参数：name=[%s],user_id=[%d]", err, name, req.UserId)
	}
	if len(dbAddresses) > 0 {
		var addr []string
		for _, a := range dbAddresses {
			addr = append(addr, a.Address)
		}
		//删除数据
		err = po.DeleteAddresses(name, req.UserId, addr, "used")
		if err != nil {
			return fmt.Errorf("数据库删除监听[%s]地址错误：%v", name, err)
		}
	}
	//// todo 调用接口，删除内存数据
	//url := fmt.Sprintf("%s/%s/remove", conf.Config.ScanServices[name].Url, name)
	//var reqParams []interface{}
	return nil
}

func DeleteContractAddress(req *model.RemoveContractRequest) error {
	return nil
}
