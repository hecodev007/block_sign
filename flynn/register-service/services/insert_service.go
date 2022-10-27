package services

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/flynn/register-service/conf"
	"github.com/group-coldwallet/flynn/register-service/model"
	"github.com/group-coldwallet/flynn/register-service/model/po"
	"github.com/group-coldwallet/flynn/register-service/util"
	log "github.com/sirupsen/logrus"
	"strings"
)

/*
resp:

*/
func InsertWatchAddress(req *model.InsertAddressReq) error {

	var (
		addrInfos  []*po.Addresses
		reqParams  []model.InsertWatchAddressSrvReq
		updateAddr []string
		err        error
	)
	name := strings.ToLower(req.Name)
	err = validServiceIsCanUse(req.Name)
	if err != nil {
		return err
	}
	////第一步，判断是否支持该币种插入地址功能
	//name := strings.ToLower(req.Name)
	//if !conf.IsSupportThisCoin(name) {
	//	return fmt.Errorf("不支持该币种：%s", req.Name)
	//}
	////第二步： 先判断数据服务是否能调通
	//err = testServiceOk(name)
	//if err != nil {
	//	return fmt.Errorf("测试数据服务访问失败: %v", err)
	//}
	// 第三步，整理数据
	findAddresses, err := po.FindAddresses(name, req.UserId, req.Addresses)
	if err != nil {
		return fmt.Errorf("查询数据库是否有已经插入地址失败：%v", err)
	}
	log.Infof("查询到重复地址个数： %d", len(findAddresses))
	for _, address := range req.Addresses {
		addrInfo := new(po.Addresses)
		addrInfo.CoinType = name
		addrInfo.UserId = req.UserId
		addrInfo.Status = "used"
		addrInfo.Address = address
		var params model.InsertWatchAddressSrvReq
		params.UserId = req.UserId
		params.Url = req.Url
		params.Address = address

		reqParams = append(reqParams, params)
		// 避免重复数据到数据库
		var isRepeat bool
		for _, fa := range findAddresses {
			if strings.ToLower(fa.Address) == strings.ToLower(address) {
				isRepeat = true
				if fa.Status != "used" {
					updateAddr = append(updateAddr, address)
				}
				break
			}

		}
		if !isRepeat {
			addrInfos = append(addrInfos, addrInfo)
		}
	}

	if len(updateAddr) > 0 {
		log.Infof("查询到未分配的地址个数为：%d，需要更新为分配状态", len(updateAddr))
		err = po.UpdateAddressesStatus(name, req.UserId, updateAddr, "used")
		if err != nil {
			return fmt.Errorf("地址更新status失败，Error: %v", err)
		}
	}
	//第四步，插入到数据库中去
	if len(addrInfos) > 0 {
		err = po.InsertWatchAddress(addrInfos)
		if err != nil {
			return fmt.Errorf("地址插入到数据库失败，所有地址需要重新插入,Error: %v", err)
		}
	}
	// 第五步，通过接口插入到内存去
	url := fmt.Sprintf("%s/%s/insert", conf.Config.ScanServices[name].Url, name)
	reqData, _ := json.Marshal(reqParams)
	httpReq := util.HttpPost(url)
	httpReq.Body(reqData)
	respData, err := httpReq.Bytes()
	if err != nil {
		return fmt.Errorf("地址插入内存错误： %v", err)
	}
	var resp model.ResponseData
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return fmt.Errorf("地址插入内存错误： %v", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("地址插入内存接口返回错误：%s", resp.Message)
	}

	return nil
}

/*
插入合约
*/
func InsertContractAddress(req *model.InsertContractReq) error {
	//第一步，判断是否支持该主链币种功能
	name := strings.ToLower(req.CoinType)
	//if !conf.IsSupportThisCoin(name) {
	//	return fmt.Errorf("不支持该主链币种：%s", req.CoinType)
	//}
	////第二步： 先判断数据服务是否能调通
	//err := testServiceOk(name)
	//if err != nil {
	//	return fmt.Errorf("测试数据服务访问失败: %v", err)
	//}
	err := validServiceIsCanUse(name)
	if err != nil {
		return err
	}
	//第三步 判数据中是否已经有该数据，没有的话插入到数据库
	ci, err := po.FindContract(req.Name, name, req.ContractAddress)
	if err != nil {
		return fmt.Errorf("查询合约信息错误：%v", err)
	}
	if len(ci) > 0 {
		log.Infof("数据库已经存在该合约地址：%s", req.ContractAddress)
		if ci[0].Invaild != 0 {
			log.Infof("需要更新[%s]合约信息invaild", req.ContractAddress)
			err = po.UpdateContractInvaild(req.Name, name, req.ContractAddress, 0)
			if err != nil {
				return fmt.Errorf("更新合约信息invalid错误：%v", err)
			}
		}
	} else {
		fmt.Println(req.Decimal)
		// 需要插入到数据库
		c := &po.ContractInfo{
			Name:            req.Name,
			ContractAddress: req.ContractAddress,
			CoinType:        req.CoinType,
			Decimal:         req.Decimal,
			Invaild:         0,
		}
		err = po.InsertContractInfo(c)
		if err != nil {
			return fmt.Errorf("插入合约[%s]到数据库错误：%v", req.ContractAddress, err)
		}
	}
	// 插入到内存去
	url := fmt.Sprintf("%s/%s/insertcontract", conf.Config.ScanServices[name].Url, name)
	var reqParams []interface{}
	reqParams = append(reqParams, req)
	data, _ := json.Marshal(reqParams)
	httpReq := util.HttpPost(url)
	httpReq.Body(data)
	respData, err := httpReq.Bytes()
	if err != nil {
		return fmt.Errorf("合约信息插入内存错误： %v", err)
	}
	var resp model.ResponseData
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return fmt.Errorf("合约信息插入内存错误： %v", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("合约信息插入内存接口返回错误：%s", resp.Message)
	}
	return nil
}
