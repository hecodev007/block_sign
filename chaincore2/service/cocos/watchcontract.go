package cocos

import (
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: 合约名称
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {
	InsertContract("COCOS", 8, "1.3.0")
	InsertContract("CFS", 5, "1.3.48")
}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	if WatchContractList[address] != nil {
		return
	}
	contract := new(models.ContractInfo)
	contract.Name = name
	contract.Decimal = decimal
	contract.ContractAddress = address
	WatchContractList[address] = contract
}

// 删除关注合约
func RemoveContract(address string) {
	delete(WatchContractList, address)
}