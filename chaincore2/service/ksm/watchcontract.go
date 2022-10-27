package ksm

import (
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: 合约名称
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {

}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	if WatchContractList[name] != nil {
		return
	}

	contract := new(models.ContractInfo)
	contract.Name = name
	contract.Decimal = decimal
	contract.ContractAddress = address
	WatchContractList[name] = contract
}

// 删除关注合约
func RemoveContract(name string) {
	delete(WatchContractList, name)
}
